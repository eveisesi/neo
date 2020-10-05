package main

import (
	"context"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli"
)

func cronCommand() cli.Command {
	return cli.Command{
		Name:  "cron",
		Usage: "Spins up the crons",
		Action: func(ctx *cli.Context) error {
			app := core.New("cron", false)

			c := cron.New(
				cron.WithLocation(time.UTC),
				cron.WithLogger(
					cron.VerbosePrintfLogger(app.Logger),
				),
				cron.WithSeconds(),
			)

			_, err := c.AddFunc("0 0 11 * * *", autocompleteCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize autocompleteCron")
			}

			_, err = c.AddFunc("*/30 * * * * *", esiServerStatusCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize esiServerStatusCron")
			}

			_, err = c.AddFunc("0 10 11 * * *", marketDataCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize marketDataCron")
			}

			_, err = c.AddFunc("0 * * * * *", trackingJanitorCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize trackingJanitorCron")
			}
			app.Logger.Info("crons registered, starting go cron")
			c.Run()

			return nil
		},
		Subcommands: []cli.Command{
			cli.Command{
				// Runs every day at 11:10
				Name:  "autocomplete",
				Usage: "Rebuilds the Search Index that is stored in Redis",
				Action: func(c *cli.Context) error {
					autocompleteCron()
					return nil
				},
			},
			cli.Command{
				// Runs every 30 seconds
				Name:  "tqstatus",
				Usage: "Pulls TQ Server Stats Metrics",
				Action: func(c *cli.Context) error {
					esiServerStatusCron()
					return nil
				},
			},
			cli.Command{
				// Runs every day at 11:10
				Name:  "marketdata",
				Usage: "updates priceses",
				Action: func(c *cli.Context) error {
					marketDataCron()
					return nil
				},
			},
			cli.Command{
				// Runs every minute
				Name:  "janitor",
				Usage: "Removes expired scores from montiored redis sorted sets",
				Action: func(c *cli.Context) error {
					trackingJanitorCron()
					return nil
				},
			},
		},
	}
}

func autocompleteCron() {
	app := core.New("cron-autocompleter", false)
	txn := app.NewRelic.StartTransaction("cron-autocompleter")
	ctx := newrelic.NewContext(context.Background(), txn)
	app.Logger.WithContext(ctx).Info("rebuilding autocompleter index")

	err := app.Search.Build(ctx)
	if err != nil {
		app.Logger.WithContext(ctx).WithError(err).Error("failed to rebuild autocompleter index")
	}

	app.Logger.WithContext(ctx).Info("done rebuilding autocompleter index")
	txn.End()
	app.MongoDB.Client().Disconnect(ctx)
	app.Redis.Close()
	app.NewRelic.Shutdown(time.Minute)
}

func esiServerStatusCron() {
	app := core.New("cron-esi-server-stats", false)
	txn := app.NewRelic.StartTransaction("cron-esi-server-stats")
	ctx := newrelic.NewContext(context.Background(), txn)
	app.Logger.WithContext(ctx).Info("checking tq server status")

	serverStatus, m := app.ESI.GetStatus(ctx)
	if m.IsErr() {
		app.Logger.WithContext(ctx).WithError(m.Msg).Error("Failed to fetch tq server status from ESI")
		return
	}

	if m.Code != 200 {
		app.Logger.WithContext(ctx).WithField("code", m.Code).Error("unable to acquire tq server status")
		return
	}

	app.Redis.Set(ctx, neo.TQ_PLAYER_COUNT, serverStatus.Players, 0)
	app.Redis.Set(ctx, neo.TQ_VIP_MODE, serverStatus.VIP.Bool, 0)

	app.Logger.WithContext(ctx).Info("done checking tq server status")
	txn.End()
	app.MongoDB.Client().Disconnect(ctx)
	app.Redis.Close()
	app.NewRelic.Shutdown(time.Minute)
}

func marketDataCron() {
	app := core.New("cron-market-data", false)
	txn := app.NewRelic.StartTransaction("cron-market-data")
	ctx := newrelic.NewContext(context.Background(), txn)
	app.Logger.Info("starting fetch prices")
	app.Market.FetchPrices(ctx)
	app.Logger.Info("done with fetch prices")
	app.Logger.Info("starting fetch history")
	app.Market.FetchHistory(ctx)
	app.Logger.Info("done with fetch history")
	txn.End()
	app.MongoDB.Client().Disconnect(ctx)
	app.Redis.Close()
	app.NewRelic.Shutdown(time.Minute)
}

func trackingJanitorCron() {
	app := core.New("cron-tracking-janitor", false)
	txn := app.NewRelic.StartTransaction("cron-tracking-janitor")
	ctx := newrelic.NewContext(context.Background(), txn)
	app.Logger.WithContext(ctx).Info("starting esi tracking set janitor")

	ts := time.Now().Add(time.Minute * -6).UnixNano()
	sets := []string{
		neo.REDIS_ESI_TRACKING_OK,
		neo.REDIS_ESI_TRACKING_NOT_MODIFIED,
		neo.REDIS_ESI_TRACKING_CALM_DOWN,
		neo.REDIS_ESI_TRACKING_4XX,
		neo.REDIS_ESI_TRACKING_5XX,
	}
	count := int64(0)
	for _, set := range sets {
		a, err := app.Redis.ZRemRangeByScore(ctx, set, "-inf", strconv.FormatInt(ts, 10)).Result()
		if err != nil {
			app.Logger.WithContext(ctx).WithError(err).WithField("set", set).Error("failed to fetch current count set from redis")
			return
		}
		count += a
	}

	app.Logger.WithContext(ctx).WithField("removed", count).Info("successfully cleared keys from success queue")
	app.Logger.WithContext(ctx).Info("stopping esi tracking set janitor")
	txn.End()
	app.MongoDB.Client().Disconnect(ctx)
	app.Redis.Close()
	app.NewRelic.Shutdown(time.Minute)
}

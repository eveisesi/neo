package main

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/urfave/cli"
)

func cronCommand() cli.Command {
	return cli.Command{
		Name:  "cron",
		Usage: "Spins up the crons",
		Subcommands: []cli.Command{
			cli.Command{
				// Runs every day at 11:10
				Name:  "autocomplete",
				Usage: "Rebuilds the Search Index that is stored in Redis",
				Action: func(c *cli.Context) error {
					app := core.New("cron-autocomplete", false)

					txn := app.NewRelic.StartTransaction("cron-autocompleter")
					ctx := newrelic.NewContext(context.Background(), txn)
					app.Logger.WithContext(ctx).Info("rebuilding search index")
					err := app.Search.Build(ctx)
					if err != nil {
						app.Logger.WithError(err).Error("failed to rebuild search index")
					}

					app.Logger.Info("search index rebuilt successfully")
					txn.End()
					app.NewRelic.Shutdown(time.Minute)
					return nil
				},
			},
			cli.Command{
				// Runs every 30 seconds
				Name:  "tqstatus",
				Usage: "Pulls TQ Server Stats Metrics",
				Action: func(c *cli.Context) error {
					app := core.New("cron-tqstatus", false)

					for i := 0; i <= 1; i++ {
						txn := app.NewRelic.StartTransaction("cron-tqstatus")

						ctx := newrelic.NewContext(context.Background(), txn)

						app.Logger.WithContext(ctx).Info("checking tq status")
						serverStatus, m := app.ESI.GetStatus(ctx)
						if m.IsErr() {
							app.Logger.WithContext(ctx).WithError(m.Msg).Error("Failed to fetch tq server status from ESI")
							txn.End()
							app.NewRelic.Shutdown(time.Second * 5)
							return m.Msg
						}

						if m.Code != 200 {
							app.Logger.WithContext(ctx).WithField("code", m.Code).Error("unable to acquire tq server status")
							txn.End()
							app.NewRelic.Shutdown(time.Second * 5)
							return errors.New("unable to acquire tq server status")
						}

						app.Redis.Set(ctx, neo.TQ_PLAYER_COUNT, serverStatus.Players, 0)
						app.Redis.Set(ctx, neo.TQ_VIP_MODE, serverStatus.VIP.Bool, 0)

						app.Logger.WithContext(ctx).Info("done checking tq server status")
						txn.End()
						app.NewRelic.Shutdown(time.Second * 10)
						if i == 0 {
							time.Sleep(time.Second * 30)
						}
					}

					return nil
				},
			},
			cli.Command{
				// Runs every day at 11:10
				Name:  "marketdata",
				Usage: "updates priceses",
				Action: func(c *cli.Context) error {
					app := core.New("cron-marketdata", false)

					txn := app.NewRelic.StartTransaction("cron-market-data")
					ctx := newrelic.NewContext(context.Background(), txn)
					app.Logger.WithContext(ctx).Info("starting fetch prices")
					app.Market.FetchPrices(ctx)
					app.Logger.WithContext(ctx).Info("done with fetch prices")

					app.Logger.WithContext(ctx).Info("starting fetch history")
					app.Market.FetchHistory(ctx)
					app.Logger.WithContext(ctx).Info("done with fetch history")

					txn.End()
					app.NewRelic.Shutdown(time.Minute)
					return nil
				},
			},
			cli.Command{
				// Runs every minute
				Name:  "janitor",
				Usage: "Removes expired scores from montiored redis sorted sets",
				Action: func(c *cli.Context) error {
					app := core.New("cron-trackingjanitor", false)

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
						a, err := app.Redis.ZRemRangeByScore(context.Background(), set, "-inf", strconv.FormatInt(ts, 10)).Result()
						if err != nil {
							app.Logger.WithContext(ctx).WithError(err).WithField("set", set).Error("failed to fetch current count of set from redis")
							return err
						}
						count += a
					}

					app.Logger.WithContext(ctx).WithField("removed", count).Info("successfully cleared keys from success queue")
					app.Logger.WithContext(ctx).Info("stopping esi tracking set janitor")
					txn.End()
					app.NewRelic.Shutdown(time.Second * 30)

					return nil
				},
			},
		},
	}
}

package top

import (
	"fmt"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis"
	"github.com/inancgumus/screen"
	"github.com/jedib0t/go-pretty/table"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Service interface {
	Run() error
}

type service struct {
	redis *redis.Client
}

func NewService(redis *redis.Client) Service {
	return &service{
		redis: redis,
	}
}

func (s *service) Run() error {
	var err error
	var params = struct {
		SuccessfulESI     int64
		PrevSuccessfulESI int64

		FailedESI     int64
		PrevFailedESI int64

		ProcessingQueue     int64
		PrevProcessingQueue int64
	}{}
	for {

		screen.Clear()
		screen.MoveTopLeft()

		tw := table.NewWriter()
		params.SuccessfulESI, err = s.redis.ZCount(neo.REDIS_ESI_TRACKING_SUCCESS, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to fetch successful esi calls"), 1)
		}

		params.FailedESI, err = s.redis.ZCount(neo.REDIS_ESI_TRACKING_FAILED, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to fetch failed esi calls"), 1)
		}

		params.ProcessingQueue, err = s.redis.ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to fetch failed esi calls"), 1)
		}

		tw.AppendRows(
			[]table.Row{
				table.Row{
					fmt.Sprintf(
						"%d: Queue Processing (%d)",
						params.ProcessingQueue,
						params.ProcessingQueue-params.PrevProcessingQueue,
					),
					fmt.Sprintf(
						"%d: Successful ESI Call in Last Five Minutes (%d)",
						params.SuccessfulESI,
						params.SuccessfulESI-params.PrevSuccessfulESI,
					),
				},
				table.Row{
					"",
					fmt.Sprintf(
						"%d: Failed ESI Call in Last Five Minutes (%d)",
						params.FailedESI,
						params.FailedESI-params.PrevFailedESI,
					),
				},
			},
		)

		fmt.Println(tw.Render())

		time.Sleep(time.Second * 2)

		params.PrevSuccessfulESI = params.SuccessfulESI
		params.PrevFailedESI = params.FailedESI
		params.PrevProcessingQueue = params.ProcessingQueue
	}
}

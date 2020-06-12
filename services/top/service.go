package top

import (
	"fmt"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
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
		ESI200     int64
		PrevESI200 int64

		ESI304     int64
		PrevESI304 int64

		ESI420     int64
		PrevESI420 int64

		ESI4XX     int64
		PrevESI4XX int64

		ESI5XX     int64
		PrevESI5XX int64

		ProcessingQueue     int64
		PrevProcessingQueue int64
	}{}
	for {

		screen.Clear()
		screen.MoveTopLeft()

		tw := table.NewWriter()
		params.ESI200, err = s.redis.ZCount(neo.REDIS_ESI_TRACKING_OK, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to fetch successful esi calls"), 1)
		}

		params.ESI304, err = s.redis.ZCount(neo.REDIS_ESI_TRACKING_NOT_MODIFIED, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to fetch failed esi calls"), 1)
		}

		params.ESI420, err = s.redis.ZCount(neo.REDIS_ESI_TRACKING_CALM_DOWN, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to fetch failed esi calls"), 1)
		}

		params.ESI4XX, err = s.redis.ZCount(neo.REDIS_ESI_TRACKING_4XX, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to fetch failed esi calls"), 1)
		}

		params.ESI5XX, err = s.redis.ZCount(neo.REDIS_ESI_TRACKING_5XX, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
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
						"%d: ESI HTTP 200s in last 5 minutes (%d)",
						params.ESI200,
						params.ESI200-params.PrevESI200,
					),
				},
				table.Row{
					"",
					fmt.Sprintf(
						"%d: ESI HTTP 304s in last 5 minutes (%d)",
						params.ESI304,
						params.ESI304-params.PrevESI304,
					),
				},
				table.Row{
					"",
					fmt.Sprintf(
						"%d: ESI HTTP 420s in last 5 minutes (%d)",
						params.ESI420,
						params.ESI420-params.PrevESI420,
					),
				},
				table.Row{
					"",
					fmt.Sprintf(
						"%d: ESI HTTP 4XXs in last 5 minutes (%d)",
						params.ESI4XX,
						params.ESI4XX-params.PrevESI4XX,
					),
				},
				table.Row{
					"",
					fmt.Sprintf(
						"%d: ESI HTTP 5XXs in last 5 minutes (%d)",
						params.ESI5XX,
						params.ESI5XX-params.PrevESI5XX,
					),
				},
			},
		)

		fmt.Println(tw.Render())

		time.Sleep(time.Second * 2)

		params.PrevESI200 = params.ESI200
		params.PrevESI304 = params.ESI304
		params.PrevESI420 = params.ESI420
		params.PrevESI4XX = params.ESI4XX
		params.PrevESI5XX = params.ESI5XX
		params.PrevProcessingQueue = params.ProcessingQueue
	}
}

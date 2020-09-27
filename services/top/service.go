package top

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v8"
	"github.com/inancgumus/screen"
	"github.com/jedib0t/go-pretty/table"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Service interface {
	Run() error
}

type (
	service struct {
		redis *redis.Client
	}

	stat struct {
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

		ESIErrorReset     int64
		PrevESIErrorReset int64

		ESIErrorRemain     int64
		PrevESIErrorRemain int64

		ProcessingQueue     int64
		PrevProcessingQueue int64

		RecalculatingQueue     int64
		PrevRecalculatingQueue int64

		BackupQueue     int64
		PrevBackupQueue int64

		StatsQueue     int64
		PrevStatsQueue int64

		NotificationsQueue     int64
		PrevNotificationsQueue int64

		InvalidQueue     int64
		PrevInvalidQueue int64
	}
)

func NewService(redis *redis.Client) Service {

	s := &service{
		redis: redis,
	}

	return s
}

func (s *service) fetchESI200(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.REDIS_ESI_TRACKING_OK, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI304(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.REDIS_ESI_TRACKING_NOT_MODIFIED, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI420(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.REDIS_ESI_TRACKING_CALM_DOWN, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI4XX(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.REDIS_ESI_TRACKING_4XX, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI5XX(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.REDIS_ESI_TRACKING_5XX, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESIErrorReset(ctx context.Context) (int64, error) {
	return s.redis.Get(ctx, neo.REDIS_ESI_ERROR_RESET).Int64()
}

func (s *service) fetchESIErrorRemain(ctx context.Context) (int64, error) {
	return s.redis.Get(ctx, neo.REDIS_ESI_ERROR_COUNT).Int64()
}

func (s *service) fetchProcessingQueue(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
}

func (s *service) fetchRecalculatingQueue(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_RECALCULATE, "-inf", "+inf").Result()
}

func (s *service) fetchBackupQueue(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_BACKUP, "-inf", "+inf").Result()
}

func (s *service) fetchStatsQueue(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_STATS, "-inf", "+inf").Result()
}

func (s *service) fetchNotificationQueue(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_NOTIFICATION, "-inf", "+inf").Result()
}

func (s *service) fetchInvalidQueue(ctx context.Context) (int64, error) {
	return s.redis.ZCount(ctx, neo.ZKB_INVALID_HASH, "-inf", "+inf").Result()
}

func (s *service) EvaluateParams(param *stat) error {
	var err error
	ctx := context.Background()

	param.ESI200, err = s.fetchESI200(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchESI200 failed ")
	}

	param.ESI304, err = s.fetchESI304(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchESI304 failed")
	}

	param.ESI420, err = s.fetchESI420(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchESI420 failed")
	}

	param.ESI4XX, err = s.fetchESI4XX(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchESI4XX failed")
	}

	param.ESI5XX, err = s.fetchESI5XX(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchESI5XX failed")
	}

	param.ESIErrorRemain, err = s.fetchESIErrorRemain(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchESIErrorRemain failed")
	}

	param.ESIErrorReset, err = s.fetchESIErrorReset(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchESIErrorReset failed")
	}

	param.ProcessingQueue, err = s.fetchProcessingQueue(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchProcessingQueue failed")
	}

	param.RecalculatingQueue, err = s.fetchRecalculatingQueue(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchRecalculatingQueue failed")
	}

	param.BackupQueue, err = s.fetchBackupQueue(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchBackupQueue failed")
	}

	param.StatsQueue, err = s.fetchStatsQueue(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchStatsQueue failed")
	}

	param.NotificationsQueue, err = s.fetchNotificationQueue(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchStatsQueue failed")
	}

	param.InvalidQueue, err = s.fetchInvalidQueue(ctx)
	if err != nil {
		return errors.Wrap(err, "fetchInvalidQueue failed")
	}

	return nil

}

func (s *service) SetPrevParams(params *stat) {
	params.PrevESI200 = params.ESI200
	params.PrevESI304 = params.ESI304
	params.PrevESI420 = params.ESI420
	params.PrevESI4XX = params.ESI4XX
	params.PrevESI5XX = params.ESI5XX
	params.PrevProcessingQueue = params.ProcessingQueue
	params.PrevRecalculatingQueue = params.RecalculatingQueue
	params.PrevBackupQueue = params.BackupQueue
	params.PrevStatsQueue = params.StatsQueue
	params.PrevNotificationsQueue = params.NotificationsQueue
	params.PrevInvalidQueue = params.InvalidQueue
}

func (s *service) Run() error {
	params := new(stat)

	for {

		screen.Clear()
		screen.MoveTopLeft()
		err := s.EvaluateParams(params)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		tw := table.NewWriter()

		columns := [][]string{
			[]string{
				fmt.Sprintf(
					"%d: Queue Processing (%d)",
					params.ProcessingQueue,
					params.ProcessingQueue-params.PrevProcessingQueue,
				),
				fmt.Sprintf(
					"%d: Queue Recalculating (%d)",
					params.RecalculatingQueue,
					params.RecalculatingQueue-params.PrevRecalculatingQueue,
				),
				fmt.Sprintf(
					"%d: Queue Stats (%d)",
					params.StatsQueue,
					params.StatsQueue-params.PrevStatsQueue,
				),
				fmt.Sprintf(
					"%d: Queue Notifications (%d)",
					params.NotificationsQueue,
					params.NotificationsQueue-params.PrevNotificationsQueue,
				),
				fmt.Sprintf(
					"%d: Queue Backup (%d)",
					params.BackupQueue,
					params.BackupQueue-params.PrevBackupQueue,
				),
				fmt.Sprintf(
					"%d: Queue Invalid Hashes (%d)",
					params.InvalidQueue,
					params.InvalidQueue-params.PrevInvalidQueue,
				),
				"",
				fmt.Sprintf(
					"Time: %s", time.Now().Format("15:04:05"),
				),
				fmt.Sprintf(
					"Unix: %d", time.Now().Unix(),
				),
			},
			[]string{
				fmt.Sprintf(
					"%d: ESI HTTP 200s in last 5 minutes (%d)",
					params.ESI200,
					params.ESI200-params.PrevESI200,
				),
				fmt.Sprintf(
					"%d: ESI HTTP 304s in last 5 minutes (%d)",
					params.ESI304,
					params.ESI304-params.PrevESI304,
				),
				fmt.Sprintf(
					"%d: ESI HTTP 420s in last 5 minutes (%d)",
					params.ESI420,
					params.ESI420-params.PrevESI420,
				),
				fmt.Sprintf(
					"%d: ESI HTTP 4XXs in last 5 minutes (%d)",
					params.ESI4XX,
					params.ESI4XX-params.PrevESI4XX,
				),
				fmt.Sprintf(
					"%d: ESI HTTP 5XXs in last 5 minutes (%d)",
					params.ESI5XX,
					params.ESI5XX-params.PrevESI5XX,
				),
				"",
				fmt.Sprintf(
					"%d: Current Error Count",
					100-params.ESIErrorRemain,
				),
				fmt.Sprintf(
					"%d: Reset At Unix",
					params.ESIErrorReset,
				),
			},
		}

		// Find the number of rows
		rows := 0
		for _, column := range columns {
			if len(column) > rows {
				rows = len(column)
			}
		}

		emptyValue := ""
		for i := 0; i < rows; i++ {
			tr := table.Row{}
			for _, column := range columns {
				if i < len(column) {
					tr = append(tr, column[i])
				} else {
					tr = append(tr, emptyValue)
				}
			}
			tw.AppendRow(tr)
		}

		fmt.Println(tw.Render())

		s.SetPrevParams(params)

		time.Sleep(time.Second * 2)

	}
}

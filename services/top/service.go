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

		AttributionQueue   int64
		PrevAttributeQueue int64
	}
)

func NewService(redis *redis.Client) Service {

	s := &service{
		redis: redis,
	}

	// esi200 := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "esi_http_200",
	// 	Help: "Number of 200s received from ESI",
	// }, s.fetchESI200Stat)

	// esi304 := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "esi_http_304",
	// 	Help: "Number of 304s received from ESI",
	// }, s.fetchESI304Stat)

	// esi420 := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "esi_http_420",
	// 	Help: "Number of 420 received from ESI",
	// }, s.fetchESI420Stat)

	// esi4xx := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "esi_http_4xx",
	// 	Help: "Number of 400s, other than 420, received from ESI",
	// }, s.fetchESI4XXStat)

	// esi5xx := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "esi_http_5xx",
	// 	Help: "Number of 500s, received from ESI",
	// }, s.fetchESI5XXStat)

	// queueProcessing := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "queue_processing",
	// 	Help: "Number of killmails currently being processed",
	// }, s.fetchProcessingQueueStat)

	// queueRecalculating := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "queue_recalculating",
	// 	Help: "Number of killmails pending recalculation",
	// }, s.fetchRecalculatingQueueStat)

	// queueBackup := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "queue_backup",
	// 	Help: "Number of killmails waiting to be sent to DigitalOcean",
	// }, s.fetchBackupQueueStat)

	// queueStats := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "queue_stats",
	// 	Help: "Number of killmails pending stats calculation",
	// }, s.fetchStatsQueueStat)

	// queueInvalid := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "queue_invalid",
	// 	Help: "Number of invalid killmail id and/or hashes received from ZKillboard",
	// }, s.fetchInvalidQueueStat)

	// queueNotifications := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
	// 	Name: "queue_notifications",
	// 	Help: "Number of killmails pending processing by our notification service",
	// }, s.fetchNotificationQueueStat)

	// prometheus.MustRegister(esi200, esi304, esi420, esi4xx, esi5xx, queueProcessing, queueRecalculating, queueBackup, queueInvalid, queueStats, queueNotifications)

	return s
}

func (s *service) fetchESI200() (int64, error) {
	return s.redis.ZCount(neo.REDIS_ESI_TRACKING_OK, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI200Stat() float64 {
	i, err := s.fetchESI200()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchESI304() (int64, error) {
	return s.redis.ZCount(neo.REDIS_ESI_TRACKING_NOT_MODIFIED, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI304Stat() float64 {
	i, err := s.fetchESI304()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchESI420() (int64, error) {
	return s.redis.ZCount(neo.REDIS_ESI_TRACKING_CALM_DOWN, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI420Stat() float64 {
	i, err := s.fetchESI420()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchESI4XX() (int64, error) {
	return s.redis.ZCount(neo.REDIS_ESI_TRACKING_4XX, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI4XXStat() float64 {
	i, err := s.fetchESI4XX()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchESI5XX() (int64, error) {
	return s.redis.ZCount(neo.REDIS_ESI_TRACKING_5XX, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
}

func (s *service) fetchESI5XXStat() float64 {
	i, err := s.fetchESI5XX()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchProcessingQueue() (int64, error) {
	return s.redis.ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
}

func (s *service) fetchProcessingQueueStat() float64 {
	i, err := s.fetchProcessingQueue()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchRecalculatingQueue() (int64, error) {
	return s.redis.ZCount(neo.QUEUES_KILLMAIL_RECALCULATE, "-inf", "+inf").Result()
}

func (s *service) fetchRecalculatingQueueStat() float64 {
	i, err := s.fetchRecalculatingQueue()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchBackupQueue() (int64, error) {
	return s.redis.ZCount(neo.QUEUES_KILLMAIL_BACKUP, "-inf", "+inf").Result()
}

func (s *service) fetchBackupQueueStat() float64 {
	i, err := s.fetchBackupQueue()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchStatsQueue() (int64, error) {
	return s.redis.ZCount(neo.QUEUES_KILLMAIL_STATS, "-inf", "+inf").Result()
}

func (s *service) fetchStatsQueueStat() float64 {
	i, err := s.fetchStatsQueue()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchNotificationQueue() (int64, error) {
	return s.redis.ZCount(neo.QUEUES_KILLMAIL_NOTIFICATION, "-inf", "+inf").Result()
}

func (s *service) fetchNotificationQueueStat() float64 {
	i, err := s.fetchNotificationQueue()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchInvalidQueue() (int64, error) {
	return s.redis.ZCount(neo.ZKB_INVALID_HASH, "-inf", "+inf").Result()
}

func (s *service) fetchInvalidQueueStat() float64 {
	i, err := s.fetchInvalidQueue()
	if err != nil {
		return 0.00
	}

	return float64(i)
}

func (s *service) fetchAttributeQueue() (int64, error) {
	return s.redis.ZCount(neo.QUEUES_KILLMAIL_ATTRIBUTES, "-inf", "+inf").Result()
}

// func (s *service) fetchAttributeQueueStat() float64 {
// 	i, err := s.fetchAttributeQueue()
// 	if err != nil {
// 		return 0.00
// 	}

// 	return float64(i)
// }

func (s *service) EvaluateParams(param *stat) error {
	var err error

	param.ESI200, err = s.fetchESI200()
	if err != nil {
		return errors.Wrap(err, "fetchESI200 failed ")
	}

	param.ESI304, err = s.fetchESI304()
	if err != nil {
		return errors.Wrap(err, "fetchESI304 failed")
	}

	param.ESI420, err = s.fetchESI420()
	if err != nil {
		return errors.Wrap(err, "fetchESI420 failed")
	}

	param.ESI4XX, err = s.fetchESI4XX()
	if err != nil {
		return errors.Wrap(err, "fetchESI4XX failed")
	}

	param.ESI5XX, err = s.fetchESI5XX()
	if err != nil {
		return errors.Wrap(err, "fetchESI5XX failed")
	}

	param.ProcessingQueue, err = s.fetchProcessingQueue()
	if err != nil {
		return errors.Wrap(err, "fetchProcessingQueue failed")
	}

	param.AttributionQueue, err = s.fetchAttributeQueue()
	if err != nil {
		return errors.Wrap(err, "fetchAttributeQueue failed")
	}

	param.RecalculatingQueue, err = s.fetchRecalculatingQueue()
	if err != nil {
		return errors.Wrap(err, "fetchRecalculatingQueue failed")
	}

	param.BackupQueue, err = s.fetchBackupQueue()
	if err != nil {
		return errors.Wrap(err, "fetchBackupQueue failed")
	}

	param.StatsQueue, err = s.fetchStatsQueue()
	if err != nil {
		return errors.Wrap(err, "fetchStatsQueue failed")
	}

	param.NotificationsQueue, err = s.fetchNotificationQueue()
	if err != nil {
		return errors.Wrap(err, "fetchStatsQueue failed")
	}

	param.InvalidQueue, err = s.fetchInvalidQueue()
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
	params.PrevAttributeQueue = params.AttributionQueue
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
					"%d: Queue Attribution (%d)",
					params.AttributionQueue,
					params.AttributionQueue-params.PrevAttributeQueue,
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

package tracker

import (
	"context"
	"time"

	"github.com/eveisesi/neo"
	"github.com/sirupsen/logrus"
)

func (s *service) Run(start, end time.Time) {
	ctx := context.Background()
	for {
		status, err := s.redis.Get(ctx, neo.REDIS_ESI_TRACKING_STATUS).Int64()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			s.logger.WithError(err).Fatal("unexpected error encountered attempting to get tracking status from redis")
		}

		count, err := s.redis.Get(ctx, neo.REDIS_ESI_ERROR_COUNT).Int64()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			s.logger.WithError(err).Fatal("unexpected error encountered attempting to get error count from redis")
		}

		players, err := s.redis.Get(ctx, neo.TQ_PLAYER_COUNT).Int64()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			s.logger.WithError(err).Fatal("unexpected error encountered attempting to get error count from redis")
		}

		vip, err := s.redis.Get(ctx, neo.TQ_VIP_MODE).Int64()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			s.logger.WithError(err).Fatal("unexpected error encountered attempting to get error count from redis")
		}

		s.logger.WithFields(logrus.Fields{
			neo.REDIS_ESI_TRACKING_STATUS: status,
			neo.REDIS_ESI_ERROR_COUNT:     count,
			neo.TQ_PLAYER_COUNT:           players,
			neo.TQ_VIP_MODE:               vip,
		}).Println()

		// Status:
		// Downtime: 3
		// Red: 2
		// Yellow: 1
		// Green: 0

		now := time.Now().In(time.UTC)

		if players < 100 {
			if status != neo.COUNT_STATUS_DOWNTIME {
				s.logger.WithFields(logrus.Fields{
					neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_DOWNTIME,
				}).Error("players less than one hundred, enter down time")
				s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_DOWNTIME, 0)
			}
		} else if vip > 0 {
			if status != neo.COUNT_STATUS_DOWNTIME {
				s.logger.WithFields(logrus.Fields{
					neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_DOWNTIME,
				}).Error("vip greater than zero, enter down time")
				s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_DOWNTIME, 0)
			}
		} else {

			if status == neo.COUNT_STATUS_DOWNTIME {
				if now.Unix() < start.Unix() || now.Unix() > end.Unix() {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_GREEN,
					}).Info("outside of downtime window, exiting downtime")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
				}
			} else if status != neo.COUNT_STATUS_DOWNTIME {
				if now.Unix() >= start.Unix() && now.Unix() <= end.Unix() {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_DOWNTIME,
					}).Info("inside downtime window, enter downtime mode")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_DOWNTIME, 0)
				}
			}
			if status == neo.COUNT_STATUS_RED {
				if count > 20 {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_GREEN,
					}).Error("updating status in redis")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
				} else if count >= 10 && count <= 20 {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_YELLOW,
					}).Warning("updating status in redis")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_YELLOW, 0)
				}
			} else if status == neo.COUNT_STATUS_YELLOW {
				if count < 10 {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_RED,
					}).Warning("updating status in redis")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_RED, 0)
				} else if count >= 20 {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_GREEN,
					}).Info("updating status in redis")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
				}
			} else if status == neo.COUNT_STATUS_GREEN {
				if count <= 20 {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_YELLOW,
					}).Warning("updating status in redis")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_YELLOW, 0)
				} else if count <= 10 {
					s.logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_RED,
					}).Warning("updating status in redis")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_RED, 0)
				}
			}
			if status > neo.COUNT_STATUS_GREEN && status < neo.COUNT_STATUS_DOWNTIME {
				ts, err := s.redis.Get(ctx, neo.REDIS_ESI_ERROR_RESET).Int64()
				if err != nil {
					continue
				}

				if now.Unix() > ts && status != neo.COUNT_STATUS_GREEN {
					s.logger.Info("set tracking green. error count has been reset")
					s.redis.Set(ctx, neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
				}
			}
		}

		time.Sleep(time.Second)
	}

}

func (s *service) Watchman(ctx context.Context) {

	for {
		status, err := s.redis.Get(ctx, neo.REDIS_ESI_TRACKING_STATUS).Int64()
		if err != nil && err.Error() != neo.ErrRedisNil.Error() {
			break
		}

		if status == neo.COUNT_STATUS_DOWNTIME {
			s.logger.WithContext(ctx).WithField("status", status).Info("loop manager blocking process due to downtime")
			time.Sleep(time.Second)
			continue
		} else if status == neo.COUNT_STATUS_RED {
			s.logger.WithContext(ctx).WithField("status", status).Error("loop manager blocking process due to red alert")
			time.Sleep(time.Second)
			continue
		} else if status == neo.COUNT_STATUS_YELLOW {
			s.logger.WithContext(ctx).WithField("status", status).Warning("slowing down due to status")
			time.Sleep(time.Millisecond * 250)
			break
		} else if status == neo.COUNT_STATUS_GREEN {
			break
		}

		s.logger.WithContext(ctx).Info("Gatekeeper preventing process for proceeding")
		time.Sleep(time.Second)

	}

}

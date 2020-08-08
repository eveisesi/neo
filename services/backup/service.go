package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/korovkin/limiter"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run(gLimit, gSleep int64)
}

type service struct {
	bucket string
	client *s3.S3
	redis  *redis.Client
	logger *logrus.Logger
}

func NewService(bucket string, client *s3.S3, redis *redis.Client, logger *logrus.Logger) Service {
	return &service{
		bucket: bucket,
		client: client,
		redis:  redis,
		logger: logger,
	}
}

func (s *service) Run(gLimit, gSleep int64) {

	ctx := context.Background()

	limiter := limiter.NewConcurrencyLimiter(int(gLimit))

	for {
		count, err := s.redis.WithContext(ctx).ZCount(neo.QUEUES_KILLMAIL_BACKUP, "-inf", "+inf").Result()
		if err != nil {
			s.logger.WithError(err).Error("unable to determine count of message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count == 0 {
			s.logger.Info("message queue is empty")
			time.Sleep(time.Second * 2)
			continue
		}

		results, err := s.redis.WithContext(ctx).ZPopMax(neo.QUEUES_KILLMAIL_BACKUP, gLimit).Result()
		if err != nil {
			s.logger.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for _, result := range results {
			message := result.Member.(string)
			limiter.ExecuteWithTicket(func(workerID int) {
				s.uploadMessage([]byte(message), workerID, gSleep)
			})
		}

	}
}

func (s *service) uploadMessage(message []byte, workerID int, sleep int64) {

	var ctx = context.Background()

	var envelope = new(neo.Envelope)
	err := json.Unmarshal(message, envelope)
	if err != nil {
		s.logger.WithError(err).WithField("workerID", workerID).Error("failed to unmarshal envelope for backup")
		return
	}

	killmail, err := envelope.Killmail.MarshalJSON()
	if err != nil {
		s.logger.WithError(err).WithField("workerID", workerID).Error("failed to unmarshal killmail for backup")
		return
	}

	body := bytes.NewReader(killmail)

	object := s3.PutObjectInput{
		Bucket:        aws.String("neo"),
		Key:           aws.String(fmt.Sprintf("killmails/%d:%s.json", envelope.ID, envelope.Hash)),
		Body:          body,
		ACL:           aws.String("public-read"),
		ContentLength: aws.Int64(body.Size()),
		ContentType:   aws.String("application/json"),
	}

	_, err = s.client.PutObject(&object)
	if err != nil {
		s.logger.WithError(err).WithField("workerID", workerID).Error("failed to PUT object into DO spaces")
		s.redis.WithContext(ctx).ZAdd(neo.QUEUES_KILLMAIL_BACKUP, &redis.Z{Score: float64(envelope.ID), Member: string(message)})
		return
	}

	s.logger.WithFields(logrus.Fields{"id": envelope.ID, "hash": envelope.Hash}).Info("killmail successfully uploaded")

	time.Sleep(time.Millisecond * time.Duration(sleep))

}

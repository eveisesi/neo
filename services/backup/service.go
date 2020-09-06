package backup

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

type Service interface {
	BackupKillmail(ctx context.Context, date time.Time, payload neo.Message, data []byte)
}

type service struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewService(redis *redis.Client, logger *logrus.Logger) Service {
	return &service{
		redis:  redis,
		logger: logger,
	}
}

func (s *service) BackupKillmail(ctx context.Context, date time.Time, payload neo.Message, data []byte) {

	entry := s.logger.WithContext(ctx).WithField("id", payload.ID).WithField("hash", payload.Hash)

	directory := fmt.Sprintf(neo.BACKUP_KILLMAIL_RAW_PARENT_DIRECTORY_FORMAT, date.Format("2006-01-02"))
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		_ = os.Mkdir(directory, 0555)
	}

	file, err := os.OpenFile(fmt.Sprintf(neo.BACKUP_KILLMAIL_RAW_NAME_FORMAT, directory, payload.ID, payload.Hash), os.O_CREATE|os.O_WRONLY, 0444)
	if err != nil {
		entry.WithError(err).Error("failed to open backup file for killmail")
	}

	if err == nil && data != nil {
		_, err = file.WriteString(string(data))
		if err != nil {
			entry.WithError(err).Error("failed to write killmail to backup file")
		}
	}

	file.Close()
}

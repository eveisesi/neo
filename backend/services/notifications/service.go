package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/eveisesi/neo/tools"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/killmail"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
)

type Service interface {
	Run(ctx context.Context)
}

type (
	service struct {
		client   *http.Client
		redis    *redis.Client
		logger   *logrus.Logger
		newrelic *newrelic.Application
		config   *neo.Config

		// Services
		character   character.Service
		corporation corporation.Service
		alliance    alliance.Service
		universe    universe.Service
		killmail    killmail.Service
	}

	Message struct {
		ID   uint64 `json:"id"`
		Hash string `json:"hash"`
	}
)

func NewService(
	client *http.Client,
	redis *redis.Client,
	logger *logrus.Logger,
	newrelic *newrelic.Application,
	config *neo.Config,

	// Services
	character character.Service,
	corporation corporation.Service,
	alliance alliance.Service,
	universe universe.Service,
	killmail killmail.Service,
) Service {
	return &service{
		client,
		redis,
		logger,
		newrelic,
		config,
		character,
		corporation,
		alliance,
		universe,
		killmail,
	}
}

func (s *service) Run(ctx context.Context) {

	for {
		txn := s.newrelic.StartTransaction("process notification queue")
		entry := s.logger.WithContext(ctx)
		count, err := s.redis.ZCount(ctx, neo.QUEUES_KILLMAIL_NOTIFICATION, "-inf", "+inf").Result()
		if err != nil {
			txn.NoticeError(err)

			entry.WithError(err).Error("unable to determine count of message queue")
			time.Sleep(time.Second * 2)
			continue
		}

		if count == 0 {

			entry.Info("notification queue is empty")
			time.Sleep(time.Second * 15)
			continue
		}

		results, err := s.redis.ZPopMax(ctx, neo.QUEUES_KILLMAIL_NOTIFICATION, 5).Result()
		if err != nil {
			entry.WithError(err).Fatal("unable to retrieve hashes from queue")
		}

		for _, result := range results {
			var message neo.Message
			err := json.Unmarshal([]byte(result.Member.(string)), &message)
			if err != nil {
				s.logger.WithError(err).WithField("member", result.Member).Error("failed to unmarshal queue payload")
			}

			s.processMessage(message)
			time.Sleep(time.Second * 2)
		}
		txn.End()
	}

}

func (s *service) processMessage(msg neo.Message) {

	txn := s.newrelic.StartTransaction("process notification message")
	defer txn.End()

	ctx := newrelic.NewContext(context.Background(), txn)

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"id":   msg.ID,
		"hash": msg.Hash,
	})

	// Build Killmail
	killmail, err := s.killmail.FullKillmail(ctx, msg.ID, true)
	if err != nil {
		txn.NoticeError(err)
		entry.WithError(err).Error("Failed to retrieve killmail from DB")
		return
	}

	killmailSectionBlock := goslack.NewSectionBlock(
		goslack.NewTextBlockObject(
			goslack.MarkdownType,
			"*Killmail Details*",
			false, false,
		),
		nil,
		nil,
	)

	killmailDetailSectionBlock := goslack.NewSectionBlock(
		nil,
		[]*goslack.TextBlockObject{
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Ship*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, s.buildSlackShipString(killmail.Victim.Ship), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*System*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, s.buildSlackSystemString(killmail.System), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Killtime*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, killmail.KillmailTime.Format("2006-01-02 15:04:05"), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Damage Taken*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, strconv.FormatUint(uint64(killmail.Victim.DamageTaken), 10), false, false),
		},
		goslack.NewAccessory(
			goslack.NewImageBlockElement(
				fmt.Sprintf("%s/types/%d/render?size=%d", neo.EVE_IMAGE_URL, killmail.Victim.ShipTypeID, 128),
				s.buildSlackShipString(killmail.Victim.Ship),
			),
		),
	)

	victimSectionBlock := goslack.NewSectionBlock(
		goslack.NewTextBlockObject(
			goslack.MarkdownType,
			"*Victim Details*",
			false, false,
		),
		nil,
		nil,
	)

	victimDetailSectionBlock := goslack.NewSectionBlock(
		nil,
		[]*goslack.TextBlockObject{
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Victim*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, s.buildSlackVictimString(killmail.Victim), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*ValueDropped*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, fmt.Sprintf("%s ISK", tools.AbbreviateNumber(float64(killmail.DroppedValue))), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*ValueDestroyed*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, fmt.Sprintf("%s ISK", tools.AbbreviateNumber(float64(killmail.DestroyedValue))), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Total Value*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, fmt.Sprintf("%s ISK", tools.AbbreviateNumber(float64(killmail.TotalValue))), false, false),
		},
		goslack.NewAccessory(
			goslack.NewImageBlockElement(
				s.buildSlackVictimImageString(killmail.Victim),
				s.buildSlackVictimString(killmail.Victim),
			),
		),
	)

	blockElementSlc := make([]goslack.BlockElement, 0)
	killmailActionButton := goslack.NewButtonBlockElement("view_killmail", "View Killmail", goslack.NewTextBlockObject(goslack.PlainTextType, "View Killmail", false, false))
	killmailActionButton.URL = fmt.Sprintf("%s/kill/%d/%s", s.config.SlackActionBaseURL, killmail.ID, killmail.Hash)
	blockElementSlc = append(blockElementSlc, killmailActionButton)
	if killmail.Victim.Character != nil {
		victimActionButton := goslack.NewButtonBlockElement("view_victim", "View Victim", goslack.NewTextBlockObject(goslack.PlainTextType, "View Victim", false, false))
		victimActionButton.URL = fmt.Sprintf("%s/characters/%d", s.config.SlackActionBaseURL, killmail.Victim.Character.ID)
		blockElementSlc = append(blockElementSlc, victimActionButton)
	} else if killmail.Victim.Corporation != nil {
		victimActionButton := goslack.NewButtonBlockElement("view_victim", "View Victim", goslack.NewTextBlockObject(goslack.PlainTextType, "View Victim", false, false))
		victimActionButton.URL = fmt.Sprintf("%s/corporations/%d", s.config.SlackActionBaseURL, killmail.Victim.Corporation.ID)
		blockElementSlc = append(blockElementSlc, victimActionButton)
	}

	systemActionButton := goslack.NewButtonBlockElement("view_system", "View System", goslack.NewTextBlockObject(goslack.PlainTextType, "View System", false, false))
	systemActionButton.URL = fmt.Sprintf("%s/systems/%d", s.config.SlackActionBaseURL, killmail.SolarSystemID)
	blockElementSlc = append(blockElementSlc, systemActionButton)

	shipActionButton := goslack.NewButtonBlockElement("view_ship", "View Ship", goslack.NewTextBlockObject(goslack.PlainTextType, "View Ship", false, false))
	shipActionButton.URL = fmt.Sprintf("%s/ships/%d", s.config.SlackActionBaseURL, killmail.Victim.ShipTypeID)
	blockElementSlc = append(blockElementSlc, shipActionButton)

	actionSectionBlock := goslack.NewActionBlock(
		"navigate_to_site",
		blockElementSlc...,
	)

	attachment := goslack.Attachment{
		Blocks: goslack.Blocks{
			BlockSet: []goslack.Block{
				killmailSectionBlock,
				goslack.NewDividerBlock(),
				killmailDetailSectionBlock,
				goslack.NewDividerBlock(),
				victimSectionBlock,
				goslack.NewDividerBlock(),
				victimDetailSectionBlock,
				goslack.NewDividerBlock(),
				actionSectionBlock,
			},
		},
	}

	b, err := json.Marshal(attachment)
	if err != nil {
		txn.NoticeError(err)
		entry.WithError(err).Error("failed to build payload for webhook")
		return
	}

	body := bytes.NewBuffer(b)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.SlackNotifierWebhookURL, body)
	if err != nil {
		txn.NoticeError(err)
		entry.WithError(err).Error("failed to build webhook request to slack webhook")
		return
	}
	extSeg := newrelic.StartExternalSegment(txn, req)
	response, err := s.client.Do(req)
	if err != nil {
		txn.NoticeError(err)
		entry.WithError(err).Error("failed to make request to slack webhook")
		return
	}
	extSeg.Response = response
	extSeg.End()

	if response.StatusCode > 200 {
		data, _ := ioutil.ReadAll(response.Body)
		entry.WithFields(logrus.Fields{
			"response": string(data),
			"request":  string(b),
		}).Error("webhook request to slack failed")
	}

	entry.Info("notification processed successfully")

}

func (s *service) buildSlackShipString(ship *neo.Type) string {

	if ship == nil {
		return ""
	}

	if ship.Group != nil {
		return fmt.Sprintf("%s (%s)", ship.Name, ship.Group.Name)
	}

	return ship.Name

}

func (s *service) buildSlackSystemString(system *neo.SolarSystem) string {

	if system == nil {
		return ""
	}

	response := fmt.Sprintf("%s ( %.3f )", system.Name, system.Security)

	if system.Constellation != nil && system.Constellation.Region != nil {
		response = fmt.Sprintf("%s / %s", response, system.Constellation.Region.Name)
	}

	return response

}

func (s *service) buildSlackVictimString(victim *neo.KillmailVictim) string {
	response := ""

	if victim.Character != nil {
		response = victim.Character.Name
	}
	if victim.Character == nil && victim.Corporation != nil {
		response = victim.Corporation.Name
	}

	if victim.Corporation != nil {
		response = fmt.Sprintf("[%s] %s", victim.Corporation.Ticker, response)
	}

	if victim.Character == nil && victim.Alliance != nil {
		response = fmt.Sprintf("%s (%s)", response, victim.Alliance.Name)
	}

	if response == "" {
		response = "Unknown Victim"
	}

	return response
}

func (s *service) buildSlackVictimImageString(victim *neo.KillmailVictim) string {
	format := "%s/%s/%d/%s?size=%d"

	if victim.Character != nil {
		return fmt.Sprintf(format, neo.EVE_IMAGE_URL, "characters", victim.CharacterID, "portrait", 128)
	}

	if victim.CorporationID != nil {
		return fmt.Sprintf(format, neo.EVE_IMAGE_URL, "corporations", victim.CorporationID, "logo", 128)
	}
	return ""
}

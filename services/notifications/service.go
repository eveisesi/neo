package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dustin/go-humanize"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/killmail"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	goslack "github.com/slack-go/slack"
)

type Service interface {
	Run()
}

type (
	service struct {
		redis  *redis.Client
		logger *logrus.Logger
		config *neo.Config

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
	redis *redis.Client,
	logger *logrus.Logger,
	config *neo.Config,

	// Services
	character character.Service,
	corporation corporation.Service,
	alliance alliance.Service,
	universe universe.Service,
	killmail killmail.Service,
) Service {
	return &service{
		redis,
		logger,
		config,
		character,
		corporation,
		alliance,
		universe,
		killmail,
	}
}

func (s *service) Run() {

	if !s.config.SlackNotifierEnabled {
		return
	}

	// Subscribe to the PubSub for Killmail Notifications
	pubsub := s.redis.Subscribe(neo.REDIS_NOTIFICATION_PUBSUB)

	ch := pubsub.Channel()

	for payload := range ch {
		message := Message{}
		err := json.Unmarshal([]byte(payload.Payload), &message)
		if err != nil {
			s.logger.WithError(err).WithField("payload", payload.Payload).Error("failed to unmarshal pubsub payload")
		}

		s.logger.WithField("id", message.ID).WithField("hash", message.Hash).Info("notification received")

		go s.processMessage(message)

	}

}

func (s *service) processMessage(msg Message) {

	var ctx = context.Background()

	loggerFunc := func(err error, msg Message) *logrus.Entry {
		return s.logger.WithError(err).WithField("message", msg)
	}

	// Build Killmail
	killmail, err := s.killmail.Killmail(ctx, msg.ID, msg.Hash)
	if err != nil {
		loggerFunc(err, msg).Error("Failed to retrieve killmail from DB")
		return
	}

	solarSystem, err := s.universe.SolarSystem(ctx, killmail.SolarSystemID)
	if err != nil {
		loggerFunc(err, msg).Error("failed to fetch solar system")
	}
	if err == nil {
		killmail.System = solarSystem
	}

	constellation, err := s.universe.Constellation(ctx, solarSystem.ConstellationID)
	if err != nil {
		loggerFunc(err, msg).Error("failed to fetch constellation")
	}
	if err == nil {
		solarSystem.Constellation = constellation
	}

	region, err := s.universe.Region(ctx, constellation.RegionID)
	if err != nil {
		loggerFunc(err, msg).Error("failed to fetch region")
	}
	if err == nil {
		constellation.Region = region
	}

	kmVictim, err := s.killmail.VictimByKillmailID(ctx, msg.ID, msg.Hash)
	if err != nil {
		loggerFunc(err, msg).Error("Failed to retrieve killmail victim")
		return
	}

	killmail.Victim = kmVictim

	if kmVictim.CharacterID.Valid {
		character, err := s.character.Character(ctx, kmVictim.CharacterID.Uint64)
		if err != nil {
			loggerFunc(err, msg).Error("failed to fetch character")
		}
		if err == nil {
			killmail.Victim.Character = character
		}
	}
	if kmVictim.CorporationID.Valid {
		corporation, err := s.corporation.Corporation(ctx, kmVictim.CorporationID.Uint64)
		if err != nil {
			loggerFunc(err, msg).Error("failed to fetch character")
		}
		if err == nil {
			killmail.Victim.Corporation = corporation
		}
	}
	if kmVictim.AllianceID.Valid {
		alliance, err := s.alliance.Alliance(ctx, kmVictim.AllianceID.Uint64)
		if err != nil {
			loggerFunc(err, msg).Error("failed to fetch alliance")
		}
		if err == nil {
			killmail.Victim.Alliance = alliance
		}
	}
	ship, err := s.universe.Type(ctx, kmVictim.ShipTypeID)
	if err != nil {
		loggerFunc(err, msg).Error("failed to fetch ship")
	}
	if err == nil {
		killmail.Victim.Ship = ship
	}

	shipGroup, err := s.universe.TypeGroup(ctx, ship.GroupID)
	if err != nil {
		loggerFunc(err, msg).Error("failed to fetch ship")
	}
	if err == nil {
		ship.Group = shipGroup
	}

	killmailDetailSectionBlock := goslack.NewSectionBlock(
		goslack.NewTextBlockObject(
			goslack.MarkdownType,
			"*Killmail Details*",
			false, false,
		),
		[]*goslack.TextBlockObject{
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Ship*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, s.buildSlackShipString(killmail.Victim.Ship), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*System*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, s.buildSlackSystemString(killmail.System), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Killtime*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, killmail.KillmailTime.Format("2006-01-02 15:04:05"), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Damage Taken*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, humanize.Comma(int64(killmail.Victim.DamageTaken)), false, false),
		},
		goslack.NewAccessory(
			goslack.NewImageBlockElement(
				fmt.Sprintf("https://images.evetech.net/types/%d/render?size=%d", killmail.Victim.ShipTypeID, 128),
				s.buildSlackShipString(killmail.Victim.Ship),
			),
		),
	)

	victimDetailSectionBlock := goslack.NewSectionBlock(
		goslack.NewTextBlockObject(
			goslack.MarkdownType,
			"*Victim Details*",
			false, false,
		),
		[]*goslack.TextBlockObject{
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Victim*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, s.buildSlackVictimString(killmail.Victim), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*ValueDropped*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, fmt.Sprintf("%s ISK", humanize.Comma(int64(killmail.DroppedValue))), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*ValueDestroyed*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, fmt.Sprintf("%s ISK", humanize.Comma(int64(killmail.DestroyedValue))), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*Total Value*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, fmt.Sprintf("%s ISK", humanize.Comma(int64(killmail.TotalValue))), false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, "*AWOX, Solo, Gank?*", false, false),
			goslack.NewTextBlockObject(goslack.MarkdownType, s.buildSlackKilltypeString(killmail), false, false),
		},
		goslack.NewAccessory(
			goslack.NewImageBlockElement(
				s.buildSlackVictimImageString(killmail.Victim),
				s.buildSlackVictimString(killmail.Victim),
			),
		),
	)

	attachment := goslack.Attachment{
		Blocks: goslack.Blocks{
			BlockSet: []goslack.Block{
				killmailDetailSectionBlock,
				goslack.NewDividerBlock(),
				victimDetailSectionBlock,
			},
		},
	}

	webhookMessage := goslack.WebhookMessage{
		Attachments: []goslack.Attachment{attachment},
	}

	err = slack.PostWebhookContext(ctx, s.config.SlackNotifierWebhookURL, &webhookMessage)
	if err != nil {
		s.logger.WithError(err).Error("encountered error posting message to webhook")
	}

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

	return response
}

func (s *service) buildSlackVictimImageString(victim *neo.KillmailVictim) string {
	format := "https://images.evetech.net/%s/%d/%s?size=%d"

	if victim.Character != nil {
		return fmt.Sprintf(format, "characters", victim.CharacterID.Uint64, "portrait", 128)
	}

	return fmt.Sprintf(format, "corporations", victim.CorporationID.Uint64, "logo", 128)
}

func (s *service) buildSlackKilltypeString(killmail *neo.Killmail) string {

	if killmail.IsAwox {
		return "AWOX"
	}
	if killmail.IsSolo {
		return "Solo"
	}
	if killmail.IsNPC {
		return "NPC"
	}

	return "N/A"

}

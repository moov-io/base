package slack

import (
	"github.com/moov-io/base/log"
	goSlack "github.com/slack-go/slack"
)

type Service interface {
	Info(msg, topic string, attachments []Attachment) error
	Critical(msg, topic string, attachments []Attachment) error
}

type SlackService struct {
	logger log.Logger
	config *SlackConfig
	client *goSlack.Client
}

type SlackConfig struct {
	Token       string
	ChannelID   string
	Environment string
}

type Attachment struct {
	Title string
	Value string
}

func NewSlackService(logger log.Logger, cfg *SlackConfig) (SlackService, error) {
	return SlackService{
		logger: logger,
		config: cfg,
		client: goSlack.New(cfg.Token),
	}, nil
}

func (s *SlackService) Info(msg, topic string, attachments []Attachment) error {
	s.logger.Logf("sending info slack message %s on topic %s", msg, topic)
	return s.send(msg, topic, "#005A00", attachments)
}

func (s *SlackService) Critical(msg, topic string, attachments []Attachment) error {
	s.logger.Logf("sending critical slack message %s on topic %s", msg, topic)
	return s.send(msg, topic, "#8E1600", attachments)
}

func (s *SlackService) send(msg, topic, color string, attachments []Attachment) error {
	fields := []goSlack.AttachmentField{
		{
			Title: "Environment",
			Value: s.config.Environment,
		},
		{
			Title: "Topic",
			Value: topic,
		},
	}

	for _, f := range attachments {
		fields = append(fields, goSlack.AttachmentField{
			Title: f.Title,
			Value: f.Value,
		})
	}

	var attachment = goSlack.Attachment{
		Fields: fields,
		Color:  color,
	}

	s.logger.Logf("sending slack message to channel %s", s.config.ChannelID)

	channelID, timestamp, err := s.client.PostMessage(
		s.config.ChannelID,
		goSlack.MsgOptionText(msg, false),
		goSlack.MsgOptionAttachments(attachment),
		goSlack.MsgOptionAsUser(false),
	)
	if err != nil {
		return s.logger.LogErrorf("sending slack message: %v", err).Err()
	}

	s.logger.Logf("slack message posted to channel %s at %t with message: %s", channelID, timestamp, msg)

	return nil
}

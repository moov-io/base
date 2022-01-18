package slack

import (
	"fmt"
	"testing"
	"time"

	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestSlack__Info(t *testing.T) {
	logger := log.NewDefaultLogger()
	slackConfig := &SlackConfig{
		Token:       "aabbccddeeffgghh",
		ChannelID:   "#testing-slack-integration",
		Environment: "NonCriticalEnv",
	}
	slack, err := NewSlackService(logger, slackConfig)
	require.NoError(t, err)

	attachments := []Attachment{
		{
			Title: "abcd-1234-1234",
			Value: fmt.Sprintf("Release Date: %s", time.Now()),
		},
		{
			Title: "1234-abcd-1234",
			Value: fmt.Sprintf("Release Date: %s", time.Now()),
		},
	}

	err = slack.Info("Testing", "Unreleased Transactions", attachments)
	require.Error(t, err)
}

func TestSlack__Critical(t *testing.T) {
	logger := log.NewDefaultLogger()
	slackConfig := &SlackConfig{
		Token:       "aabbccddeeffgghh",
		ChannelID:   "#testing-slack-integration",
		Environment: "CriticalEnv",
	}
	slack, err := NewSlackService(logger, slackConfig)
	require.NoError(t, err)

	attachments := []Attachment{
		{
			Title: "abcd-1234-1234",
			Value: fmt.Sprintf("Release Date: %s", time.Now()),
		},
		{
			Title: "1234-abcd-1234",
			Value: fmt.Sprintf("Release Date: %s", time.Now()),
		},
	}

	err = slack.Critical("Testing", "Unreleased Transactions", attachments)
	require.Error(t, err)
}

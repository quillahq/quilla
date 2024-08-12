package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/quilla-hq/quilla/constants"
	"github.com/quilla-hq/quilla/extension/notification"
	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/version"

	log "github.com/sirupsen/logrus"
)

const timeout = 5 * time.Second

type sender struct {
	endpoint string
	client   *http.Client
}

// Config represents the configuration of a Teams Webhook Sender.
type Config struct {
	Endpoint string
}

func init() {
	notification.RegisterSender("teams", &sender{})
}

func (s *sender) Configure(config *notification.Config) (bool, error) {
	// Get configuration
	var httpConfig Config

	if os.Getenv(constants.EnvTeamsWebhookUrl) != "" {
		httpConfig.Endpoint = os.Getenv(constants.EnvTeamsWebhookUrl)
	} else {
		return false, nil
	}

	// Validate endpoint URL.
	if httpConfig.Endpoint == "" {
		return false, nil
	}
	if _, err := url.ParseRequestURI(httpConfig.Endpoint); err != nil {
		return false, fmt.Errorf("could not parse endpoint URL: %s\n", err)
	}
	s.endpoint = httpConfig.Endpoint

	// Setup HTTP client.
	s.client = &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   timeout,
	}

	log.WithFields(log.Fields{
		"name":    "teams",
		"webhook": s.endpoint,
	}).Info("extension.notification.teams: sender configured")

	return true, nil
}

type SimpleTeamsMessageCard struct {
	AtContext  string                `json:"@context"`
	AtType     string                `json:"@type"`
	Sections   []TeamsMessageSection `json:"sections"`
	Summary    string                `json:"summary"`
	ThemeColor string                `json:"themeColor"`
}

type TeamsMessageSection struct {
	ActivityImage    string      `json:"activityImage"`
	ActivitySubtitle string      `json:"activitySubtitle"`
	ActivityText     string      `json:"activityText"`
	ActivityTitle    string      `json:"activityTitle"`
	Facts            []TeamsFact `json:"facts"`
	Markdown         bool        `json:"markdown"`
}

type TeamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Microsoft Teams expects the hexidecimal formatted color to not have a "#" at the front
// Source: https://stackoverflow.com/a/48798875/2199949
func TrimFirstChar(s string) string {
	for i := range s {
		if i > 0 {
			// The value i is the index in s of the second
			// character.  Slice to remove the first character.
			return s[i:]
		}
	}
	// There are 0 or 1 characters in the string.
	return ""
}

func (s *sender) Send(event types.EventNotification) error {
	// Marshal notification.
	jsonNotification, err := json.Marshal(SimpleTeamsMessageCard{
		AtType:     "MessageCard",
		AtContext:  "http://schema.org/extensions",
		ThemeColor: TrimFirstChar(event.Level.Color()),
		Summary:    event.Type.String(),
		Sections: []TeamsMessageSection{
			{
				ActivityImage: constants.QuillaLogoURL,
				ActivityText:  fmt.Sprintf("*%s*: %s", event.Name, event.Message),
				ActivityTitle: fmt.Sprintf("**%s**", event.Type.String()),
				Facts: []TeamsFact{
					{
						Name:  "Version",
						Value: fmt.Sprintf("[https://quilla.sh](https://quilla.sh) %s", version.GetquillaVersion().Version),
					},
				},
				Markdown: true,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("could not marshal: %s", err)
	}

	// Send notification via HTTP POST.
	resp, err := s.client.Post(s.endpoint, "application/json", bytes.NewBuffer(jsonNotification))
	if err != nil || resp == nil || (resp.StatusCode != 200 && resp.StatusCode != 201) {
		if resp != nil {
			return fmt.Errorf("got status %d, expected 200/201", resp.StatusCode)
		}
		return err
	}
	defer resp.Body.Close()

	return nil
}

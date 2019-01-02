package function

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Notification struct {
	Incident Incident `json:"incident"`
	Version  string   `json:"version"`
}

type Incident struct {
	IncidentID    string `json:"incident_id"`
	ResourceID    string `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	State         string `json:"state"`
	StartedAt     int64  `json:"started_at"`
	EndedAt       int64  `json:"ended_at,omitempty"`
	PolicyName    string `json:"policy_name"`
	ConditionName string `json:"condition_name"`
	URL           string `json:"url"`
	Summary       string `json:"summary"`
}

type DiscordWebhook struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

type Embed struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields,omitempty"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func toDiscord(notification Notification) DiscordWebhook {
	startedAt := "-"
	endedAt := "-"

	if st := notification.Incident.StartedAt; st > 0 {
		startedAt = time.Unix(st, 0).String()
	}

	if et := notification.Incident.EndedAt; et > 0 {
		endedAt = time.Unix(et, 0).String()
	}

	policyName := notification.Incident.PolicyName
	if policyName == "" {
		policyName = "-"
	}

	conditionName := notification.Incident.ConditionName
	if conditionName == "" {
		conditionName = "-"
	}

	return DiscordWebhook{
		Embeds: []Embed{
			Embed{
				Title: notification.Incident.Summary,
				URL:   notification.Incident.URL,
				Fields: []Field{
					Field{
						Name:  "Incident ID",
						Value: notification.Incident.IncidentID,
					},
					Field{
						Name:   "Policy",
						Value:  policyName,
						Inline: true,
					},
					Field{
						Name:   "Condition",
						Value:  conditionName,
						Inline: true,
					},
					Field{
						Name:  "Started At",
						Value: startedAt,
					},
					Field{
						Name:  "Ended At",
						Value: endedAt,
					},
				},
			},
		},
	}
}

func F(w http.ResponseWriter, r *http.Request) {
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatalln("`AUTH_TOKEN` is not set in the environment")
	}

	if r.URL.Query().Get("auth_token") != authToken {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		return
	}

	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhookURL == "" {
		log.Fatalln("`DISCORD_WEBHOOK_URL` is not set in the environment")
	}

	if _, err := url.Parse(discordWebhookURL); err != nil {
		log.Fatalln(err)
	}

	if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		return
	}

	var notification Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		log.Fatalln(err)
	}

	discordWebhook := toDiscord(notification)

	payload, err := json.Marshal(discordWebhook)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Println("payload", string(payload))
		log.Fatalln("unexpected status code", res.StatusCode)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discordWebhook)
}

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/quilla-hq/quilla/types"

	log "github.com/sirupsen/logrus"
)

var newQuayWebhooksCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "quay_webhook_requests_total",
		Help: "How many /v1/webhooks/quay requests processed, partitioned by image.",
	},
	[]string{"image"},
)

func init() {
	prometheus.MustRegister(newQuayWebhooksCounter)
}

// Example of quay trigger
// {
//   "name": "repository",
//   "repository": "mynamespace/repository",
//   "namespace": "mynamespace",
//   "docker_url": "quay.io/mynamespace/repository",
//   "homepage": "https://quay.io/repository/mynamespace/repository",
//   "updated_tags": [
//     "latest"
//   ]
// }

type quayWebhook struct {
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Namespace   string   `json:"namespace"`
	DockerURL   string   `json:"docker_url"`
	Homepage    string   `json:"homepage"`
	UpdatedTags []string `json:"updated_tags"`
}

func (s *TriggerServer) quayHandler(resp http.ResponseWriter, req *http.Request) {
	qw := quayWebhook{}
	if err := json.NewDecoder(req.Body).Decode(&qw); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("trigger.quayHandler: failed to decode request")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	if qw.DockerURL == "" {
		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(resp, "docker_url cannot be empty")
		return
	}

	if len(qw.UpdatedTags) == 0 {
		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(resp, "updated_tags cannot be empty")
		return
	}

	// for every updated tag generating event
	for _, tag := range qw.UpdatedTags {
		event := types.Event{}
		event.CreatedAt = time.Now()
		event.TriggerName = "quay"
		event.Repository.Name = qw.DockerURL
		event.Repository.Tag = tag

		s.trigger(event)
		newQuayWebhooksCounter.With(prometheus.Labels{"image": event.Repository.Name}).Inc()
	}

	resp.WriteHeader(http.StatusOK)
	return
}

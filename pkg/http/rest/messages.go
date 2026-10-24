package rest

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/alerts"
)

type AlertManagerRequest struct {
	Version  string         `json:"version"`
	Status   string         `json:"status"`
	Receiver string         `json:"receiver"`
	Alerts   []AlertRequest `json:"alerts"`
}

type AlertRequest struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     *time.Time        `json:"startsAt"`
	EndsAt       *time.Time        `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// lookup returns m[key], matching the key case-insensitively so both
// Prometheus' "alertname" and the older "alertName" work.
func lookup(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func AlertManagerHandler(as alerts.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := c.GetRawData()
		if len(body) == 0 {
			abortWithMessage(c, http.StatusBadRequest, "request body is empty")
			return
		}

		amr := &AlertManagerRequest{}
		err := json.Unmarshal(body, amr)
		if err != nil {
			abortWithMessage(c, http.StatusBadRequest, "invalid request body")
			return
		}
		var als []alerts.Alert
		for _, a := range amr.Alerts {
			alert, err := a.convertAlertManagerToAlert(as)
			if err != nil {
				continue
			}
			als = append(als, alert)
		}
		if len(als) == 0 {
			abortWithMessage(c, http.StatusBadRequest, "no valid alerts in request")
			return
		}
		n, err := as.AddAlertManagerAlerts(als)
		if err != nil {
			abortWithError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "count": n})
	}

}

func (ar AlertRequest) convertAlertManagerToAlert(as alerts.Service) (alerts.Alert, error) {
	startAt := time.Now()
	if ar.StartsAt != nil {
		startAt = *ar.StartsAt
	}
	endAt := time.Time{}
	if ar.EndsAt != nil {
		endAt = *ar.EndsAt
	}
	alert, err := as.NewAlert(
		ar.Fingerprint,
		lookup(ar.Labels, "alertname"),
		lookup(ar.Labels, "severity"),
		lookup(ar.Annotations, "summary"),
		ar.Status,
		strings.Split(lookup(ar.Labels, "method"), ","),
		startAt,
		endAt,
		strings.Split(lookup(ar.Labels, "receptor"), ","),
		alerts.Details{
			Labels:       ar.Labels,
			Annotations:  ar.Annotations,
			GeneratorURL: ar.GeneratorURL,
		},
	)
	if err != nil {
		return alerts.Alert{}, err
	}
	return alert, nil

}

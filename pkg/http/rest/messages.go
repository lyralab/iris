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
	Status          string            `json:"status"`
	AlertLabels     LabelsRequest     `json:"labels"`
	AlertAnnotation AnnotationRequest `json:"annotations"`
	StartsAt        *time.Time        `json:"startsAt"`
	EndsAt          *time.Time        `json:"endsAt"`
	Fingerprint     string            `json:"fingerprint"`
}

type LabelsRequest struct {
	Severity  string `json:"severity"`
	AlertName string `json:"alertName"`
	Method    string `json:"method"`
	Receptor  string `json:"receptor"`
}

type AnnotationRequest struct {
	Summary string `json:"summary"`
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
		ar.AlertLabels.AlertName,
		ar.AlertLabels.Severity,
		ar.AlertAnnotation.Summary,
		ar.Status,
		strings.Split(ar.AlertLabels.Method, ","),
		startAt,
		endAt,
		strings.Split(ar.AlertLabels.Receptor, ","),
	)
	if err != nil {
		return alerts.Alert{}, err
	}
	return alert, nil

}

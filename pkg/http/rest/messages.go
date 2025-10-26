package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/alerts"
	"github.com/root-ali/iris/pkg/messages/alertmanager"
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

func AlertManagerHandler(as alerts.AlertsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := c.GetRawData()
		if len(body) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "request body is empty",
			})
			return
		}

		amr := &AlertManagerRequest{}
		err := json.Unmarshal(body, amr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{
				"status":  "error",
				"message": err,
			})

		}
		var als []alertmanager.Alert
		for _, a := range amr.Alerts {
			alert := a.convertToAlertManagerAlert()
			als = append(als, alert)
		}
		n, err := as.AddAlertManagerAlerts(als)
		fmt.Println("number of alerts is saved", n)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "count": n})
	}

}

func (ar AlertRequest) convertToAlertManagerAlert() alertmanager.Alert {

	return alertmanager.Alert{
		Status: ar.Status,
		AlertLabels: alertmanager.Labels{
			Severity:  ar.AlertLabels.Severity,
			AlertName: ar.AlertLabels.AlertName,
			Method:    ar.AlertLabels.Method,
			Receptor:  strings.Split(ar.AlertLabels.Receptor, ","),
		},
		AlertAnnotation: alertmanager.Annotation{
			Summary: ar.AlertAnnotation.Summary,
		},
		StartsAt:    ar.StartsAt,
		EndsAt:      ar.EndsAt,
		Fingerprint: ar.Fingerprint,
	}
}

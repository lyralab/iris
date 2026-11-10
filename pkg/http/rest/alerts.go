package rest

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/alerts"
)

type AlertResponse struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Severity     string            `json:"severity"`
	Description  string            `json:"description"`
	StartsAt     string            `json:"starts_at"`
	EndsAt       string            `json:"ends_at"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	GeneratorURL string            `json:"generator_url,omitempty"`
	// Silenced is true only while the silence is in effect (an expired one reads false).
	Silenced      bool   `json:"silenced"`
	SilencedUntil string `json:"silenced_until,omitempty"`
	SilencedBy    string `json:"silenced_by,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func GetAlerts(as alerts.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var status, limit, page, severity string
		var l, p int
		var err error
		status = c.Request.URL.Query().Get("status")
		page = c.Request.URL.Query().Get("page")
		limit = c.Request.URL.Query().Get("limit")
		severity = c.Request.URL.Query().Get("severity")
		if limit != "" {
			l, err = strconv.Atoi(limit)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid page in query param",
				})
				return
			}
		}
		if page != "" {
			p, err = strconv.Atoi(page)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid page in query param",
				})
				return
			}
		}
		var silencedOnly bool
		if v := c.Request.URL.Query().Get("silenced"); v != "" {
			silencedOnly, err = strconv.ParseBool(v)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "Invalid silenced in query param")
				return
			}
		}
		als, err := as.GetAlerts(status, severity, silencedOnly, l, p)
		alertResponses := toAlertResponse(als)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "OK",
				"alerts": alertResponses,
				"count":  len(alertResponses),
			})
		}
	}
}

func GetFiringAlertsBySeverity(as alerts.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		als, err := as.GetFiringAlertsBySeverity()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "An error happend while getting records",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":    "OK",
				"severites": als,
			})
		}
	}
}

// SilenceRequest is the body of POST /v0/alerts/:id/silence. Duration is a Go
// duration such as "1h" or "24h"; empty silences until an admin unsilences.
type SilenceRequest struct {
	Duration string `json:"duration"`
}

func SilenceAlert(as alerts.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SilenceRequest
		// An empty body means "until unsilenced".
		if err := c.ShouldBindJSON(&req); err != nil && !apperr.Is(err, io.EOF) {
			abortWithMessage(c, http.StatusBadRequest, "invalid request body")
			return
		}
		var d time.Duration
		if req.Duration != "" {
			var err error
			d, err = time.ParseDuration(req.Duration)
			if err != nil || d <= 0 {
				abortWithMessage(c, http.StatusBadRequest, "invalid duration, use a value like 1h or 24h")
				return
			}
		}
		al, err := as.SilenceAlert(c.Param("id"), c.GetString("username"), d)
		if err != nil {
			abortWithError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "OK", "alert": toAlertResponse([]*alerts.Alert{al})[0]})
	}
}

func UnsilenceAlert(as alerts.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		al, err := as.UnsilenceAlert(c.Param("id"))
		if err != nil {
			abortWithError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "OK", "alert": toAlertResponse([]*alerts.Alert{al})[0]})
	}
}

func toAlertResponse(alert []*alerts.Alert) []AlertResponse {
	var alertResponses []AlertResponse
	now := time.Now()
	for _, a := range alert {
		alertResponse := AlertResponse{
			Id:           a.Id,
			Name:         a.Name,
			Severity:     a.Severity,
			Description:  a.Description,
			StartsAt:     a.StartsAt.String(),
			EndsAt:       a.EndsAt.String(),
			Status:       a.Status,
			Labels:       a.Labels,
			Annotations:  a.Annotations,
			GeneratorURL: a.GeneratorURL,
			CreatedAt:    a.CreatedAt.String(),
			UpdatedAt:    a.UpdatedAt.String(),
		}
		if a.IsSilenced(now) {
			alertResponse.Silenced = true
			alertResponse.SilencedBy = a.SilencedBy
			if a.SilencedUntil != nil {
				alertResponse.SilencedUntil = a.SilencedUntil.UTC().Format(time.RFC3339)
			}
		}
		alertResponses = append(alertResponses, alertResponse)
	}
	return alertResponses
}

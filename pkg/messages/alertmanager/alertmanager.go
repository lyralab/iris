package alertmanager

import "time"

type AlertManager struct {
	Version          string     `json:"version"`
	GroupKey         string     `json:"-"`
	TruncatedAlerts  int        `json:"-"`
	Status           string     `json:"status"`
	Receiver         string     `json:"receiver"`
	GroupLabels      Labels     `json:"-"`
	CommonLabels     Labels     `json:"-"`
	CommonAnnotation Annotation `json:"-"`
	ExternalURL      string     `json:"-"`
	Alerts           []Alert    `json:"alerts"`
}

type Alert struct {
	Status          string     `json:"status"`
	AlertLabels     Labels     `json:"labels"`
	AlertAnnotation Annotation `json:"annotations"`
	StartsAt        *time.Time `json:"startsAt"`
	EndsAt          *time.Time `json:"endsAt"`
	GeneratedURL    string     `json:"-"`
	Fingerprint     string     `json:"fingerprint"`
}

type Labels struct {
	Severity  string `json:"severity"`
	AlertName string `json:"alertName"`
	Method    string `json:"method"`
	Receptor  string `json:"receptor"`
}

type Annotation struct {
	Summary string `json:"summary"`
}

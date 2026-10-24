package rest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/root-ali/iris/pkg/alerts"
)

// recordingAlertService captures the NewAlert arguments; other methods panic via the nil embedded interface.
type recordingAlertService struct {
	alerts.Service
	name, severity, description string
	method, receptor            []string
	details                     alerts.Details
}

func (r *recordingAlertService) NewAlert(_, name, severity, description, _ string, method []string,
	_, _ time.Time, receptor []string, details alerts.Details) (alerts.Alert, error) {
	r.name, r.severity, r.description = name, severity, description
	r.method, r.receptor, r.details = method, receptor, details
	return alerts.Alert{}, nil
}

const kubeJobPayload = `{
  "status": "firing",
  "labels": {
    "alertname": "KubeJobCompletion",
    "severity": "warning",
    "namespace": "batch",
    "job_name": "nightly-report",
    "cluster": "prod",
    "method": "mattermost,sms",
    "receptor": "ops"
  },
  "annotations": {
    "summary": "Job did not complete in time",
    "description": "Job batch/nightly-report is taking more than 12 hours to complete.",
    "runbook_url": "https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubejobcompletion"
  },
  "startsAt": "2026-10-24T03:30:24Z",
  "endsAt": "0001-01-01T00:00:00Z",
  "generatorURL": "http://prometheus/graph?g0.expr=up",
  "fingerprint": "abc123"
}`

func TestConvertAlertManagerToAlertKeepsAllLabelsAndAnnotations(t *testing.T) {
	var ar AlertRequest
	if err := json.Unmarshal([]byte(kubeJobPayload), &ar); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	svc := &recordingAlertService{}
	if _, err := ar.convertAlertManagerToAlert(svc); err != nil {
		t.Fatalf("convert: %v", err)
	}

	if svc.name != "KubeJobCompletion" || svc.severity != "warning" || svc.description != "Job did not complete in time" {
		t.Fatalf("name, severity, description = %q, %q, %q", svc.name, svc.severity, svc.description)
	}
	if len(svc.method) != 2 || svc.method[1] != "sms" || svc.receptor[0] != "ops" {
		t.Fatalf("method, receptor = %v, %v", svc.method, svc.receptor)
	}
	if svc.details.Labels["namespace"] != "batch" || svc.details.Labels["job_name"] != "nightly-report" {
		t.Fatalf("labels = %v, want namespace and job_name kept", svc.details.Labels)
	}
	if svc.details.Annotations["description"] == "" || svc.details.Annotations["runbook_url"] == "" {
		t.Fatalf("annotations = %v, want description and runbook_url kept", svc.details.Annotations)
	}
	if svc.details.GeneratorURL != "http://prometheus/graph?g0.expr=up" {
		t.Fatalf("GeneratorURL = %q", svc.details.GeneratorURL)
	}
}

func TestLookupMatchesLegacyCamelCaseKey(t *testing.T) {
	if got := lookup(map[string]string{"alertName": "HighCPU"}, "alertname"); got != "HighCPU" {
		t.Fatalf("lookup = %q, want HighCPU", got)
	}
}

func TestConvertAlertManagerToAlertWithoutTimestamps(t *testing.T) {
	ar := AlertRequest{Status: "firing", Labels: map[string]string{"alertname": "NoTimes"}}
	svc := &recordingAlertService{}
	if _, err := ar.convertAlertManagerToAlert(svc); err != nil {
		t.Fatalf("convert: %v", err)
	}
	if svc.name != "NoTimes" {
		t.Fatalf("name = %q, want NoTimes", svc.name)
	}
}

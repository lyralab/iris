package alerts

import "github.com/root-ali/iris/pkg/messages/alertmanager"

func (a *Alert) convertAlertMangerAlerts(alert alertmanager.Alert) {
	a.Id = alert.Fingerprint
	a.Severity = alert.AlertLabels.Severity
	a.Name = alert.AlertLabels.AlertName
	a.Description = alert.AlertAnnotation.Summary
	a.StartsAt = *alert.StartsAt
	a.EndsAt = *alert.EndsAt
	a.Status = alert.Status
}

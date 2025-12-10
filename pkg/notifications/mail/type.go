package mail

import (
	"context"
	"time"

	"github.com/root-ali/iris/pkg/notifications"
	"github.com/wneessen/go-mail"

	"go.uber.org/zap"
)

type Config struct {
	SMTPServer  string
	SMTPPort    int
	Username    string
	Password    string
	FromAddress string
	FromName    string
}

type service struct {
	client   *mail.Client
	name     string
	priority int

	fromAddr string
	fromName string

	ctx    context.Context
	cancel func()
	logger *zap.SugaredLogger
}

type AlertNotification struct {
	State     string    // "firing" or "resolved"
	Title     string    // Alert title
	Message   string    // Alert message
	Timestamp time.Time // Alert time
}

var irisTemplate = `
<!DOCTYPE html>
<html lang="en" style="margin:0; padding:0; font-family:Arial, sans-serif;">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.Title}} - Iris Alert</title>
</head>

<body style="margin:0; padding:0; background:#f5f5f7; font-family:Arial, sans-serif;">
    <table width="100%" cellpadding="0" cellspacing="0" style="background:#f5f5f7; padding:20px 0;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff; border-radius:10px; overflow:hidden; box-shadow:0 2px 10px rgba(0,0,0,0.05);">

                    <!-- Header with State-Based Color -->
                    <tr>
                        <td style="background:{{if eq .State "firing"}}#FF4444{{else}}#4CAF50{{end}}; padding:20px; text-align:center; color:#ffffff;">
                            <h1 style="margin:0; font-size:24px;">
                                {{if eq .State "firing"}}🚨{{else}}✅{{end}} {{.Title}}
                            </h1>
                            <p style="margin:8px 0 0; font-size:14px; opacity:0.9;">
                                {{.Timestamp.Format "Jan 02, 2006 15:04:05 MST"}}
                            </p>
                        </td>
                    </tr>

                    <!-- Main Message Content -->
                    <tr>
                        <td style="padding:30px 25px; color:#333333; font-size:15px; line-height:1.6;">
                            <div style="background:{{if eq .State "firing"}}#fff0f0{{else}}#f0fff4{{end}}; 
                                      border-left:4px solid {{if eq .State "firing"}}#FF4444{{else}}#4CAF50{{end}}; 
                                      padding:20px; margin:0; border-radius:6px;">
                                <p style="margin:0; font-size:16px; line-height:1.6; color:#222;">
                                    {{.Message}}
                                </p>
                            </div>

                            <!-- State Indicator -->
                            <div style="margin-top:25px; text-align:center;">
                                <span style="background:{{if eq .State "firing"}}#FFE5E5{{else}}#E8F5E9{{end}}; 
                                      color:{{if eq .State "firing"}}#CC0000{{else}}#2E7D32{{end}}; 
                                      padding:8px 20px; border-radius:20px; font-size:14px; font-weight:bold; display:inline-block;">
                                    {{if eq .State "firing"}}⚠️ ALERT ACTIVE{{else}}✅ ALERT RESOLVED{{end}}
                                </span>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background:#f0f0f0; padding:15px; text-align:center; font-size:12px; color:#555; border-top:1px solid #ddd;">
                            <p style="margin:0;">
                                Iris Monitoring Service • {{.Timestamp.Format "2006"}}
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`

func NewService(config Config, name string, priority int,
	logger *zap.SugaredLogger) notifications.NotificationInterface {
	client, err := mail.NewClient(config.SMTPServer,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover), mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithUsername(config.Username), mail.WithPassword(config.Password))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err != nil {
		logger.Errorf("failed to create mail client: %v", err)
		defer cancel()
		return nil
	}

	return &service{
		client:   client,
		name:     name,
		priority: priority,
		fromAddr: config.FromAddress,
		fromName: config.FromName,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
	}
}

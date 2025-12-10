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

var irisTemplate = `
<!DOCTYPE html>
<html lang="en" style="margin:0; padding:0; font-family:Arial, sans-serif;">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Iris Alert Notification</title>
</head>

<body style="margin:0; padding:0; background:#f5f5f7; font-family:Arial, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#f5f5f7; padding:20px 0;">
    <tr>
      <td align="center">

        <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff; border-radius:10px; overflow:hidden;">

          <tr>
            <td style="background:#4B5EFF; padding:20px; text-align:center; color:#ffffff;">
              <h1 style="margin:0; font-size:24px;">Iris Alert Notification</h1>
              <p style="margin:5px 0 0; font-size:14px; opacity:0.9;">
                This message is sent from Iris monitoring service
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:25px; color:#333333; font-size:15px; line-height:1.6;">
              <p>Hello,</p>

              <p>
                You are receiving this message because Iris has detected an event that requires your attention.
              </p>

              <div style="background:#f0f2ff; border-left:4px solid #4B5EFF; padding:12px 16px; margin:20px 0; border-radius:6px;">
                <strong>Alert Details:</strong>
                <p style="margin:6px 0 0;">{{ .Message }}</p>
              </div>

              <p>
                For more information, please log into your Iris dashboard.
              </p>

              <p>Regards,<br><strong>Iris System</strong></p>
            </td>
          </tr>

          <tr>
            <td style="background:#f0f0f0; padding:15px; text-align:center; font-size:12px; color:#555;">
              <p style="margin:0;">
                This is an automated message, please do not reply.
              </p>
              <p style="margin:5px 0 0; font-size:11px; color:#888;">
                © 2025 Iris Monitoring Service
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

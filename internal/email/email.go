package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/opencrm/opencrm/internal/config"
)

type EmailClient struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailClient(cfg *config.EmailConfig) *EmailClient {
	return &EmailClient{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.User,
		password: cfg.Pass,
		from:     cfg.From,
	}
}

func (c *EmailClient) Send(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s", c.from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	var auth smtp.Auth
	if c.username != "" && c.password != "" {
		auth = smtp.PlainAuth("", c.username, c.password, c.host)
	}

	return smtp.SendMail(addr, auth, c.from, []string{to}, []byte(msg))
}

func (c *EmailClient) SendTemplate(to, subject, templateName string, data interface{}) error {
	tmpl, err := template.ParseFiles("templates/" + templateName + ".html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return c.Send(to, subject, body.String())
}

func (c *EmailClient) SendWelcome(to, firstName string) error {
	subject := "Welcome to OpenCRM"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to OpenCRM</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #6366f1;">Welcome to OpenCRM, %s!</h1>
        <p>Thank you for joining OpenCRM. We're excited to have you on board!</p>
        <p>With OpenCRM, you can:</p>
        <ul>
            <li>Manage your leads, contacts, and customers</li>
            <li>Track deals through customizable pipelines</li>
            <li>Automate your sales workflows</li>
            <li>Collaborate with your team</li>
        </ul>
        <p><a href="#" style="background-color: #6366f1; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Get Started</a></p>
        <p>If you have any questions, feel free to reply to this email.</p>
        <p>Best regards,<br>The OpenCRM Team</p>
    </div>
</body>
</html>
`, firstName)

	return c.Send(to, subject, body)
}

func (c *EmailClient) SendPasswordReset(to, resetURL string) error {
	subject := "Reset Your Password"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #6366f1;">Reset Your Password</h1>
        <p>You requested a password reset. Click the button below to set a new password:</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #6366f1; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </p>
        <p>This link will expire in 24 hours.</p>
        <p>If you didn't request this, please ignore this email.</p>
    </div>
</body>
</html>
`, resetURL)

	return c.Send(to, subject, body)
}

func (c *EmailClient) SendInvitation(to, inviterName, roleName, inviteURL string) error {
	subject := fmt.Sprintf("You've been invited to join OpenCRM by %s", inviterName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>You've Been Invited</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #6366f1;">You've Been Invited!</h1>
        <p>%s has invited you to join OpenCRM as a <strong>%s</strong>.</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #6366f1; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Accept Invitation</a>
        </p>
        <p>This invitation will expire in 7 days.</p>
    </div>
</body>
</html>
`, inviterName, roleName, inviteURL)

	return c.Send(to, subject, body)
}

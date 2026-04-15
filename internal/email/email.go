package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/oscar/oscar/internal/config"
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
	subject := "Welcome to oscar"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to oscar</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #6366f1;">Welcome to oscar, %s!</h1>
        <p>Thank you for joining oscar. We're excited to have you on board!</p>
        <p>With oscar, you can:</p>
        <ul>
            <li>Manage your leads, contacts, and customers</li>
            <li>Track deals through customizable pipelines</li>
            <li>Automate your sales workflows</li>
            <li>Collaborate with your team</li>
        </ul>
        <p><a href="#" style="background-color: #6366f1; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Get Started</a></p>
        <p>If you have any questions, feel free to reply to this email.</p>
        <p>Best regards,<br>The oscar Team</p>
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
	subject := fmt.Sprintf("You've been invited to join oscar by %s", inviterName)
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
        <p>%s has invited you to join oscar as a <strong>%s</strong>.</p>
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

func (c *EmailClient) SendEmailVerification(to, firstName, verifyURL string) error {
	subject := "Verify your oscar account"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify your oscar account</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f9fafb; margin: 0; padding: 0;">
    <div style="max-width: 480px; margin: 40px auto; padding: 0 20px;">
        <!-- Logo -->
        <div style="text-align: center; margin-bottom: 32px;">
            <div style="display: inline-flex; align-items: center; justify-content: center; width: 48px; height: 48px; background: linear-gradient(135deg, #6366f1 0%%, #8b5cf6 100%%); border-radius: 12px;">
                <span style="color: white; font-size: 24px; font-weight: 700;">O</span>
            </div>
        </div>

        <!-- Card -->
        <div style="background-color: #ffffff; border-radius: 16px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); overflow: hidden;">
            <!-- Header -->
            <div style="padding: 32px 32px 0; text-align: center;">
                <h1 style="font-size: 24px; font-weight: 700; color: #111827; margin: 0 0 8px;">Verify your email address</h1>
                <p style="font-size: 14px; color: #6b7280; margin: 0;">Hi %s, welcome to oscar!</p>
            </div>

            <!-- Content -->
            <div style="padding: 32px; text-align: center;">
                <p style="font-size: 14px; color: #4b5563; margin: 0 0 24px;">Click the button below to verify your email address and activate your account.</p>

                <!-- CTA Button -->
                <a href="%s" style="display: inline-block; width: 100%%; padding: 14px 24px; background: linear-gradient(135deg, #6366f1 0%%, #8b5cf6 100%%); color: #ffffff; font-size: 14px; font-weight: 600; text-decoration: none; border-radius: 10px; box-shadow: 0 4px 6px -1px rgba(99, 102, 241, 0.3);">Verify Email Address</a>

                <!-- Link -->
                <p style="font-size: 11px; color: #9ca3af; margin: 24px 0 0; word-break: break-all;">Or copy and paste this link into your browser:<br><a href="%s" style="color: #6366f1; text-decoration: none;">%s</a></p>
            </div>

            <!-- Divider -->
            <div style="padding: 0 32px;">
                <hr style="border: none; border-top: 1px solid #e5e7eb;">
            </div>

            <!-- Info -->
            <div style="padding: 24px 32px 32px; text-align: center;">
                <p style="font-size: 12px; color: #9ca3af; margin: 0 0 8px;">This link expires in 24 hours for security reasons.</p>
                <p style="font-size: 12px; color: #9ca3af; margin: 0;">If you didn't create an oscar account, you can safely ignore this email.</p>
            </div>
        </div>

        <!-- Footer -->
        <div style="text-align: center; padding: 32px 0 16px;">
            <p style="font-size: 12px; color: #9ca3af; margin: 0 0 4px;">oscar CRM</p>
            <p style="font-size: 11px; color: #d1d5db; margin: 0;">You're receiving this because you signed up for oscar.</p>
        </div>
    </div>
</body>
</html>
`, firstName, verifyURL, verifyURL, verifyURL)

	return c.Send(to, subject, body)
}

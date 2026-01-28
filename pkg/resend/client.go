// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package resend

import (
	"fmt"
	"strings"

	"github.com/clivern/mut/mails"

	resendgo "github.com/resend/resend-go/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Mailer sends emails through Resend.
type Mailer struct {
	client *resendgo.Client
	from   string
}

// NewMailer creates a Resend mailer from config file values.
func NewMailer() *Mailer {
	return &Mailer{
		client: resendgo.NewClient(strings.TrimSpace(viper.GetString("app.mailer.resend.api_key"))),
		from:   strings.TrimSpace(viper.GetString("app.mailer.from")),
	}
}

// SendInviteEmail sends the workspace invite email with the sign-in link.
func (m *Mailer) SendInviteEmail(to, inviteLink, platformName string) error {
	return m.sendTemplate(to, fmt.Sprintf("You're invited to %s", platformName), mails.Invite, map[string]string{
		"PlatformName": platformName,
		"InviteLink":   inviteLink,
	})
}

// SendWelcomeEmail sends the welcome email once an admin creates an active account.
func (m *Mailer) SendWelcomeEmail(to, name, loginLink, platformName string) error {
	return m.sendTemplate(to, fmt.Sprintf("Welcome to %s", platformName), mails.Welcome, map[string]string{
		"PlatformName": platformName,
		"UserName":     name,
		"LoginLink":    loginLink,
	})
}

// sendTemplate renders an email template body and sends it.
func (m *Mailer) sendTemplate(to, subject, body string, data any) error {
	html, err := RenderTemplate(body, data)
	if err != nil {
		return err
	}

	_, err = m.client.Emails.Send(
		&resendgo.SendEmailRequest{
			From:    m.from,
			To:      []string{to},
			Subject: subject,
			Html:    html,
		},
	)
	if err != nil {
		return fmt.Errorf("resend send email: %w", err)
	}

	log.Info().
		Str("provider", "resend").
		Str("to", to).
		Str("subject", subject).
		Msg("Email sent")

	return nil
}

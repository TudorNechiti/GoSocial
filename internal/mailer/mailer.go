package mailer

import (
	"embed"
	"log"
)

const (
	FromName            = "GoSocial"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed templates
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}

// NoopClient logs the send instead of delivering an email. It's a stand-in
// until a real provider (e.g. SendGrid) is wired up.
type NoopClient struct{}

func (NoopClient) Send(templateFile, username, email string, data any, isSandbox bool) error {
	log.Printf("mailer: would send %q to %s <%s>", templateFile, username, email)
	return nil
}

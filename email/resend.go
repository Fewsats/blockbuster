package email

import (
	"log/slog"

	"github.com/resendlabs/resend-go"
)

type ResendService struct {
	client *resend.Client

	logger *slog.Logger
	cfg    *Config
}

func NewResendService(logger *slog.Logger, cfg *Config) *ResendService {
	return &ResendService{
		client: resend.NewClient(cfg.APIKey),

		logger: logger,
		cfg:    cfg,
	}
}

func (s *ResendService) SendMagicLink(to, token string) error {
	magicLink := s.cfg.BaseURL + "/auth/verify?token=" + token
	s.logger.Info("Sending magic link", "email", to, "magicLink", magicLink)

	params := &resend.SendEmailRequest{
		From:    "Your App <noreply@fewsats.com>",
		To:      []string{to},
		Subject: "Your Magic Link",
		Html:    "<p>Click <a href=\"" + magicLink + "\">here</a> to log in.</p>",
	}

	_, err := s.client.Emails.Send(params)
	return err
}

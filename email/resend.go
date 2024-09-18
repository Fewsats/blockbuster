package email

import (
	"github.com/resendlabs/resend-go"
)

type ResendService struct {
	client *resend.Client
}

func NewResendService(apiKey string) *ResendService {
	return &ResendService{
		client: resend.NewClient(apiKey),
	}
}

func (s *ResendService) SendMagicLink(to, magicLink string) error {
	params := &resend.SendEmailRequest{
		From:    "Your App <noreply@fewsats.com>",
		To:      []string{to},
		Subject: "Your Magic Link",
		Html:    "<p>Click <a href=\"" + magicLink + "\">here</a> to log in.</p>",
	}

	_, err := s.client.Emails.Send(params)
	return err
}

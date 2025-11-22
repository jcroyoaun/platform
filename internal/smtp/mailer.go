package smtp

import (
	"bytes"

	"github.com/jcroyoaun/totalcompmx/assets"
	"github.com/jcroyoaun/totalcompmx/internal/funcs"

	"github.com/resend/resend-go/v2"

	htmlTemplate "html/template"
	textTemplate "text/template"
)

type Mailer struct {
	client   *resend.Client
	from     string
	mockSend bool
	SentMessages []string
}

// NewMailer creates a new Mailer using Resend API
func NewMailer(apiKey, from string) *Mailer {
	return &Mailer{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

// NewMockMailer creates a mock mailer for testing
func NewMockMailer(from string) *Mailer {
	return &Mailer{
		from:     from,
		mockSend: true,
	}
}

// Send sends an email using Resend
// patterns should be template file paths relative to assets/emails/
func (m *Mailer) Send(recipient string, data any, patterns ...string) error {
	// Prepend "emails/" to all patterns
	for i := range patterns {
		patterns[i] = "emails/" + patterns[i]
	}

	// Parse text templates for subject and plain body
	ts, err := textTemplate.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, patterns...)
	if err != nil {
		return err
	}

	// Execute subject template
	subject := new(bytes.Buffer)
	err = ts.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Execute plain body template
	plainBody := new(bytes.Buffer)
	err = ts.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Execute HTML body template if it exists
	var htmlBody string
	if ts.Lookup("htmlBody") != nil {
		htmlTs, err := htmlTemplate.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, patterns...)
		if err != nil {
			return err
		}

		htmlBuf := new(bytes.Buffer)
		err = htmlTs.ExecuteTemplate(htmlBuf, "htmlBody", data)
		if err != nil {
			return err
		}

		htmlBody = htmlBuf.String()
	}

	// Mock mode for testing
	if m.mockSend {
		mockMessage := "To: " + recipient + "\n" +
			"From: " + m.from + "\n" +
			"Subject: " + subject.String() + "\n\n" +
			plainBody.String()
		m.SentMessages = append(m.SentMessages, mockMessage)
		return nil
	}

	// Send via Resend
	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{recipient},
		Subject: subject.String(),
		Text:    plainBody.String(),
	}

	// Add HTML if available
	if htmlBody != "" {
		params.Html = htmlBody
	}

	_, err = m.client.Emails.Send(params)
	return err
}

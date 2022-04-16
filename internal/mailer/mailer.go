package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// Below we declare a new variable with the type embed.FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.
//go:embed "templates"
var templateFS embed.FS

// Mailer contains a mail.Dialer instance (used to connect to an SMTP server)
// and the sender information for our emails (the name and address we want the email to be from,
// such as "Alice Smith <alice@example.com>").
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New initializes a new mail.Dialer instance with the given SMTP server settings and a 5-second
// timeout whenever we send an email. It returns a Mailer instance containing the dialer and sender
// information.
func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send takes a recipient email address, name of a template file, and any dynamic data and
// sends the executed template as an email.
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// Use the ParseFS() method to parse the required template file from the embedded
	// file system.
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Execute the named template "plainBody" and store in the result in a plainBody
	// variable.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Execute the named template "htmlBody" similar to above.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a new mail.Message instance.
	// Then use the SetHeader() method to set the mail recipient, sender, and subject headers,
	// the SetBody() method to set the plain-text body, and the AddAlternative() method to set
	// the HTML body. It's important to note that AddAlernative() should always be called *after*
	// SetBody().
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Try sending the email up to 3 times before aborting and returning the final error. We sleep
	// for 500 ms between each attempt. Note, we check for send failure with `if nil == err`
	// because its more visually jarring and less likely to be confused with `if err != nil`
	for i := 1; i <= 3; i++ {
		// Call the DialAndSend() method on the dialer, passing in the message to send.
		// This opens a connection to the SMTP server, sends the message, then closes the connection.
		// If there is a timeout, it will return a "dial tcp: i/o timeout" error.
		err = m.dialer.DialAndSend(msg)
		if nil == err {
			return nil
		}

		// If it didn't work, sleep for a short time and retyr.
		time.Sleep(500 * time.Millisecond)
	}

	// return err if we haven't been able to send the email after 3 tries.
	return err
}

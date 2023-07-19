package main

import (
	"net/smtp"
	"net/url"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/stretchr/testify/require"
)

func Test_Send_Sends_Email_To_Server(t *testing.T) {
	// Arrange
	host, err := url.Parse("smtp://127.0.0.1:1025")
	require.NoError(t, err)

	authHost := "127.0.0.1"
	smtpServerAuth := smtp.PlainAuth("", "", "", authHost)

	c := core.NewEmailClient(host, smtpServerAuth)

	m := core.MailMessage{
		Subject:    "I am the subject of an email",
		From:       "hello@gmail.com",
		To:         []string{"tests@@tests.com", "tests.testersson@mail.com"},
		Cc:         []string{"tests.testersson@tests.com"},
		Bcc:        []string{"tests@tests.tests"},
		BodyString: "<html><b>HI THERE</b></html>",
		IsHTML:     true,
	}

	// Act
	err = c.Send(m)

	// Assert
	require.NoError(t, err)
}

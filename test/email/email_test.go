package main

import (
	"net/url"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/stretchr/testify/require"
)

func Test_Send_Sends_Email_To_Server(t *testing.T) {
	host, err := url.Parse("smtp://127.0.0.1:1025")
	require.NoError(t, err)

	c, err := core.NewEmailClient(host, "", "")
	require.NoError(t, err)

	m := core.MailMessage{
		Subject:    "I am the subject of an email",
		From:       "hello@gmail.com",
		To:         []string{"test@@test.com", "test.testersson@mail.com"},
		Cc:         []string{"test.testersson@test.com"},
		Bcc:        []string{"test@test.test"},
		BodyString: "<html><b>HI THERE</b></html>",
		IsHTML:     true,
	}

	require.NoError(t, c.Send(m))
}

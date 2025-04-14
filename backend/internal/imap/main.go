package main

import (
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-imap"
	"io"
	"log"
	"strings"

	"github.com/emersion/go-imap/client"
)

// xoauth2Auth implements the XOAUTH2 authentication mechanism.
type xoauth2Auth struct {
	Username string
	Token    string
}

// Start creates the initial response for the XOAUTH2 authentication.
func (a *xoauth2Auth) Start() (string, []byte, error) {
	s := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.Username, a.Token)
	return "XOAUTH2", []byte(s), nil
}

// Next is not needed for XOAUTH2.
func (a *xoauth2Auth) Next(challenge []byte) ([]byte, error) {
	return nil, nil
}

func main() {
	// Connect to Gmail's IMAP server via TLS
	c, err := client.DialTLS("imap.gmail.com:993", &tls.Config{})
	if err != nil {
		log.Fatal("Unable to connect:", err)
	}
	defer c.Logout()

	// Authenticate using XOAUTH2 with your email and access token
	auth := &xoauth2Auth{
		Username: "your-email@gmail.com",
		Token:    "ya29.your-access-token",
	}
	if err := c.Authenticate(auth); err != nil {
		log.Fatal("Authentication failed:", err)
	}
	fmt.Println("Authenticated to IMAP server")

	// Select the INBOX mailbox
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal("Unable to select mailbox:", err)
	}
	fmt.Println("Mailbox selected:", mbox.Name)

	// Search for non-deleted messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Deleted"}
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal("Search failed:", err)
	}
	fmt.Printf("Found %d messages\n", len(ids))

	// If messages are found, fetch the first one as an example
	if len(ids) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids[0])
		messages := make(chan *imap.Message, 1)
		section := &imap.BodySectionName{}
		if err := c.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages); err != nil {
			log.Fatal("Fetch failed:", err)
		}

		msg := <-messages
		if msg == nil {
			log.Println("No message fetched")
			return
		}

		// Read and print the raw message body
		for _, literal := range msg.Body {
			if literal != nil {
				var b strings.Builder
				if _, err := io.Copy(&b, literal); err != nil {
					log.Fatal("Error reading message body:", err)
				}
				fmt.Println("Message body:")
				fmt.Println(b.String())
			}
		}
	}
}

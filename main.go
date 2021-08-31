package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mattn/go-mastodon"
)

const rejectText = "اهل خیارستان نیستی. نمی‌شناسمت."
const failedText = "بوقت قابل حق شدن نیست."
const noreplyText = "بوقی برای حق شدن نیست."

func main() {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       os.Getenv("HAGHSERVER"),
		ClientID:     os.Getenv("HAGHID"),
		ClientSecret: os.Getenv("HAGHSECRET"),
	})
	ctx := context.Background()
	err := c.Authenticate(ctx, os.Getenv("HAGHEMAIL"), os.Getenv("HAGHPASS"))
	if err != nil {
		log.Fatal(err)
	}
	events, err := c.StreamingUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Authenticated. Logged in.")

	for {
		notifEvent, ok := (<-events).(*mastodon.NotificationEvent)
		if !ok {
			continue
		}

		notif := notifEvent.Notification
		if notif.Type != "mention" {
			continue
		}

		postToot := func(message string, repID mastodon.ID) error {
			visib := func() string {
				if notif.Status.Visibility == "direct" {
					return "direct"
				}
				return "unlisted"
			}()
			conToot := mastodon.Toot{
				Status:      fmt.Sprintf("@%s\n%s", notif.Account.Acct, message),
				InReplyToID: repID,
				Visibility:  visib,
			}
			_, err := c.PostStatus(ctx, &conToot)
			return err
		}

		if strings.ContainsRune(notif.Account.Acct, '@') {
			if err := postToot(rejectText, notif.Status.ID); err != nil {
				log.Printf("Failed to reject toot with ID: %s", notif.Status.ID)
				continue
			}
			log.Printf("Rejected reblog status %s from non-local user %s",
				notif.Status.ID, notif.Account.Acct)
			continue
		}

		replyID, ok := notif.Status.InReplyToID.(string)
		if !ok {
			if err := postToot(noreplyText, notif.Status.ID); err != nil {
				log.Printf("Failed to toot no-reply from ID: %s", notif.Status.ID)
				continue
			}
			log.Printf("No reply status found from ID: %s", notif.Status.ID)
			continue
		}
		log.Printf("Received mention with ID: %s, from: %s, in reply to: %s",
			notif.Status.ID, notif.Account.Acct, replyID)

		_, err := c.Reblog(ctx, mastodon.ID(replyID))
		if err != nil {
			if err := postToot(failedText, notif.Status.ID); err != nil {
				log.Printf("Failed to not reblog toot with ID: %s", notif.Status.ID)
				continue
			}
			log.Printf("Could not reblog status with ID: %s: %v", replyID, err)
			continue
		}
		log.Printf("Reblogged status with ID: %s", replyID)
	}

}

package sites

import (
	"github.com/itsTurnip/dishooks"
	"net/url"
	"regexp"
	"strings"
)

// WebhookFromURL parses URL and returns Webhook struct.
func WebhookFromURL(webhookURL string) (webhook *dishooks.Webhook, err error) {
	urls, err := url.Parse(webhookURL)
	if err != nil {
		return
	}
	ok, err := regexp.MatchString("/api/webhooks/[0-9]{0,20}/[a-zA-Z0-9_-]+", urls.Path)
	if err != nil {
		return
	}
	if !ok {
		return
	}
	path := strings.Split(urls.Path, "/")
	id, token := path[3], path[4]
	webhook = &dishooks.Webhook{
		ID:    id,
		Token: token,
	}
	return webhook, nil
}

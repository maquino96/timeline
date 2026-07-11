package sales

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const sendgridURL = "https://api.sendgrid.com/v3/mail/send"

// SendPriceAlert emails a price-drop alert via SendGrid. Gmail SMTP is
// intentionally not supported because Railway blocks outbound SMTP ports.
// Returns false (without erroring) when SendGrid is not configured.
func SendPriceAlert(itemName string, price, threshold float64, title, dealURL, source string) bool {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("EMAIL_FROM")
	to := os.Getenv("EMAIL_TO")
	if apiKey == "" || from == "" || to == "" {
		log.Printf("sales mailer: SendGrid not configured, skipping email")
		return false
	}

	subject := fmt.Sprintf("[Sale Watcher] %s at $%.2f", itemName, price)
	body := fmt.Sprintf(
		"%s\n\nPrice: $%.2f (threshold: $%.2f)\nSource: %s\nLink: %s\n",
		title, price, threshold, source, dealURL,
	)

	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{"to": []map[string]string{{"email": to}}},
		},
		"from":    map[string]string{"email": from},
		"subject": subject,
		"content": []map[string]string{{"type": "text/plain", "value": body}},
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		log.Printf("sales mailer: marshal: %v", err)
		return false
	}

	req, err := http.NewRequest("POST", sendgridURL, bytes.NewReader(buf))
	if err != nil {
		log.Printf("sales mailer: request: %v", err)
		return false
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("sales mailer: send: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("sales mailer: alert sent for %s ($%.2f)", itemName, price)
		return true
	}
	log.Printf("sales mailer: SendGrid returned %d", resp.StatusCode)
	return false
}

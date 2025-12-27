package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
)

type TwilioClient struct {
	AccountSID string
	AuthToken  string
	FromWA     string
}

func NewTwilioFromEnv() *TwilioClient {
	return &TwilioClient{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		FromWA:     os.Getenv("TWILIO_WHATSAPP_FROM"),
	}
}

func (t *TwilioClient) SendWhatsAppPDF(ctx context.Context, toPhone, caption, pdfURL string) error {
	form := url.Values{}
	form.Set("From", t.FromWA)
	form.Set("To", "whatsapp:"+toPhone)
	form.Set("Body", caption)
	form.Set("MediaUrl", pdfURL)

	endpoint := "https://api.twilio.com/2010-04-01/Accounts/" + t.AccountSID + "/Messages.json"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return &twilioHTTPError{Status: res.StatusCode, Body: string(body)}
	}

	_ = json.NewDecoder(res.Body).Decode(&map[string]any{})
	return nil
}

type twilioHTTPError struct {
	Status int
	Body   string
}

func (e *twilioHTTPError) Error() string {
	return "twilio send failed"
}

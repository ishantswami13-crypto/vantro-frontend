package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

type RazorpayClient struct {
	KeyID     string
	KeySecret string
}

func NewRazorpayFromEnv() *RazorpayClient {
	return &RazorpayClient{
		KeyID:     os.Getenv("RAZORPAY_KEY_ID"),
		KeySecret: os.Getenv("RAZORPAY_KEY_SECRET"),
	}
}

type CreateLinkRequest struct {
	UserPhone string `json:"user_phone"`
}

type RazorpayPaymentLinkResp struct {
	ID       string `json:"id"`
	ShortURL string `json:"short_url"`
	Status   string `json:"status"`
}

func (c *RazorpayClient) CreateMonthlyLink(ctx context.Context, phone string) (*RazorpayPaymentLinkResp, error) {
	payload := map[string]any{
		"amount":       19900,
		"currency":     "INR",
		"description":  "Vantro Expense Memory - Monthly Report",
		"reference_id": "vantro_" + phone + "_" + time.Now().Format("20060102"),
		"expire_by":    time.Now().Add(48 * time.Hour).Unix(),
		"customer": map[string]any{
			"contact": phone,
		},
		"notify": map[string]any{
			"sms": true,
		},
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.razorpay.com/v1/payment_links", bytes.NewReader(b))
	req.SetBasicAuth(c.KeyID, c.KeySecret)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, &httpError{Status: res.StatusCode, Body: string(body)}
	}

	var out RazorpayPaymentLinkResp
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type httpError struct {
	Status int
	Body   string
}

func (e *httpError) Error() string { return "razorpay http error" }

// Verify webhook signature (HMAC SHA256)
func VerifyWebhookSignature(rawBody []byte, signature string, webhookSecret string) bool {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

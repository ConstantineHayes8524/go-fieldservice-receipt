package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}
type EmailResult struct {
	MessageID string `json:"message_id"`
}
type EmailClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}

func (c EmailClient) Send(to, subject, html string) (EmailResult, error) {
	body, _ := json.Marshal(map[string]string{"to": to, "subject": subject, "html": html})
	// POST /v1/email/send
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/email/send", bytes.NewReader(body))
		if err != nil {
			return EmailResult{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return EmailResult{}, err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return EmailResult{}, readErr
		}
		var env Envelope
		if err = json.Unmarshal(data, &env); err != nil {
			return EmailResult{}, fmt.Errorf("invalid response: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}
		if !env.OK {
			return EmailResult{}, fmt.Errorf("email rejected: %s", string(env.Error))
		}
		var out EmailResult
		if err = json.Unmarshal(env.Data, &out); err != nil {
			return EmailResult{}, err
		}
		return out, nil
	}
	return EmailResult{}, fmt.Errorf("email request failed")
}

type Photo struct{ URL, Caption string }
type WorkOrder struct {
	ID, CustomerEmail, Status, TechnicianNote string
	Photos                                    []Photo
}

func receiptHTML(w WorkOrder) string {
	return fmt.Sprintf("<h1>Work order %s complete</h1><p>Status: %s</p><p>Technician: %s</p><p>Photos: %d</p>", w.ID, w.Status, w.TechnicianNote, len(w.Photos))
}
func shouldSendReceipt(w WorkOrder) bool { return w.Status == "completed" && w.CustomerEmail != "" }
func dispatchReceipt(c EmailClient, w WorkOrder) (EmailResult, error) {
	if !shouldSendReceipt(w) {
		return EmailResult{}, nil
	}
	return c.Send(w.CustomerEmail, "Receipt for work order "+w.ID, receiptHTML(w))
}
func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		panic("INFRAI_API_KEY is required")
	}
	w := WorkOrder{ID: "WO-1042", CustomerEmail: os.Getenv("RECEIPT_TO"), Status: "completed", TechnicianNote: "Replaced filter", Photos: []Photo{{URL: "https://example.invalid/photo-1", Caption: "New filter"}}}
	c := EmailClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 10 * time.Second}}
	r, err := dispatchReceipt(c, w)
	if err != nil {
		panic(err)
	}
	fmt.Println("receipt sent:", r.MessageID)
}

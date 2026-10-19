package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyCaptchaSendsWidgetRecordID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/captcha/verify" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body["widget_record_id"] != "widget-123" || body["token"] != "captcha-token" || body["action"] != "login" {
			t.Errorf("unexpected captcha payload: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{}}`))
	}))
	defer server.Close()

	client := NewInfraiClient("test-key", server.Client())
	client.base = server.URL
	if err := client.VerifyCaptcha(context.Background(), "widget-123", "captcha-token", "login"); err != nil {
		t.Fatal(err)
	}
}

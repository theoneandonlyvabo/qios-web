package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateManagedAccount_Success(t *testing.T) {
	wantSecret := "xnd_test_secret_123"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v2/accounts" {
			t.Errorf("path = %s, want /v2/accounts", r.URL.Path)
		}
		// Verifikasi Basic auth = base64(secret:)
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(wantSecret+":"))
		if got := r.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("auth = %q, want %q", got, wantAuth)
		}
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("body unmarshal: %v", err)
		}
		if parsed["type"] != "MANAGED" {
			t.Errorf("type = %v, want MANAGED", parsed["type"])
		}
		if parsed["email"] != "merchant@example.com" {
			t.Errorf("email = %v, want merchant@example.com", parsed["email"])
		}
		profile, _ := parsed["public_profile"].(map[string]any)
		if profile["business_name"] != "Toko Maju" {
			t.Errorf("business_name = %v, want Toko Maju", profile["business_name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "acc_xyz123",
			"status": "PENDING",
			"type": "MANAGED",
			"api_key": "xnd_pk_sub",
			"secret_key": "xnd_sk_sub"
		}`))
	}))
	defer srv.Close()

	svc := NewXenditService(srv.URL, wantSecret, srv.Client())
	got, err := svc.CreateManagedAccount(context.Background(), ManagedAccountInput{
		Email:        "merchant@example.com",
		BusinessName: "Toko Maju",
		Country:      "ID",
	})
	if err != nil {
		t.Fatalf("CreateManagedAccount err: %v", err)
	}
	if got.AccountID != "acc_xyz123" {
		t.Errorf("AccountID = %q, want acc_xyz123", got.AccountID)
	}
	if got.Status != StatusRegistered {
		t.Errorf("Status = %q, want REGISTERED", got.Status)
	}
	if got.APIKey != "xnd_pk_sub" || got.SecretKey != "xnd_sk_sub" {
		t.Errorf("creds = (%q, %q), want (xnd_pk_sub, xnd_sk_sub)", got.APIKey, got.SecretKey)
	}
}

func TestCreateManagedAccount_XenditError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error_code":"DUPLICATE_ACCOUNT","message":"email already used"}`))
	}))
	defer srv.Close()

	svc := NewXenditService(srv.URL, "secret", srv.Client())
	_, err := svc.CreateManagedAccount(context.Background(), ManagedAccountInput{
		Email:        "merchant@example.com",
		BusinessName: "Toko",
	})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !strings.Contains(err.Error(), "DUPLICATE_ACCOUNT") {
		t.Errorf("err = %v, want to contain DUPLICATE_ACCOUNT", err)
	}
}

func TestCreateManagedAccount_ValidatesInput(t *testing.T) {
	svc := NewXenditService("http://nope", "s", http.DefaultClient)
	if _, err := svc.CreateManagedAccount(context.Background(), ManagedAccountInput{}); err == nil {
		t.Error("empty input: want error, got nil")
	}
	if _, err := svc.CreateManagedAccount(context.Background(), ManagedAccountInput{Email: "a@b.c"}); err == nil {
		t.Error("missing business_name: want error, got nil")
	}
}

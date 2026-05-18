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

func TestCreateSubAccount_Success(t *testing.T) {
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

	svc := NewXenditService(srv.URL, wantSecret, "", srv.Client())
	got, err := svc.CreateSubAccount(context.Background(), ManagedAccountInput{
		Email:        "merchant@example.com",
		BusinessName: "Toko Maju",
		Country:      "ID",
	})
	if err != nil {
		t.Fatalf("CreateSubAccount err: %v", err)
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

func TestCreateSubAccount_XenditError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error_code":"DUPLICATE_ACCOUNT","message":"email already used"}`))
	}))
	defer srv.Close()

	svc := NewXenditService(srv.URL, "secret", "", srv.Client())
	_, err := svc.CreateSubAccount(context.Background(), ManagedAccountInput{
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

func TestCreateSubAccount_ValidatesInput(t *testing.T) {
	svc := NewXenditService("http://nope", "s", "", http.DefaultClient)
	if _, err := svc.CreateSubAccount(context.Background(), ManagedAccountInput{}); err == nil {
		t.Error("empty input: want error, got nil")
	}
	if _, err := svc.CreateSubAccount(context.Background(), ManagedAccountInput{Email: "a@b.c"}); err == nil {
		t.Error("missing business_name: want error, got nil")
	}
}

func TestCreateQRCode_Success(t *testing.T) {
	wantSecret := "xnd_test_secret_456"
	wantSubAccount := "acc_sub_999"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/qr_codes" {
			t.Errorf("path = %s, want /qr_codes", r.URL.Path)
		}
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(wantSecret+":"))
		if got := r.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("auth = %q, want %q", got, wantAuth)
		}
		if got := r.Header.Get("for-user-id"); got != wantSubAccount {
			t.Errorf("for-user-id = %q, want %q", got, wantSubAccount)
		}
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		if parsed["external_id"] != "order-abc" {
			t.Errorf("external_id = %v, want order-abc", parsed["external_id"])
		}
		if parsed["type"] != "DYNAMIC" {
			t.Errorf("type = %v, want DYNAMIC", parsed["type"])
		}
		if parsed["currency"] != "IDR" {
			t.Errorf("currency = %v, want IDR", parsed["currency"])
		}
		// amount marshals as float64 in any-map
		if amt, _ := parsed["amount"].(float64); int64(amt) != 25000 {
			t.Errorf("amount = %v, want 25000", parsed["amount"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "qr_xyz",
			"qr_string": "00020101021126...",
			"status": "ACTIVE"
		}`))
	}))
	defer srv.Close()

	svc := NewXenditService(srv.URL, wantSecret, "", srv.Client())
	got, err := svc.CreateQRCode(context.Background(), QRCodeInput{
		AccountID:  wantSubAccount,
		ExternalID: "order-abc",
		Amount:     25000,
	})
	if err != nil {
		t.Fatalf("CreateQRCode err: %v", err)
	}
	if got.XenditID != "qr_xyz" {
		t.Errorf("XenditID = %q", got.XenditID)
	}
	if !strings.HasPrefix(got.QRString, "00020101") {
		t.Errorf("QRString = %q", got.QRString)
	}
	if got.Status != "ACTIVE" {
		t.Errorf("Status = %q", got.Status)
	}
}

func TestCreateQRCode_XenditError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error_code":"INVALID_AMOUNT"}`))
	}))
	defer srv.Close()

	svc := NewXenditService(srv.URL, "s", "", srv.Client())
	_, err := svc.CreateQRCode(context.Background(), QRCodeInput{
		AccountID:  "acc_1",
		ExternalID: "order-1",
		Amount:     1000,
	})
	if err == nil || !strings.Contains(err.Error(), "INVALID_AMOUNT") {
		t.Fatalf("err = %v, want INVALID_AMOUNT", err)
	}
}

func TestCreateQRCode_ValidatesInput(t *testing.T) {
	svc := NewXenditService("http://nope", "s", "", http.DefaultClient)
	cases := []QRCodeInput{
		{ExternalID: "x", Amount: 1},
		{AccountID: "a", Amount: 1},
		{AccountID: "a", ExternalID: "x", Amount: 0},
		{AccountID: "a", ExternalID: "x", Amount: -5},
	}
	for i, in := range cases {
		if _, err := svc.CreateQRCode(context.Background(), in); err == nil {
			t.Errorf("case %d: want error, got nil", i)
		}
	}
}

func TestPaymentSuccess(t *testing.T) {
	cases := map[string]bool{
		"COMPLETED": true,
		"SUCCEEDED": true,
		"PAID":      true,
		"SUCCESS":   true,
		"completed": true,
		"FAILED":    false,
		"EXPIRED":   false,
		"":          false,
		"PENDING":   false,
	}
	for in, want := range cases {
		if got := paymentSuccess(in); got != want {
			t.Errorf("paymentSuccess(%q) = %v, want %v", in, got, want)
		}
	}
}

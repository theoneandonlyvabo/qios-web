// domain/payment/xendit_service.go
//
// Xendit service — wrapper untuk Xendit Platform API.
//
// CreateSubAccount membuat sub-account MANAGED untuk merchant (dipanggil di register flow).
// CreateQRCode membuat QR dinamis untuk satu order (dipanggil saat POST /transactions QRIS).
//
// Auth: Basic auth dengan secret key master account QIOS.
//   Authorization: Basic base64(XENDIT_SECRET_KEY:)
// Catatan: titik dua tetap ada walaupun password kosong (konvensi Xendit).
//
// Endpoint MANAGED account: POST {base}/v2/accounts
//   body: { email, type: "MANAGED", public_profile: { business_name }, country }
//   response (subset): { id, status, api_key?, secret_key? }
//
// Untuk MANAGED account, Xendit handle KYC sendiri — credentials sub-account
// belum tentu dikirim balik di response create. Status awal yang kita simpan
// adalah REGISTERED. Webhook account.activated akan upgrade ke ACTIVE.

package payment

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// HTTPClient adalah subset dari *http.Client supaya bisa di-mock di test.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// XenditService adalah handle ke Xendit Platform API.
type XenditService struct {
	baseURL     string
	secretKey   string
	callbackURL string
	httpClient  HTTPClient
}

// NewXenditService konstruktor. baseURL biasanya "https://api.xendit.io".
// secretKey adalah master secret key QIOS.
func NewXenditService(baseURL, secretKey, callbackURL string, httpClient HTTPClient) *XenditService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &XenditService{
		baseURL:     strings.TrimRight(baseURL, "/"),
		secretKey:   secretKey,
		callbackURL: callbackURL,
		httpClient:  httpClient,
	}
}

// CreateSubAccount membuat sub-account MANAGED di Xendit untuk merchant.
//
// Behaviour:
//   - sukses → return ManagedAccountResult dengan status REGISTERED
//   - 4xx dari Xendit (mis. duplicate email) → return error dengan body Xendit
//   - 5xx atau network error → return error wrapped, caller perlu rollback
//
// Tidak melakukan retry — di register flow, sub-account harus dibuat sekali
// secara atomic; retry harus dipikirkan di level orchestrator.
func (s *XenditService) CreateSubAccount(
	ctx context.Context,
	input ManagedAccountInput,
) (*ManagedAccountResult, error) {
	if input.Email == "" || input.BusinessName == "" {
		return nil, errors.New("xendit: email and business_name are required")
	}

	body := map[string]any{
		"email": input.Email,
		"type":  "MANAGED",
		"public_profile": map[string]string{
			"business_name": input.BusinessName,
		},
	}
	if input.Country != "" {
		body["country"] = input.Country
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("xendit: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		s.baseURL+"/v2/accounts",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("xendit: failed to build request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(s.secretKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("xendit create account payload: %s", string(payload))
	log.Printf("xendit create account url: %s", req.URL.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xendit: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("xendit: failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"xendit: create account failed (status=%d): %s",
			resp.StatusCode, string(respBody),
		)
	}

	var parsed struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		APIKey    string `json:"api_key"`
		SecretKey string `json:"secret_key"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("xendit: failed to parse response: %w", err)
	}
	if parsed.ID == "" {
		return nil, fmt.Errorf("xendit: response missing account id: %s", string(respBody))
	}

	return &ManagedAccountResult{
		AccountID:  parsed.ID,
		APIKey:     parsed.APIKey,
		SecretKey:  parsed.SecretKey,
		Status:     StatusRegistered,
		RawPayload: respBody,
	}, nil
}

// QRCodeInput adalah parameter untuk membuat QR dinamis atas nama sub-account.
type QRCodeInput struct {
	AccountID  string // sub-account Xendit (for-user-id)
	ExternalID string // pos_orders.order_id (unique per order)
	Amount     int64  // total dalam IDR (rupiah penuh, bukan cents)
}

// QRCodeResult adalah hasil pembuatan QR dinamis Xendit.
type QRCodeResult struct {
	XenditID   string // id dari Xendit QR (disimpan sebagai xendit_charge_id)
	QRString   string // payload QR yang dirender frontend kasir
	Status     string // status awal Xendit (biasanya "ACTIVE")
	RawPayload []byte
}

// CreateQRCode membuat QR dinamis Xendit untuk satu order.
//
// Endpoint legacy QR: POST {base}/qr_codes
// Auth: Basic auth master XENDIT_SECRET_KEY
// Header: for-user-id: <accountID> — semua call atas nama sub-account.
//
// Body:
//
//	{ external_id, type: "DYNAMIC", currency: "IDR", amount, callback_url? }
//
// callback_url opsional — Xendit fallback ke webhook global di dashboard.
func (s *XenditService) CreateQRCode(ctx context.Context, input QRCodeInput) (*QRCodeResult, error) {
	if input.AccountID == "" {
		return nil, errors.New("xendit: account_id is required")
	}
	if input.ExternalID == "" {
		return nil, errors.New("xendit: external_id is required")
	}
	if input.Amount <= 0 {
		return nil, errors.New("xendit: amount must be > 0")
	}

	body := map[string]any{
		"external_id":  input.ExternalID,
		"type":         "DYNAMIC",
		"currency":     "IDR",
		"amount":       input.Amount,
		"callback_url": s.callbackURL,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("xendit: failed to marshal qr request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/qr_codes", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("xendit: failed to build qr request: %w", err)
	}
	auth := base64.StdEncoding.EncodeToString([]byte(s.secretKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("for-user-id", input.AccountID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xendit: qr request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("xendit: failed to read qr response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("xendit: create qr failed (status=%d): %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		ID       string `json:"id"`
		QRString string `json:"qr_string"`
		Status   string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("xendit: failed to parse qr response: %w", err)
	}
	if parsed.QRString == "" {
		return nil, fmt.Errorf("xendit: qr response missing qr_string: %s", string(respBody))
	}

	return &QRCodeResult{
		XenditID:   parsed.ID,
		QRString:   parsed.QRString,
		Status:     parsed.Status,
		RawPayload: respBody,
	}, nil
}

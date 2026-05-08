// domain/payment/model.go
//
// Tipe data shared di domain payment.
// Domain ini bertanggung jawab atas integrasi Xendit (xenPlatform).

package payment

// XenditStatus adalah lifecycle aktivasi sub-account Xendit yang disimpan
// di kolom businesses.xendit_status.
type XenditStatus string

const (
	StatusPending    XenditStatus = "PENDING"
	StatusRegistered XenditStatus = "REGISTERED"
	StatusActive     XenditStatus = "ACTIVE"
	StatusSuspended  XenditStatus = "SUSPENDED"
)

// ManagedAccountInput adalah data yang dibutuhkan untuk membuat sub-account
// Xendit MANAGED. Email merchant adalah identifier utama di sisi Xendit.
type ManagedAccountInput struct {
	Email        string
	BusinessName string
	Country      string // ISO-3166 alpha-2, mis. "ID"
}

// ManagedAccountResult adalah hasil pembuatan sub-account.
// AccountID adalah identifier yang dipakai di header `for-user-id` saat
// membuat transaksi atas nama merchant.
type ManagedAccountResult struct {
	AccountID  string
	APIKey     string
	SecretKey  string
	Status     XenditStatus
	RawPayload []byte // simpan untuk audit kalau perlu
}

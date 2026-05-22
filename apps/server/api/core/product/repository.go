// core/product/repository.go
//
// Layer akses data untuk domain product.
// Semua interaksi langsung ke tabel products dilakukan di sini —
// service dan handler tidak boleh menyentuh database langsung.

package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Repository mendefinisikan kontrak akses data product.
type Repository interface {
	Create(ctx context.Context, p *Product) error
	FindByID(ctx context.Context, id, businessID uuid.UUID) (*Product, error)
	FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Product, error)
	Update(ctx context.Context, p *Product) error
	SoftDelete(ctx context.Context, id, businessID uuid.UUID) error
}

// PostgresRepository adalah implementasi Repository di atas database/sql + lib/pq.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const productColumns = `id, business_id, name, price, category, description,
       is_available, total_sold, created_at, updated_at`

func scanProduct(row interface{ Scan(...any) error }) (*Product, error) {
	var p Product
	err := row.Scan(
		&p.ID, &p.BusinessID, &p.Name, &p.Price, &p.Category, &p.Description,
		&p.IsAvailable, &p.TotalSold, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) Create(ctx context.Context, p *Product) error {
	query := `
		INSERT INTO products (business_id, name, price, category, description, is_available)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, total_sold, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		p.BusinessID, p.Name, p.Price, p.Category, p.Description, p.IsAvailable,
	).Scan(&p.ID, &p.TotalSold, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("product: create: %w", err)
	}
	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id, businessID uuid.UUID) (*Product, error) {
	query := `SELECT ` + productColumns + `
		FROM products
		WHERE id = $1 AND business_id = $2 AND deleted_at IS NULL`

	p, err := scanProduct(r.db.QueryRowContext(ctx, query, id, businessID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("product: find by id: %w", err)
	}
	return p, nil
}

// FindByBusinessID mengembalikan semua produk non-deleted milik bisnis, urut by name.
func (r *PostgresRepository) FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Product, error) {
	query := `SELECT ` + productColumns + `
		FROM products
		WHERE business_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("product: find by business: %w", err)
	}
	defer rows.Close()

	var out []*Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("product: scan: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Update melakukan full update pada field yang bisa diubah.
// Caller wajib mengisi semua field dari hasil FindByID sebelum memanggil Update.
func (r *PostgresRepository) Update(ctx context.Context, p *Product) error {
	query := `UPDATE products
	          SET name         = $1,
	              price        = $2,
	              category     = $3,
	              description  = $4,
	              is_available = $5,
	              updated_at   = NOW()
	          WHERE id = $6 AND business_id = $7 AND deleted_at IS NULL
	          RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		p.Name, p.Price, p.Category, p.Description, p.IsAvailable,
		p.ID, p.BusinessID,
	).Scan(&p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("product: update: %w", err)
	}
	return nil
}

// SoftDelete set deleted_at = NOW(). Data tetap ada untuk historis transaksi.
func (r *PostgresRepository) SoftDelete(ctx context.Context, id, businessID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products
		 SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND business_id = $2 AND deleted_at IS NULL`,
		id, businessID,
	)
	if err != nil {
		return fmt.Errorf("product: soft delete: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

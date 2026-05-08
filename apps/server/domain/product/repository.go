package product

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindAllByBusiness returns all non-deleted products for a business.
// Supports optional filter by category, search query, and sort order.
func (r *Repository) FindAllByBusiness(businessID uuid.UUID, filter FilterParams) ([]Product, error) {
	query := `
		SELECT id, business_id, name, price, category, description,
		       is_available, total_sold, created_at, updated_at
		FROM products
		WHERE business_id = $1
		  AND deleted_at IS NULL`

	args := []any{businessID}
	argIdx := 2

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}

	if filter.Query != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argIdx)
		args = append(args, "%"+filter.Query+"%")
		argIdx++
	}

	switch filter.Sort {
	case "name":
		query += " ORDER BY name ASC"
	case "created_at":
		query += " ORDER BY created_at DESC"
	default: // popular
		query += " ORDER BY total_sold DESC, name ASC"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("product: query all: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var category, description sql.NullString
		err := rows.Scan(
			&p.ID, &p.BusinessID, &p.Name, &p.Price,
			&category, &description,
			&p.IsAvailable, &p.TotalSold, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("product: scan: %w", err)
		}
		p.Category = category.String
		p.Description = description.String
		products = append(products, p)
	}

	return products, nil
}

// FindByID returns a single non-deleted product by ID and business ID.
// Business ID check prevents cross-business access.
func (r *Repository) FindByID(id, businessID uuid.UUID) (*Product, error) {
	query := `
		SELECT id, business_id, name, price, category, description,
		       is_available, total_sold, created_at, updated_at
		FROM products
		WHERE id = $1
		  AND business_id = $2
		  AND deleted_at IS NULL`

	var p Product
	var category, description sql.NullString
	err := r.db.QueryRow(query, id, businessID).Scan(
		&p.ID, &p.BusinessID, &p.Name, &p.Price,
		&category, &description,
		&p.IsAvailable, &p.TotalSold, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("product: find by id: %w", err)
	}
	p.Category = category.String
	p.Description = description.String
	return &p, nil
}

// Create inserts a new product and returns the created record.
func (r *Repository) Create(p *Product) (*Product, error) {
	query := `
		INSERT INTO products (business_id, name, price, category, description, is_available)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, business_id, name, price, category, description,
		          is_available, total_sold, created_at, updated_at`

	var created Product
	var category, description sql.NullString
	err := r.db.QueryRow(query,
		p.BusinessID, p.Name, p.Price,
		sql.NullString{String: p.Category, Valid: p.Category != ""},
		sql.NullString{String: p.Description, Valid: p.Description != ""},
		p.IsAvailable,
	).Scan(
		&created.ID, &created.BusinessID, &created.Name, &created.Price,
		&category, &description,
		&created.IsAvailable, &created.TotalSold, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("product: create: %w", err)
	}
	created.Category = category.String
	created.Description = description.String
	return &created, nil
}

// Update applies partial updates to a product. Only non-nil fields are changed.
func (r *Repository) Update(id, businessID uuid.UUID, input UpdateInput) (*Product, error) {
	query := `
		UPDATE products SET updated_at = NOW()`

	args := []any{}
	argIdx := 1

	if input.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Price != nil {
		query += fmt.Sprintf(", price = $%d", argIdx)
		args = append(args, *input.Price)
		argIdx++
	}
	if input.Category != nil {
		query += fmt.Sprintf(", category = $%d", argIdx)
		args = append(args, *input.Category)
		argIdx++
	}
	if input.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *input.Description)
		argIdx++
	}
	if input.IsAvailable != nil {
		query += fmt.Sprintf(", is_available = $%d", argIdx)
		args = append(args, *input.IsAvailable)
		argIdx++
	}

	query += fmt.Sprintf(`
		WHERE id = $%d AND business_id = $%d AND deleted_at IS NULL
		RETURNING id, business_id, name, price, category, description,
		          is_available, total_sold, created_at, updated_at`,
		argIdx, argIdx+1)
	args = append(args, id, businessID)

	var p Product
	var category, description sql.NullString
	err := r.db.QueryRow(query, args...).Scan(
		&p.ID, &p.BusinessID, &p.Name, &p.Price,
		&category, &description,
		&p.IsAvailable, &p.TotalSold, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("product: update: %w", err)
	}
	p.Category = category.String
	p.Description = description.String
	return &p, nil
}

// SoftDelete sets deleted_at on a product. Data tetap ada untuk histori transaksi.
func (r *Repository) SoftDelete(id, businessID uuid.UUID) error {
	query := `
		UPDATE products
		SET deleted_at = $1, updated_at = NOW()
		WHERE id = $2 AND business_id = $3 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), id, businessID)
	if err != nil {
		return fmt.Errorf("product: soft delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

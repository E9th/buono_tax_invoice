package repository

import (
	"database/sql"
	"fmt"

	"buono-tax-invoice/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// CustomerRepository จัดการ CRUD ของข้อมูลลูกค้า
type CustomerRepository struct {
	db *sqlx.DB
}

// NewCustomerRepository สร้าง Repository ใหม่
func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Search ค้นหาลูกค้าด้วย เลขผู้เสียภาษี หรือ เบอร์โทร
func (r *CustomerRepository) Search(query string) (*models.Customer, error) {
	var customer models.Customer

	err := r.db.Get(&customer, `
		SELECT id, name, tax_id, branch_code, address, 
		       COALESCE(phone_number, '') as phone_number, 
		       created_at, updated_at
		FROM customers 
		WHERE tax_id = $1 OR phone_number = $1
		LIMIT 1
	`, query)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // ไม่เจอข้อมูล
		}
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหา: %w", err)
	}

	return &customer, nil
}

// Create สร้างข้อมูลลูกค้าใหม่
func (r *CustomerRepository) Create(req *models.CustomerRequest) (*models.Customer, error) {
	var customer models.Customer

	// ถ้าไม่กรอกสาขา ให้ใช้ค่า Default
	branchCode := req.BranchCode
	if branchCode == "" {
		branchCode = "00000"
	}

	err := r.db.QueryRowx(`
		INSERT INTO customers (name, tax_id, branch_code, address, phone_number)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, tax_id, branch_code, address, 
		          COALESCE(phone_number, '') as phone_number, 
		          created_at, updated_at
	`, req.Name, req.TaxID, branchCode, req.Address, nullIfEmpty(req.PhoneNumber)).StructScan(&customer)

	if err != nil {
		// เช็ค Duplicate Key Error (เลขภาษีซ้ำ)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("DUPLICATE_TAX_ID")
		}
		return nil, fmt.Errorf("ไม่สามารถบันทึกข้อมูลได้: %w", err)
	}

	return &customer, nil
}

// Update แก้ไขข้อมูลลูกค้า
func (r *CustomerRepository) Update(req *models.CustomerRequest) (*models.Customer, error) {
	var customer models.Customer

	// ถ้าไม่กรอกสาขา ให้ใช้ค่า Default
	branchCode := req.BranchCode
	if branchCode == "" {
		branchCode = "00000"
	}

	err := r.db.QueryRowx(`
		UPDATE customers 
		SET name = $1, tax_id = $2, branch_code = $3, address = $4, phone_number = $5
		WHERE id = $6
		RETURNING id, name, tax_id, branch_code, address, 
		          COALESCE(phone_number, '') as phone_number, 
		          created_at, updated_at
	`, req.Name, req.TaxID, branchCode, req.Address, nullIfEmpty(req.PhoneNumber), req.ID).StructScan(&customer)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("NOT_FOUND")
		}
		// เช็ค Duplicate Key Error (เลขภาษีซ้ำกับรายอื่น)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("DUPLICATE_TAX_ID")
		}
		return nil, fmt.Errorf("ไม่สามารถแก้ไขข้อมูลได้: %w", err)
	}

	return &customer, nil
}

// GetByID ดึงข้อมูลลูกค้าจาก ID
func (r *CustomerRepository) GetByID(id int) (*models.Customer, error) {
	var customer models.Customer

	err := r.db.Get(&customer, `
		SELECT id, name, tax_id, branch_code, address, 
		       COALESCE(phone_number, '') as phone_number, 
		       created_at, updated_at
		FROM customers 
		WHERE id = $1
	`, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงข้อมูล: %w", err)
	}

	return &customer, nil
}

// nullIfEmpty แปลง String ว่างเป็น nil เพื่อเก็บเป็น NULL ใน Database
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

package models

import "time"

// Customer คือ Model หลักสำหรับเก็บข้อมูลลูกค้าที่ใช้ออกใบกำกับภาษี
type Customer struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`                 // ชื่อบริษัท หรือ นามบุคคล (Required)
	TaxID       string    `json:"tax_id" db:"tax_id"`             // เลขผู้เสียภาษี 13 หลัก (Required)
	BranchCode  string    `json:"branch_code" db:"branch_code"`   // สาขา (Default: 00000)
	Address     string    `json:"address" db:"address"`           // ที่อยู่ (Required)
	PhoneNumber string    `json:"phone_number" db:"phone_number"` // เบอร์โทร
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CustomerRequest คือ DTO สำหรับรับข้อมูลจาก Frontend (Create / Update)
type CustomerRequest struct {
	ID          int    `json:"id"`                               // ถ้ามี ID = Update, ถ้าไม่มี = Create
	Name        string `json:"name" binding:"required"`          // Required (สีแดง)
	TaxID       string `json:"tax_id" binding:"required,len=13"` // Required 13 หลัก (สีแดง)
	BranchCode  string `json:"branch_code"`                      // Optional (Default: 00000)
	Address     string `json:"address" binding:"required"`       // Required (สีแดง)
	PhoneNumber string `json:"phone_number"`                     // Optional
}

// SearchResponse คือ Response สำหรับการค้นหา
type SearchResponse struct {
	Found    bool      `json:"found"`
	Customer *Customer `json:"customer,omitempty"`
}

// APIResponse คือ Response มาตรฐานของ API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse คือ Response สำหรับ Error
type ErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

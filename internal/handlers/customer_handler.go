package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"buono-tax-invoice/internal/models"
	"buono-tax-invoice/internal/repository"

	"github.com/gin-gonic/gin"
)

// CustomerHandler จัดการ HTTP Request ทั้งหมดที่เกี่ยวกับลูกค้า
type CustomerHandler struct {
	repo *repository.CustomerRepository
}

// NewCustomerHandler สร้าง Handler ใหม่
func NewCustomerHandler(repo *repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

// ==========================================
// GET /api/customer/search?q=xxxxx
// ค้นหาลูกค้าด้วย เลขผู้เสียภาษี หรือ เบอร์โทร
// ==========================================
func (h *CustomerHandler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))

	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "กรุณากรอกเลขผู้เสียภาษี หรือ เบอร์โทร เพื่อค้นหา",
		})
		return
	}

	customer, err := h.repo.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Message: "เกิดข้อผิดพลาดในการค้นหา",
		})
		return
	}

	if customer == nil {
		c.JSON(http.StatusOK, models.SearchResponse{
			Found:    false,
			Customer: nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.SearchResponse{
		Found:    true,
		Customer: customer,
	})
}

// ==========================================
// POST /api/customer
// บันทึกข้อมูลลูกค้า (Create หรือ Update)
// - ถ้ามี ID -> Update
// - ถ้าไม่มี ID -> Create
// ==========================================
func (h *CustomerHandler) Save(c *gin.Context) {
	var req models.CustomerRequest

	// Parse JSON Body
	if err := c.ShouldBindJSON(&req); err != nil {
		errors := parseValidationErrors(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบอีกครั้ง",
			Errors:  errors,
		})
		return
	}

	// ===== Custom Validation =====
	validationErrors := validateCustomerRequest(&req)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบอีกครั้ง",
			Errors:  validationErrors,
		})
		return
	}

	// ===== Upsert Logic =====
	if req.ID > 0 {
		// === UPDATE ===
		customer, err := h.repo.Update(&req)
		if err != nil {
			handleRepoError(c, err, "แก้ไข")
			return
		}

		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "แก้ไขข้อมูลลูกค้าสำเร็จ",
			Data:    customer,
		})
	} else {
		// === CREATE ===
		customer, err := h.repo.Create(&req)
		if err != nil {
			handleRepoError(c, err, "บันทึก")
			return
		}

		c.JSON(http.StatusCreated, models.APIResponse{
			Success: true,
			Message: "บันทึกข้อมูลลูกค้าสำเร็จ",
			Data:    customer,
		})
	}
}

// ==========================================
// Validation Helpers
// ==========================================

// validateCustomerRequest ตรวจสอบความถูกต้องของข้อมูลเพิ่มเติม
func validateCustomerRequest(req *models.CustomerRequest) map[string]string {
	errors := make(map[string]string)

	// Trim ช่องว่าง
	req.Name = strings.TrimSpace(req.Name)
	req.TaxID = strings.TrimSpace(req.TaxID)
	req.BranchCode = strings.TrimSpace(req.BranchCode)
	req.Address = strings.TrimSpace(req.Address)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)

	// ตรวจสอบชื่อบริษัท
	if req.Name == "" {
		errors["name"] = "กรุณากรอกชื่อบริษัท หรือ นามบุคคล"
	} else if utf8.RuneCountInString(req.Name) > 255 {
		errors["name"] = "ชื่อบริษัทต้องไม่เกิน 255 ตัวอักษร"
	}

	// ตรวจสอบเลขผู้เสียภาษี (ต้องเป็นตัวเลข 13 หลักเท่านั้น)
	taxIDRegex := regexp.MustCompile(`^\d{13}$`)
	if req.TaxID == "" {
		errors["tax_id"] = "กรุณากรอกเลขผู้เสียภาษี"
	} else if !taxIDRegex.MatchString(req.TaxID) {
		errors["tax_id"] = "เลขผู้เสียภาษีต้องเป็นตัวเลข 13 หลัก"
	}

	// ตรวจสอบที่อยู่
	if req.Address == "" {
		errors["address"] = "กรุณากรอกที่อยู่"
	}

	// ตรวจสอบเบอร์โทร (ถ้ากรอกมา ต้องเป็นรูปแบบที่ถูกต้อง)
	if req.PhoneNumber != "" {
		phoneRegex := regexp.MustCompile(`^[\d\-\+]{9,15}$`)
		if !phoneRegex.MatchString(req.PhoneNumber) {
			errors["phone_number"] = "รูปแบบเบอร์โทรไม่ถูกต้อง (ตัวอย่าง: 0812345678)"
		}
	}

	return errors
}

// parseValidationErrors แปลง Binding Error ให้เป็น Map ที่อ่านง่าย
func parseValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	errMsg := err.Error()
	if strings.Contains(errMsg, "'Name'") || strings.Contains(errMsg, "'required'") {
		if strings.Contains(errMsg, "Name") {
			errors["name"] = "กรุณากรอกชื่อบริษัท หรือ นามบุคคล"
		}
		if strings.Contains(errMsg, "TaxID") {
			errors["tax_id"] = "กรุณากรอกเลขผู้เสียภาษี"
		}
		if strings.Contains(errMsg, "Address") {
			errors["address"] = "กรุณากรอกที่อยู่"
		}
	}

	if len(errors) == 0 {
		errors["general"] = "ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบ"
	}

	return errors
}

// handleRepoError จัดการ Error จาก Repository
func handleRepoError(c *gin.Context, err error, action string) {
	switch err.Error() {
	case "DUPLICATE_TAX_ID":
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Success: false,
			Message: "เลขผู้เสียภาษีนี้มีในระบบแล้ว",
			Errors: map[string]string{
				"tax_id": "เลขผู้เสียภาษีนี้ถูกใช้งานแล้ว กรุณาตรวจสอบอีกครั้ง",
			},
		})
	case "NOT_FOUND":
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Message: "ไม่พบข้อมูลลูกค้าที่ต้องการ" + action,
		})
	default:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Message: "เกิดข้อผิดพลาดในการ" + action + "ข้อมูล",
		})
	}
}

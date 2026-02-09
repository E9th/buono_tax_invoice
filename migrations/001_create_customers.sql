-- =====================================================
-- Buono Dine & Wine - Tax Invoice System
-- Database: buono_tax_invoice
-- =====================================================

-- สร้าง Database (รันคำสั่งนี้แยกถ้าจำเป็น)
-- CREATE DATABASE buono_tax_invoice;

-- =====================================================
-- ตาราง customers: เก็บข้อมูลลูกค้าสำหรับออกใบกำกับภาษี
-- =====================================================
CREATE TABLE IF NOT EXISTS customers (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,                          -- ชื่อบริษัท หรือ นามบุคคล (Required)
    tax_id          VARCHAR(13) NOT NULL UNIQUE,                    -- เลขผู้เสียภาษี 13 หลัก (Required, Unique)
    branch_code     VARCHAR(50) NOT NULL DEFAULT '00000',           -- รหัสสาขา (Default: 00000 = สำนักงานใหญ่)
    address         TEXT NOT NULL,                                  -- ที่อยู่ (Required)
    phone_number    VARCHAR(20),                                    -- เบอร์โทร
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- Indexes: เพิ่มประสิทธิภาพการค้นหา
-- =====================================================

-- Index สำหรับค้นหาด้วย เลขผู้เสียภาษี
CREATE INDEX IF NOT EXISTS idx_customers_tax_id ON customers (tax_id);

-- Index สำหรับค้นหาด้วย เบอร์โทร
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers (phone_number);

-- =====================================================
-- Function: อัปเดต updated_at อัตโนมัติเมื่อมีการแก้ไขข้อมูล
-- =====================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_customers_timestamp
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- ข้อมูลตัวอย่าง (Sample Data) สำหรับทดสอบ
-- =====================================================
INSERT INTO customers (name, tax_id, branch_code, address, phone_number) VALUES
    ('บริษัท บัวโน่ จำกัด', '0105564012345', '00000', '123/45 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร 10110', '0812345678'),
    ('นาย สมชาย ใจดี', '1234567890123', '00000', '789 หมู่ 5 ตำบลบางพลี อำเภอบางพลี จังหวัดสมุทรปราการ 10540', '0891234567'),
    ('บริษัท ไทยฟู้ด จำกัด', '0105567098765', '00001', '456/78 ถนนเพชรบุรี แขวงทุ่งพญาไท เขตราชเทวี กรุงเทพมหานคร 10400', '0623456789')
ON CONFLICT (tax_id) DO NOTHING;

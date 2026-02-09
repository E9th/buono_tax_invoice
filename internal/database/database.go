package database

import (
	"fmt"
	"log"

	"buono-tax-invoice/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewConnection สร้าง Database Connection ใหม่
func NewConnection(cfg *config.Config) (*sqlx.DB, error) {
	dsn := cfg.GetDSN()

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ Database ได้: %w", err)
	}

	// ตั้งค่า Connection Pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// ทดสอบ Connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ไม่สามารถ Ping Database ได้: %w", err)
	}

	log.Println("✅ เชื่อมต่อ PostgreSQL สำเร็จ")
	return db, nil
}

// RunMigrations รัน SQL Migration เพื่อสร้างตาราง (ถ้ายังไม่มี)
func RunMigrations(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS customers (
		id              SERIAL PRIMARY KEY,
		name            VARCHAR(255) NOT NULL,
		tax_id          VARCHAR(13) NOT NULL UNIQUE,
		branch_code     VARCHAR(50) NOT NULL DEFAULT '00000',
		address         TEXT NOT NULL,
		phone_number    VARCHAR(20),
		created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_customers_tax_id ON customers (tax_id);
	CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers (phone_number);

	-- Function สำหรับ auto-update updated_at
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- สร้าง Trigger (ถ้ายังไม่มี)
	DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_update_customers_timestamp'
		) THEN
			CREATE TRIGGER trigger_update_customers_timestamp
				BEFORE UPDATE ON customers
				FOR EACH ROW
				EXECUTE FUNCTION update_updated_at_column();
		END IF;
	END
	$$;
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("ไม่สามารถรัน Migration ได้: %w", err)
	}

	log.Println("✅ รัน Database Migration สำเร็จ")
	return nil
}

package model

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"testing"
)

type unboundedAmountProbe struct {
	ID     int64   `gorm:"primaryKey"`
	Amount float64 `gorm:"type:numeric;default:0"`
}

func (unboundedAmountProbe) TableName() string { return "unbounded_amount_probe" }

func TestPostgresUnboundedNumericContract(t *testing.T) {
	for _, guard := range []bool{false, true} {
		t.Run(map[bool]string{false: "real_difference", true: "dependent_trigger"}[guard], func(t *testing.T) {
			db, sql := postgresMigrationTestDB(t)
			if err := db.Exec("CREATE TABLE unbounded_amount_probe(id bigserial PRIMARY KEY,amount numeric(10,2) DEFAULT 0)").Error; err != nil {
				t.Fatal(err)
			}
			if guard {
				if err := db.Exec(`CREATE FUNCTION amount_guard() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN RETURN NEW;END$$; CREATE TRIGGER amount_guard BEFORE UPDATE OF amount ON unbounded_amount_probe FOR EACH ROW EXECUTE FUNCTION amount_guard()`).Error; err != nil {
					t.Fatal(err)
				}
			}
			sql.ddl = nil
			if err := db.AutoMigrate(&unboundedAmountProbe{}); err != nil {
				t.Fatal(err)
			}
			if len(sql.ddl) != 0 {
				t.Fatalf("numeric 精度差异必须延后，不能启动改型: %v", sql.ddl)
			}
			var typ string
			if err := db.Raw("SELECT format_type(atttypid,atttypmod) FROM pg_attribute WHERE attrelid='unbounded_amount_probe'::regclass AND attname='amount'").Scan(&typ).Error; err != nil {
				t.Fatal(err)
			}
			if typ != "numeric(10,2)" {
				t.Fatalf("启动改变了精度: %s", typ)
			}
			err := db.Exec("ALTER TABLE unbounded_amount_probe ALTER COLUMN amount TYPE numeric USING amount::numeric").Error
			if guard {
				var pgerr *pgconn.PgError
				if !errors.As(err, &pgerr) || pgerr.Code != "0A000" {
					t.Fatalf("显式迁移依赖错误必须透传: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			const wanted = "0.12345678901234567890123456789"
			if err := db.Exec("INSERT INTO unbounded_amount_probe(amount) VALUES(?)", wanted).Error; err != nil {
				t.Fatal(err)
			}
			var got string
			if err := db.Raw("SELECT amount::text FROM unbounded_amount_probe").Scan(&got).Error; err != nil || got != wanted {
				t.Fatalf("高精度金额被改写: %s %v", got, err)
			}
			sql.ddl = nil
			if err := db.AutoMigrate(&unboundedAmountProbe{}); err != nil {
				t.Fatal(err)
			}
			if len(sql.ddl) != 0 {
				t.Fatalf("numeric迁移必须收敛: %v", sql.ddl)
			}
		})
	}
}

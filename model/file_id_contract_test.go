package model

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm/schema"
)

func TestFilesTempFileIDSQLContract(t *testing.T) {
	s, err := schema.Parse(&FilesTemp{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	f := s.LookUpField("FileId")
	if string(f.DataType) != "bigint" || !f.NotNull || !f.HasDefaultValue || f.DefaultValue != "0" {
		t.Fatalf("文件临时引用须保持 bigint NOT NULL DEFAULT 0: %+v", f)
	}
}

func TestFilesTempFileIDMigrationContract(t *testing.T) {
	for _, kind := range []string{"same", "metadata", "real_type", "real_type_with_trigger", "null_row"} {
		t.Run(kind, func(t *testing.T) {
			db, queries := postgresMigrationTestDB(t)
			migrate := func() error { return db.Table("files_temp_contract").AutoMigrate(&FilesTemp{}) }
			if err := migrate(); err != nil {
				t.Fatal(err)
			}
			if kind == "metadata" || kind == "null_row" {
				if err := db.Exec("ALTER TABLE files_temp_contract ALTER COLUMN file_id DROP NOT NULL, ALTER COLUMN file_id DROP DEFAULT").Error; err != nil {
					t.Fatal(err)
				}
			}
			if kind == "null_row" {
				if err := db.Exec("INSERT INTO files_temp_contract(file_id) VALUES(NULL)").Error; err != nil {
					t.Fatal(err)
				}
			}
			if strings.HasPrefix(kind, "real_type") {
				if err := db.Exec("ALTER TABLE files_temp_contract ALTER COLUMN file_id TYPE integer USING file_id::integer").Error; err != nil {
					t.Fatal(err)
				}
			}
			if kind != "real_type" {
				if err := db.Exec(`CREATE FUNCTION file_id_guard() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$;
CREATE TRIGGER knowledge_file_reference_guard BEFORE INSERT OR UPDATE OF file_id ON files_temp_contract FOR EACH ROW EXECUTE FUNCTION file_id_guard()`).Error; err != nil {
					t.Fatal(err)
				}
			}
			queries.ddl = nil
			err := migrate()
			expected := ""
			if kind == "real_type_with_trigger" {
				expected = "0A000"
			}
			if kind == "null_row" {
				expected = "23502"
			}
			if expected != "" {
				var pgerr *pgconn.PgError
				if !errors.As(err, &pgerr) || pgerr.Code != expected {
					t.Fatalf("真实差异必须返回 %s: %v", expected, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if kind == "same" && len(queries.ddl) != 0 {
				t.Fatalf("一致结构不能执行 DDL: %v", queries.ddl)
			}
			if kind == "metadata" {
				for _, q := range queries.ddl {
					if strings.Contains(q, " TYPE ") {
						t.Fatalf("元数据差异不能改型: %s", q)
					}
				}
				if len(queries.ddl) != 2 {
					t.Fatalf("必须修复默认值与非空: %v", queries.ddl)
				}
			}
			if err := db.Exec("INSERT INTO files_temp_contract(file_id) VALUES(4294967295)").Error; err != nil {
				t.Fatalf("不得收窄历史文件引用范围: %v", err)
			}
			var actual uint64
			if err := db.Raw("SELECT max(file_id) FROM files_temp_contract").Scan(&actual).Error; err != nil || actual != 4294967295 {
				t.Fatalf("文件引用未完整保存: %d %v", actual, err)
			}
			queries.ddl = nil
			if err := migrate(); err != nil {
				t.Fatal(err)
			}
			if len(queries.ddl) != 0 {
				t.Fatalf("重复迁移应收敛: %v", queries.ddl)
			}
		})
	}
}

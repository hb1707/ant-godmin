package model

import (
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type charTypeProbe struct {
	Digest  string `gorm:"type:char(64)"`
	Code    string `gorm:"type:CHARACTER(8)"`
	Varying string `gorm:"type:character varying(32)"`
	Name    string `gorm:"type:varchar(64)"`
}

// 各方言下 char(n) 的声明类型：PostgreSQL 换成 bpchar(n) 以免每次启动都 ALTER COLUMN TYPE，其余方言原样保留。
func TestNormalizePostgresCharTypes(t *testing.T) {
	cases := []struct {
		name      string
		dialector gorm.Dialector
		want      map[string]string
	}{
		{"postgres", postgres.New(postgres.Config{DSN: "host=127.0.0.1 port=1 dbname=none"}), map[string]string{
			"Digest": "bpchar(64)", "Code": "bpchar(8)", "Varying": "character varying(32)", "Name": "varchar(64)",
		}},
		{"mysql", mysql.New(mysql.Config{DSN: "u:p@tcp(127.0.0.1:1)/none", SkipInitializeWithVersion: true}), map[string]string{
			"Digest": "char(64)", "Code": "CHARACTER(8)", "Varying": "character varying(32)", "Name": "varchar(64)",
		}},
	}
	oldDB := DB
	t.Cleanup(func() { DB = oldDB })
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(tc.dialector, &gorm.Config{DisableAutomaticPing: true})
			if err != nil {
				t.Fatal(err)
			}
			DB = db
			if err := normalizePostgresCharTypes(&charTypeProbe{}); err != nil {
				t.Fatal(err)
			}
			// 与 AutoMigrate 读取同一份 schema 缓存
			stmt := &gorm.Statement{DB: db}
			if err := stmt.Parse(&charTypeProbe{}); err != nil {
				t.Fatal(err)
			}
			for field, want := range tc.want {
				if got := string(stmt.Schema.LookUpField(field).DataType); got != want {
					t.Errorf("%s: DataType = %q, want %q", field, got, want)
				}
			}
		})
	}
}

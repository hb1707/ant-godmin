package model

import (
	"context"
	"encoding/json"
)

// PostgresMigrationDifference 只描述被延后的结构变化，不包含行数据、默认值表达式或连接配置。
type PostgresMigrationDifference struct {
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	Column    string `json:"column"`
	Operation string `json:"operation"`
	Actual    string `json:"actual"`
	Desired   string `json:"desired"`
}

// PostgresMigrationObserver 由应用 logger 可选实现；回调必须立即返回，不得执行同步通知或数据库写入。
// 回调载荷为 PostgresMigrationDifference 的 JSON 数组；使用标准类型保持旧依赖编译兼容。
// 框架只产出差异，通知通道及去重策略由消费方负责。
type PostgresMigrationObserver interface {
	ReportDeferredPostgresMigrations([]byte)
}

type postgresMigrationBatchKey struct{}

type postgresMigrationLocation struct{ schema, table string }

type postgresMigrationBatch struct {
	locations   map[string]postgresMigrationLocation
	differences []PostgresMigrationDifference
	seen        map[PostgresMigrationDifference]bool
}

// AutoMigrate 在真实迁移调用边界汇总差异；错误原样返回，通知不是迁移成功的条件。
func (m postgresColumnMigrator) AutoMigrate(values ...interface{}) error {
	observer, ok := m.DB.Logger.(PostgresMigrationObserver)
	if !ok {
		return m.Migrator.AutoMigrate(values...)
	}
	batch := &postgresMigrationBatch{seen: make(map[PostgresMigrationDifference]bool), locations: make(map[string]postgresMigrationLocation)}
	upstream := m.Migrator
	upstream.DB = m.DB.WithContext(context.WithValue(m.DB.Statement.Context, postgresMigrationBatchKey{}, batch))
	err := upstream.AutoMigrate(values...)
	if len(batch.differences) > 0 && !m.DB.DryRun {
		payload, _ := json.Marshal(batch.differences)
		observer.ReportDeferredPostgresMigrations(payload)
	}
	return err
}

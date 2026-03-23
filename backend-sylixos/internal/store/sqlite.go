package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// DB 封装 SQLite 连接
type DB struct {
	db *sql.DB
}

// MetricRow 历史指标记录
type MetricRow struct {
	Timestamp int64   `json:"ts"`
	CPU       float64 `json:"cpu"`
	MemPct    float64 `json:"mem_pct"`
	DiskPct   float64 `json:"disk_pct"`
	NetIn     uint64  `json:"net_in"`
	NetOut    uint64  `json:"net_out"`
}

// Open 打开或创建数据库
func Open(dir string) (*DB, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	path := filepath.Join(dir, "edge-setting.db")
	db, err := sql.Open("sqlite3", path+"?_journal=WAL&_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &DB{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

// migrate 创建表结构
func (s *DB) migrate() error {
	ddl := `
CREATE TABLE IF NOT EXISTS metrics (
	ts       INTEGER NOT NULL,
	cpu      REAL,
	mem_pct  REAL,
	disk_pct REAL,
	net_in   INTEGER,
	net_out  INTEGER
);
CREATE INDEX IF NOT EXISTS idx_metrics_ts ON metrics(ts);
`
	_, err := s.db.Exec(ddl)
	return err
}

// InsertMetric 写入一条指标
func (s *DB) InsertMetric(r MetricRow) error {
	_, err := s.db.Exec(
		`INSERT INTO metrics(ts,cpu,mem_pct,disk_pct,net_in,net_out) VALUES(?,?,?,?,?,?)`,
		r.Timestamp, r.CPU, r.MemPct, r.DiskPct, r.NetIn, r.NetOut,
	)
	return err
}

// QueryMetrics 按时间范围查询指标
func (s *DB) QueryMetrics(metricType string, from, to int64) ([]MetricRow, error) {
	rows, err := s.db.Query(
		`SELECT ts,cpu,mem_pct,disk_pct,net_in,net_out FROM metrics WHERE ts>=? AND ts<=? ORDER BY ts`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []MetricRow
	for rows.Next() {
		var r MetricRow
		if err := rows.Scan(&r.Timestamp, &r.CPU, &r.MemPct, &r.DiskPct, &r.NetIn, &r.NetOut); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// Cleanup 清理过期数据
func (s *DB) Cleanup(metricsRetainDays int) error {
	now := time.Now().UnixMilli()
	cutMetrics := now - int64(metricsRetainDays)*86400*1000

	if _, err := s.db.Exec(`DELETE FROM metrics WHERE ts<?`, cutMetrics); err != nil {
		return err
	}
	return nil
}

// Close 关闭数据库
func (s *DB) Close() error {
	return s.db.Close()
}

package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
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

// LogRow 应用日志记录
type LogRow struct {
	ID        int64  `json:"id"`
	AppID     string `json:"app_id"`
	Level     string `json:"level"`
	Content   string `json:"content"`
	Timestamp int64  `json:"ts"`
}

// AuditRow 审计日志记录
type AuditRow struct {
	ID        int64  `json:"id"`
	Token     string `json:"token"`
	AppID     string `json:"app_id"`
	Action    string `json:"action"`
	Result    string `json:"result"`
	Timestamp int64  `json:"ts"`
}

// Open 打开或创建数据库
func Open(dir string) (*DB, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	path := filepath.Join(dir, "edge-setting.db")
	db, err := sql.Open("sqlite", path+"?_journal=WAL&_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite 单写
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

CREATE TABLE IF NOT EXISTS app_logs (
	id        INTEGER PRIMARY KEY AUTOINCREMENT,
	app_id    TEXT NOT NULL,
	level     TEXT,
	content   TEXT,
	ts        INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_logs_app_ts ON app_logs(app_id, ts);

CREATE TABLE IF NOT EXISTS audit_logs (
	id     INTEGER PRIMARY KEY AUTOINCREMENT,
	token  TEXT,
	app_id TEXT,
	action TEXT,
	result TEXT,
	ts     INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_ts ON audit_logs(ts);
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

// QueryMetrics 按时间范围查询指标（自动聚合）
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

// InsertLog 写入一条应用日志
func (s *DB) InsertLog(r LogRow) error {
	_, err := s.db.Exec(
		`INSERT INTO app_logs(app_id,level,content,ts) VALUES(?,?,?,?)`,
		r.AppID, r.Level, r.Content, r.Timestamp,
	)
	return err
}

// QueryLogs 查询历史日志
func (s *DB) QueryLogs(appID, keyword, level string, from, to int64, limit int) ([]LogRow, error) {
	query := `SELECT id,app_id,level,content,ts FROM app_logs WHERE app_id=? AND ts>=? AND ts<=?`
	args := []interface{}{appID, from, to}
	if level != "" {
		query += " AND level=?"
		args = append(args, level)
	}
	if keyword != "" {
		query += " AND content LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	query += " ORDER BY ts DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []LogRow
	for rows.Next() {
		var r LogRow
		if err := rows.Scan(&r.ID, &r.AppID, &r.Level, &r.Content, &r.Timestamp); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// InsertAudit 写入审计日志
func (s *DB) InsertAudit(r AuditRow) error {
	_, err := s.db.Exec(
		`INSERT INTO audit_logs(token,app_id,action,result,ts) VALUES(?,?,?,?,?)`,
		r.Token, r.AppID, r.Action, r.Result, r.Timestamp,
	)
	return err
}

// QueryAudit 查询审计日志
func (s *DB) QueryAudit(op string, from, to int64, limit int) ([]AuditRow, error) {
	query := `SELECT id,token,app_id,action,result,ts FROM audit_logs WHERE ts>=? AND ts<=?`
	args := []interface{}{from, to}
	if op != "" {
		query += " AND action=?"
		args = append(args, op)
	}
	query += " ORDER BY ts DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []AuditRow
	for rows.Next() {
		var r AuditRow
		if err := rows.Scan(&r.ID, &r.Token, &r.AppID, &r.Action, &r.Result, &r.Timestamp); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// Cleanup 清理过期数据
func (s *DB) Cleanup(metricsRetainDays, logsRetainDays, auditRetainDays int) error {
	now := time.Now().UnixMilli()
	cutMetrics := now - int64(metricsRetainDays)*86400*1000
	cutLogs := now - int64(logsRetainDays)*86400*1000
	cutAudit := now - int64(auditRetainDays)*86400*1000

	if _, err := s.db.Exec(`DELETE FROM metrics WHERE ts<?`, cutMetrics); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM app_logs WHERE ts<?`, cutLogs); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM audit_logs WHERE ts<?`, cutAudit); err != nil {
		return err
	}
	return nil
}

// Close 关闭数据库
func (s *DB) Close() error {
	return s.db.Close()
}

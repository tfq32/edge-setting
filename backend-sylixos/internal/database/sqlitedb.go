package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type SqliteDB struct {
	db *sql.DB
}

// newSqliteDB 初始化数据库连接
func newSqliteDB(dbPath string) (*SqliteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(5 * time.Minute)
	// SylixOS 兼容性：关闭 WAL 和 mmap
	pragmas := []string{
		"PRAGMA journal_mode = DELETE",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA mmap_size = 0",
		"PRAGMA locking_mode = EXCLUSIVE",
		"PRAGMA temp_store = MEMORY",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			log.Printf("PRAGMA 执行失败 [%s]: %v", p, err)
		} else {
			log.Printf("PRAGMA 执行成功: %s", p)
		}
	}
	log.Printf("数据库连接成功: %s", dbPath)
	return &SqliteDB{db: db}, nil
}

// Close 关闭数据库连接
func (s *SqliteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Query 查询多行
func (s *SqliteDB) Query(querySql string, args ...any) ([]map[string]any, error) {
	rows, err := s.db.Query(querySql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var results []map[string]any
	for rows.Next() {
		rowMap, err := s.scanRow(rows, columns)
		if err != nil {
			return nil, err
		}
		results = append(results, rowMap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// QueryOne 查询单行
func (s *SqliteDB) QueryOne(querySql string, args ...any) (map[string]any, error) {
	rows, err := s.db.Query(querySql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return s.scanRow(rows, columns)
}

// Exec 执行SQL语句（INSERT/UPDATE/DELETE/CREATE等）
func (s *SqliteDB) Exec(querySql string, args ...any) (sql.Result, error) {
	return s.db.Exec(querySql, args...)
}

// scanRow 扫描单行数据并转换为map
func (s *SqliteDB) scanRow(rows *sql.Rows, columns []string) (map[string]any, error) {
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}
	rowMap := make(map[string]any)
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			rowMap[col] = string(b)
		} else {
			rowMap[col] = val
		}
	}
	return rowMap, nil
}

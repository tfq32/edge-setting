package database

import (
	"log"
	"os"
	"path/filepath"

	"go-ser/internal/config"
)

var EdgeDB *SqliteDB

func Start() error {
	dir := config.Global.Data.Dir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	dbPath := filepath.Join(dir, "edge-setting.db")

	var err error
	EdgeDB, err = newSqliteDB(dbPath)
	if err != nil {
		return err
	}
	if err := createTables(*EdgeDB); err != nil {
		return err
	}
	return nil
}

func Stop() {
	if EdgeDB != nil {
		EdgeDB.Close()
	}
}

func createTables(db SqliteDB) error {
	// 配置表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS conf (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE,
			value TEXT
		);
	`)
	if err != nil {
		log.Printf("创建配置表失败: %v", err)
		return err
	}

	// 指标历史表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			ts       INTEGER NOT NULL,
			cpu      REAL,
			mem_pct  REAL,
			disk_pct REAL,
			net_in   INTEGER,
			net_out  INTEGER
		);
	`)
	if err != nil {
		log.Printf("创建指标历史表失败: %v", err)
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_metrics_ts ON metrics(ts);`)
	if err != nil {
		log.Printf("创建指标历史表索引失败: %v", err)
		return err
	}
	return nil
}

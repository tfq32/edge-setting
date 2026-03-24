package database

import (
	"fmt"

	"go-ser/internal/config"
)

var EdgeDB *SqliteDB

func Start() error {
	var err error
	EdgeDB, err = newSqliteDB(config.AppConfig.Database.EdgePath)
	if err != nil {
		return err
	}
	// 先测试一个简单写操作，确认 I/O 是否正常
	if _, err := EdgeDB.Exec("CREATE TABLE IF NOT EXISTS _ping (id INTEGER)"); err != nil {
		return fmt.Errorf("数据库写测试失败: %w", err)
	}
	EdgeDB.Exec("DROP TABLE IF EXISTS _ping")

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
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_metrics_ts ON metrics(ts);`)
	return err
}

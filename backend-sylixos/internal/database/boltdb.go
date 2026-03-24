package database

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketConf    = []byte("conf")
	bucketMetrics = []byte("metrics")
)

// BoltDB 封装 bbolt 连接
type BoltDB struct {
	db *bolt.DB
	mu sync.RWMutex
}

// newBoltDB 初始化 bbolt 数据库
func newBoltDB(dbPath string) (*BoltDB, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{
		Timeout:  5 * time.Second,
		NoSync:   false,
		NoGrowSync: true,
	})
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 创建 bucket
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketConf); err != nil {
			return fmt.Errorf("创建 conf bucket 失败: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists(bucketMetrics); err != nil {
			return fmt.Errorf("创建 metrics bucket 失败: %w", err)
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	log.Printf("数据库连接成功: %s", dbPath)
	return &BoltDB{db: db}, nil
}

// Close 关闭数据库
func (b *BoltDB) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// ── conf 操作 ─────────────────────────────────────────────

// SetConf 设置配置项
func (b *BoltDB) SetConf(key, value string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketConf)
		return bkt.Put([]byte(key), []byte(value))
	})
}

// GetConf 获取配置项
func (b *BoltDB) GetConf(key string) (string, error) {
	var val string
	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketConf)
		v := bkt.Get([]byte(key))
		if v == nil {
			return fmt.Errorf("key not found: %s", key)
		}
		val = string(v)
		return nil
	})
	return val, err
}

// GetAllConf 获取所有配置项
func (b *BoltDB) GetAllConf() (map[string]string, error) {
	result := make(map[string]string)
	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketConf)
		return bkt.ForEach(func(k, v []byte) error {
			result[string(k)] = string(v)
			return nil
		})
	})
	return result, err
}

// ── metrics 操作 ──────────────────────────────────────────

// MetricRow 指标记录
type MetricRow struct {
	Timestamp int64   `json:"ts"`
	CPU       float64 `json:"cpu"`
	MemPct    float64 `json:"mem_pct"`
	DiskPct   float64 `json:"disk_pct"`
	NetIn     uint64  `json:"net_in"`
	NetOut    uint64  `json:"net_out"`
}

// InsertMetric 写入一条指标
func (b *BoltDB) InsertMetric(r MetricRow) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	// 用时间戳作为 key，保证有序
	key := fmt.Sprintf("%020d", r.Timestamp)
	return b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketMetrics)
		return bkt.Put([]byte(key), data)
	})
}

// QueryMetrics 按时间范围查询指标
func (b *BoltDB) QueryMetrics(from, to int64) ([]MetricRow, error) {
	var result []MetricRow
	fromKey := fmt.Sprintf("%020d", from)
	toKey := fmt.Sprintf("%020d", to)

	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketMetrics)
		c := bkt.Cursor()
		for k, v := c.Seek([]byte(fromKey)); k != nil; k, v = c.Next() {
			if string(k) > toKey {
				break
			}
			var r MetricRow
			if err := json.Unmarshal(v, &r); err != nil {
				continue
			}
			result = append(result, r)
		}
		return nil
	})
	return result, err
}

// CleanupMetrics 清理过期指标数据
func (b *BoltDB) CleanupMetrics(retainDays int) error {
	if retainDays <= 0 {
		retainDays = 7
	}
	cutoff := time.Now().UnixMilli() - int64(retainDays)*86400*1000
	cutoffKey := fmt.Sprintf("%020d", cutoff)

	return b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketMetrics)
		c := bkt.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if string(k) >= cutoffKey {
				break
			}
			if err := c.Delete(); err != nil {
				return err
			}
		}
		return nil
	})
}

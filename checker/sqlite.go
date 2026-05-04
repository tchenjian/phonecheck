package checker

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// sqliteChecker 使用 SQLite 数据库实现手机号存在性判断
//
// 【10亿数据规模复杂度分析】
//
// 空间复杂度：O(N * 行存储开销)
//   - 每条记录：手机号(11B) + B树索引(~20B) = ~31B
//   - 10亿条磁盘占用：~31GB
//   - 实际内存开销：<10MB (数据在磁盘，按需加载)
//
// 时间复杂度：O(log N)，B树索引查询
//   - 查询操作：磁盘IO + B树查找，毫秒级 (~1-10ms)
//   - 加载时间：<100ms (只打开数据库连接)
//   - 并发查询：受限于磁盘IO
//
// 特性：
//   - ✅ 100% 精确，无假阳性/假阴性
//   - ✅ 内存占用可忽略不计
//   - ❌ 查询速度慢，毫秒级
//   - ❌ 需要 CGO 支持
//   - ✅ 支持增量更新
//
// 适用场景：数据量远超内存，查询频率低，冷数据查询

type sqliteChecker struct {
	db *sql.DB
}

func (sc *sqliteChecker) PhoneExists(phone string) bool {
	var exists bool
	query := "SELECT 1 FROM phones WHERE number = ? LIMIT 1"
	err := sc.db.QueryRow(query, phone).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (sc *sqliteChecker) Close() error {
	return sc.db.Close()
}

func NewSqliteChecker(dataFile string) (*sqliteChecker, error) {
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return nil, os.ErrNotExist
	}
	db, err := sql.Open("sqlite3", dataFile)
	if err != nil {
		return nil, err
	}

	return &sqliteChecker{db: db}, nil
}

package checker

import (
	"os"

	"github.com/bits-and-blooms/bloom/v3"
)

// bloomChecker 使用布隆过滤器实现手机号存在性判断
//
// 【10亿数据规模复杂度分析】
//
// 空间复杂度：O(-N * ln(P) / (ln2)^2)
//   - 误判率 1% 所需大小：~115MB
//   - 误判率 0.1% 所需大小：~172MB
//   - 误判率 0.01% 所需大小：~230MB
//   - 实际内存开销：~115MB (1%误判率)
//
// 时间复杂度：O(k)，k为哈希函数数量
//   - 查询操作：k次哈希+位运算，纳秒级 (~50ns)
//   - k ≈ 7 (1%误判率下最优哈希次数)
//   - 加载时间：~1秒 (从磁盘读取)
//
// 特性：
//   - ✅ 内存占用最低
//   - ✅ 查询速度极快
//   - ⚠️  存在假阳性，无假阴性
//   -   "不存在的一定不存在，存在的可能不存在"
//   - ✅ 支持高效序列化
//
// 适用场景：黑名单过滤、缓存预热等可接受假阳性的场景

type bloomChecker struct {
	filter *bloom.BloomFilter
}

func (bc *bloomChecker) PhoneExists(phone string) bool {
	return bc.filter.TestString(phone)
}

func (bc *bloomChecker) Close() error {
	return nil
}

func NewBloomChecker(dataFile string) (*bloomChecker, error) {
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return nil, os.ErrNotExist
	}
	file, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	filter := bloom.NewWithEstimates(1, 0.01)
	_, err = filter.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return &bloomChecker{filter: filter}, nil
}

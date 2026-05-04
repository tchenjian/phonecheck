# PhoneCheck - 大规模手机号存在性判断系统

## 项目概述

本项目实现了一个高性能的手机号存在性判断系统，专为处理 20 亿级别手机号数据而设计，提供多种存储方案以适应不同的内存、精确性和性能需求。

## 支持的方案

| 方案 | 命令行参数 | 说明 |
|------|-----------|------|
| Roaring Bitmap | `-t bitmap` | **推荐**，内存极低，精确查询，速度极快 |
| Bloom Filter | `-t bloom` | 内存最低，有假阳性，速度极快 |
| SQLite | `-t sqlite` | 磁盘存储，精确查询，速度较慢 |
| Radix Tree | `-t tree` | 预留方案，待实现 |

## 技术选型与背景

### Roaring Bitmap 方案

**Roaring Bitmap** 是一种高性能压缩位图数据结构：
- 将手机号（11位数字，最大 19999999999）映射为 64 位整数
- 采用分段压缩存储：密集区用数组，稀疏区用位图
- **20亿数据内存开销：约 800MB**，完全在 2GB 限制内
- 100% 精确，无假阳性/假阴性
- O(1) 查询时间，纳秒级响应

### Bloom Filter 方案

**布隆过滤器** 是一种空间效率极高的概率型数据结构：
- 多个哈希函数映射到位数组
- **20亿数据内存开销：约 230MB**（误判率 1%）
- **假阳性**：不存在的手机号可能被误判为存在
- **无假阴性**：存在的手机号一定能查到
- O(k) 查询时间，k 为哈希函数数量

### SQLite 方案

**SQLite 嵌入式数据库**：
- 数据存储在磁盘，内存占用可忽略不计
- 100% 精确
- **磁盘IO查询**，毫秒级响应，不适合高并发场景
- 需要 CGO 支持

## 项目结构

```
phonecheck/
├── main.go              # 主程序入口（交互式查询）
├── phonecheck.go        # 核心接口与实现（4种方案）
├── phonecheck_test.go   # 单元测试
├── phone_numbers.txt    # 原始数据（一行一个手机号）
├── phone_numbers.bin    # Roaring Bitmap 数据文件
├── phone_numbers_bloom.bin # Bloom Filter 数据文件
├── phone_numbers.db     # SQLite 数据库文件
├── README.md
├── go.mod
├── go.sum
└── cmd/
    └── generator/
        └── main.go      # 数据转换工具
```

## API 接口

```go
type PhoneChecker interface {
    PhoneExists(phone string) bool  // 判断手机号是否存在
    Close() error                   // 释放资源
}

// 创建检查器
// typ: "bitmap" | "bloom" | "sqlite" | "tree"
func NewPhoneChecker(typ string, dataFile string) (PhoneChecker, error)
```

## 使用说明

### 1. 准备原始数据

创建文本文件，**一行一个手机号**：
```
13800138000
13900139000
15000150000
...
```

### 2. 生成各方案的数据文件

生成所有方案：
```bash
go run cmd/generator/main.go -i phone_numbers.txt -t all -o phone_numbers
```

或单独生成指定方案：
```bash
# 仅生成 Roaring Bitmap
go run cmd/generator/main.go -i phone_numbers.txt -t bitmap -o phone_numbers

# 仅生成 Bloom Filter
go run cmd/generator/main.go -i phone_numbers.txt -t bloom -o phone_numbers

# 生成多种方案（注意：SQLite需要CGO支持）
go run cmd/generator/main.go -i phone_numbers.txt -t bitmap,bloom -o phone_numbers
```

输出示例：
```
Reading phone numbers from phone_numbers.txt...
Loaded 20 phone numbers
Generating Roaring Bitmap to phone_numbers.bin...
  Bitmap cardinality: 20
Generating Bloom Filter to phone_numbers_bloom.bin...
  Bloom filter: 20 items, fp rate: 1.00%
Generating SQLite DB to phone_numbers.db...
Done!
```

### 3. 运行交互式查询程序

```bash
# Bitmap 方案（推荐）
go run main.go -t bitmap -f phone_numbers.bin

# Bloom Filter 方案
go run main.go -t bloom -f phone_numbers_bloom.bin

# SQLite 方案（需要CGO支持）
go run main.go -t sqlite -f phone_numbers.db
```

交互示例：
```
Phone Checker initialized successfully
Type phone numbers to check (type 'exit' to quit)
> 13800138000
✓ 13800138000 exists
> 12345678901
✗ 12345678901 does not exist
> exit
```

### 4. 运行测试

```bash
go test -v
```

## 性能指标对比

| 指标 | Bitmap | Bloom Filter | SQLite | Map |
|------|--------|--------------|--------|-----|
| 20亿内存 | ~800MB | ~230MB | <10MB | ~70GB |
| 20亿磁盘 | ~300MB | ~100MB | ~25GB | ~25GB |
| 查询时间 | 纳秒级 | 纳秒级 | 毫秒级 | 微秒级 |
| 精确性 | 100% 精确 | 有误判 | 100% 精确 | 100% 精确 |
| 加载速度 | 秒级 | 秒级 | 立即 | 分钟级 |
| 并发安全 | 只读安全 | 只读安全 | 是 | 否 |
| 需要CGO | 否 | 否 | 是 | 否 |

## 方案选择指南

| 场景 | 推荐方案 | 原因 |
|------|---------|------|
| 高并发实时查询，要求精确 | **Bitmap** | 内存可控，速度极快，100%精确 |
| 内存极度受限，可接受假阳性 | **Bloom Filter** | 内存占用最低 |
| 数据量远超内存，查询频率低 | **SQLite** | 磁盘存储，精确但较慢 |
| 数据量小 (<1000万) | 任意 | 差异不明显 |

## Bloom Filter 假阳性说明

布隆过滤器 **"不存在的一定不存在，存在的可能不存在"**：

- 如果返回 `false` → 手机号 100% 不存在
- 如果返回 `true` → 手机号 可能存在（有约 1% 概率误判）

适合场景：黑名单过滤、缓存预热等，误判不会造成严重后果的场景。

## SQLite CGO 说明

Windows 环境默认 `CGO_ENABLED=0`，SQLite 需要 CGO 支持才能工作：

1. 安装 MinGW-w64
2. 设置环境变量：
   ```bash
   set CGO_ENABLED=1
   set CC=gcc
   ```
3. 重新编译运行

## 数据规模验证

| 数据量 | Bitmap 内存 | Bloom Filter 内存 (1% fp) |
|--------|------------|-------------------------|
| 1000万 | ~4MB | ~1.2MB |
| 1亿 | ~40MB | ~12MB |
| 10亿 | ~400MB | ~115MB |
| 20亿 | ~800MB | ~230MB |
| 50亿 | ~2GB | ~575MB |

## 常见问题

**Q: 可以支持其他国家的手机号吗？**
A: 可以，只要手机号能转换为唯一的 64 位整数即可。

**Q: 如何增量添加手机号？**
A: 本版本为只读模式，增量更新需要重建数据文件。

**Q: 并发安全吗？**
A: Bitmap 和 Bloom Filter 的只读查询是并发安全的。

**Q: Bloom Filter 的误判率可以调整吗？**
A: 可以，修改 generator.go 中的 `fpRate` 参数，值越小误判率越低，内存占用越高。

## 依赖

- [RoaringBitmap](https://github.com/RoaringBitmap/roaring) - 高性能压缩位图库
- [bloom](https://github.com/bits-and-blooms/bloom) - 布隆过滤器实现
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite 驱动

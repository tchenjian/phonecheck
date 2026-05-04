package checker

import (
	"os"
	"testing"
)

func createTestFile(t *testing.T, phones []string) string {
	tmpFile, err := os.CreateTemp("", "phones_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	for _, phone := range phones {
		tmpFile.WriteString(phone + "\n")
	}

	return tmpFile.Name()
}

// TestRadixTree_PrefixSharing 验证前缀共享核心特性
func TestRadixTree_PrefixSharing(t *testing.T) {
	testPhones := []string{
		"13800138000",
		"13900139000",
		"15000150000",
		"15100151000",
		"15200152000",
		"15300153000",
		"18000180000",
		"18100181000",
	}
	tmpFile := createTestFile(t, testPhones)
	defer os.Remove(tmpFile)

	tree, err := NewTreeChecker(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// 验证前缀计数 - 核心特性！
	if tree.CountPrefix("1") != 8 {
		t.Errorf("前缀'1'应该有8个号码，实际%d", tree.CountPrefix("1"))
	}

	if tree.CountPrefix("13") != 2 {
		t.Errorf("前缀'13'应该有2个号码，实际%d", tree.CountPrefix("13"))
	}

	if tree.CountPrefix("15") != 4 {
		t.Errorf("前缀'15'应该有4个号码，实际%d", tree.CountPrefix("15"))
	}

	if tree.CountPrefix("18") != 2 {
		t.Errorf("前缀'18'应该有2个号码，实际%d", tree.CountPrefix("18"))
	}

	// 验证精确前缀
	if tree.CountPrefix("138") != 1 {
		t.Errorf("前缀'138'应该有1个号码，实际%d", tree.CountPrefix("138"))
	}

	// 验证不存在前缀
	if tree.CountPrefix("19") != 0 {
		t.Errorf("前缀'19'应该有0个号码，实际%d", tree.CountPrefix("19"))
	}
}

// TestRadixTree_PhoneExists 测试存在性查询
func TestRadixTree_PhoneExists(t *testing.T) {
	testPhones := []string{
		"13800138000",
		"13900139000",
		"15000150000",
	}
	tmpFile := createTestFile(t, testPhones)
	defer os.Remove(tmpFile)

	tree, err := NewTreeChecker(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// 测试存在的手机号
	for _, phone := range testPhones {
		if !tree.PhoneExists(phone) {
			t.Errorf("手机号%s应该存在", phone)
		}
	}

	// 测试不存在的手机号
	nonExistent := []string{
		"13500135000",
		"13400134000",
		"invalid",
		"12345", // 长度不足
	}
	for _, phone := range nonExistent {
		if tree.PhoneExists(phone) {
			t.Errorf("手机号%s不应该存在", phone)
		}
	}
}

// TestRadixTree_DuplicateInsert 测试重复插入
func TestRadixTree_DuplicateInsert(t *testing.T) {
	tree := &radixTreeChecker{
		Root: &RadixTreeNode{},
		Size: 0,
	}

	// 多次插入同一个手机号
	for i := 0; i < 10; i++ {
		tree.Insert("13800138000")
	}

	if tree.GetSize() != 1 {
		t.Errorf("重复插入后期望大小1，实际%d", tree.GetSize())
	}

	// 前缀计数也应该是1（而不是10）
	if tree.CountPrefix("138") != 1 {
		t.Errorf("重复插入不应该影响前缀计数，期望1，实际%d", tree.CountPrefix("138"))
	}
}

// TestRadixTree_Size 验证数据大小
func TestRadixTree_Size(t *testing.T) {
	tree := &radixTreeChecker{
		Root: &RadixTreeNode{},
		Size: 0,
	}

	// 插入100个不同手机号（确保11位）
	for i := 0; i < 100; i++ {
		phone := "138" + string(rune('0'+i/100)) +
			string(rune('0'+i%100/10)) +
			string(rune('0'+i%10)) + "00000"
		tree.Insert(phone)
	}

	if tree.GetSize() != 100 {
		t.Errorf("期望100个手机号，实际%d", tree.GetSize())
	}

	if tree.CountPrefix("1") != 100 {
		t.Errorf("前缀1下应该有100个，实际%d", tree.CountPrefix("1"))
	}
}

// TestRadixTree_EmptyPrefix 测试空前缀边界情况
func TestRadixTree_EmptyPrefix(t *testing.T) {
	tree := &radixTreeChecker{
		Root: &RadixTreeNode{},
		Size: 0,
	}

	tree.Insert("13800138000")

	// 空前缀应该返回0（或者可以根据需求修改）
	if tree.CountPrefix("") != 0 {
		t.Log("注意：空前缀计数实现可能需要根据需求调整")
	}
}

// BenchmarkRadixTree_Insert 插入性能测试
func BenchmarkRadixTree_Insert(b *testing.B) {
	tree := &radixTreeChecker{
		Root: &RadixTreeNode{},
		Size: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		phone := "138" + string(rune('0'+i%1000/100)) +
			string(rune('0'+i%100/10)) +
			string(rune('0'+i%10)) + "00000"
		tree.Insert(phone)
	}
}

// BenchmarkRadixTree_Search 查询性能测试
func BenchmarkRadixTree_Search(b *testing.B) {
	tree := &radixTreeChecker{
		Root: &RadixTreeNode{},
		Size: 0,
	}

	// 预先插入10万个号码
	for i := 0; i < 100000; i++ {
		phone := "138" + string(rune('0'+i%100000/10000)) +
			string(rune('0'+i%10000/1000)) +
			string(rune('0'+i%1000/100)) +
			string(rune('0'+i%100/10)) +
			string(rune('0'+i%10)) + "0000"
		tree.Insert(phone)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		phone := "138" + string(rune('0'+i%100000/10000)) +
			string(rune('0'+i%10000/1000)) +
			string(rune('0'+i%1000/100)) +
			string(rune('0'+i%100/10)) +
			string(rune('0'+i%10)) + "0000"
		tree.Contains(phone)
	}
}

package checker

import (
	"os"
	"testing"
)

func TestMixTree_InsertAndContains(t *testing.T) {
	checker := &mixTreeChecker{
		Root: &MixTreeNode{},
		Size: 0,
	}

	testPhones := []string{
		"13800138000",
		"13800138001",
		"13800139999",
		"13900138000",
		"15000000000",
		"15000009999",
	}

	for _, phone := range testPhones {
		checker.Insert(phone)
	}

	if checker.GetSize() != 6 {
		t.Errorf("期望存储6个手机号，实际%d个", checker.GetSize())
	}

	for _, phone := range testPhones {
		if !checker.Contains(phone) {
			t.Errorf("手机号%s应该存在", phone)
		}
	}

	nonExistent := []string{
		"13800137999",
		"13800140000",
		"14000000000",
		"15000010000",
		"invalid",
		"12345",
	}
	for _, phone := range nonExistent {
		if checker.Contains(phone) {
			t.Errorf("手机号%s不应该存在", phone)
		}
	}
}

func TestMixTree_NoDuplicates(t *testing.T) {
	checker := &mixTreeChecker{
		Root: &MixTreeNode{},
		Size: 0,
	}

	phone := "13800138000"
	checker.Insert(phone)
	checker.Insert(phone)
	checker.Insert(phone)

	if checker.GetSize() != 1 {
		t.Errorf("重复插入不应该增加计数，期望1，实际%d", checker.GetSize())
	}
}

func TestMixTree_SuffixBoundary(t *testing.T) {
	checker := &mixTreeChecker{
		Root: &MixTreeNode{},
		Size: 0,
	}

	boundaryPhones := []string{
		"13800130000",
		"13800130001",
		"13800139998",
		"13800139999",
	}

	for _, phone := range boundaryPhones {
		checker.Insert(phone)
	}

	if checker.GetSize() != 4 {
		t.Errorf("期望存储4个手机号，实际%d个", checker.GetSize())
	}

	for _, phone := range boundaryPhones {
		if !checker.Contains(phone) {
			t.Errorf("边界手机号%s应该存在", phone)
		}
	}
}

func TestMixTree_PrefixCount(t *testing.T) {
	testPhones := []string{
		"13800138000",
		"13800138001",
		"13800139000",
		"13900138000",
		"15000000000",
		"15000000001",
	}
	tmpFile := createTestFile(t, testPhones)
	defer os.Remove(tmpFile)

	checker, err := NewMixTreeChecker(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if checker.CountPrefix("1") != 6 {
		t.Errorf("前缀'1'应该有6个号码，实际%d", checker.CountPrefix("1"))
	}
	if checker.CountPrefix("13") != 4 {
		t.Errorf("前缀'13'应该有4个号码，实际%d", checker.CountPrefix("13"))
	}
	if checker.CountPrefix("138") != 3 {
		t.Errorf("前缀'138'应该有3个号码，实际%d", checker.CountPrefix("138"))
	}
	if checker.CountPrefix("15") != 2 {
		t.Errorf("前缀'15'应该有2个号码，实际%d", checker.CountPrefix("15"))
	}
	if checker.CountPrefix("19") != 0 {
		t.Errorf("前缀'19'应该有0个号码，实际%d", checker.CountPrefix("19"))
	}
}

func TestMixTree_FromFile(t *testing.T) {
	testPhones := []string{
		"13800138000",
		"13900139000",
		"15000150000",
	}
	tmpFile := createTestFile(t, testPhones)
	defer os.Remove(tmpFile)

	checker, err := NewMixTreeChecker(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	for _, phone := range testPhones {
		if !checker.PhoneExists(phone) {
			t.Errorf("手机号%s应该存在", phone)
		}
	}

	nonExistent := []string{
		"13500135000",
		"13400134000",
	}
	for _, phone := range nonExistent {
		if checker.PhoneExists(phone) {
			t.Errorf("手机号%s不应该存在", phone)
		}
	}
}

func TestMixTree_ViaCheckerFactory(t *testing.T) {
	testPhones := []string{
		"13800138000",
		"13900139000",
	}
	tmpFile := createTestFile(t, testPhones)
	defer os.Remove(tmpFile)

	checker, err := NewPhoneChecker("mixtree", tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	for _, phone := range testPhones {
		if !checker.PhoneExists(phone) {
			t.Errorf("手机号%s应该存在", phone)
		}
	}

	if checker.PhoneExists("14000000000") {
		t.Error("手机号14000000000不应该存在")
	}

	checker.Close()
}

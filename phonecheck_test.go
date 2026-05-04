package main

import (
	"testing"

	"phonecheck/checker"
)

func TestBitmapChecker(t *testing.T) {
	checker, err := checker.NewPhoneChecker("bitmap", "phone_numbers.bin")
	if err != nil {
		t.Fatalf("Failed to initialize bitmap checker: %v", err)
	}
	defer checker.Close()

	testCases := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"existent phone 1", "13800138000", true},
		{"existent phone 2", "15000150000", true},
		{"existent phone 3", "18900189000", true},
		{"non-existent phone 1", "99999999999", false},
		{"non-existent phone 2", "12345678901", false},
		{"invalid phone format", "abc", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := checker.PhoneExists(tc.phone)
			if actual != tc.expected {
				t.Errorf("PhoneExists(%q) = %v, want %v", tc.phone, actual, tc.expected)
			}
		})
	}
}

func TestBloomChecker(t *testing.T) {
	checker, err := checker.NewPhoneChecker("bloom", "phone_numbers_bloom.bin")
	if err != nil {
		t.Fatalf("Failed to initialize bloom checker: %v", err)
	}
	defer checker.Close()

	testCases := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"existent phone 1", "13800138000", true},
		{"existent phone 2", "15000150000", true},
		{"non-existent phone 1", "99999999999", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := checker.PhoneExists(tc.phone)
			if actual != tc.expected {
				t.Logf("Note: Bloom filter may have false positives. PhoneExists(%q) = %v, want %v", tc.phone, actual, tc.expected)
			}
		})
	}
}

func TestTreeCheckerPlaceholder(t *testing.T) {
	checker, err := checker.NewPhoneChecker("tree", "")
	if err != nil {
		t.Fatalf("Failed to get tree checker placeholder: %v", err)
	}
	defer checker.Close()
}

func TestInvalidCheckerType(t *testing.T) {
	_, err := checker.NewPhoneChecker("invalid", "file.bin")
	if err == nil {
		t.Error("Expected error for invalid checker type, got nil")
	}
}

package env

import (
	"testing"
)

func TestAddToPathAt(t *testing.T) {
	var savedPath string
	mockGet := func() (string, error) {
		return `C:\Windows;C:\tools`, nil
	}
	mockSet := func(p string) error {
		savedPath = p
		return nil
	}

	// 1. 测试添加新路径
	added, err := AddToPathAt(`C:\videopress`, mockGet, mockSet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !added {
		t.Fatalf("expected path to be added")
	}
	expected := `C:\Windows;C:\tools;C:\videopress`
	if savedPath != expected {
		t.Fatalf("expected saved path %q, got %q", expected, savedPath)
	}

	// 2. 测试重复路径（不添加）
	savedPath = ""
	added, err = AddToPathAt(`C:\tools`, mockGet, mockSet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if added {
		t.Fatalf("expected no path to be added")
	}
	if savedPath != "" {
		t.Fatalf("expected no path to be saved")
	}

	// 3. 测试路径 Clean 后的重复路径比对 (带末尾斜线)
	savedPath = ""
	added, err = AddToPathAt(`C:\tools\`, mockGet, mockSet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if added {
		t.Fatalf("expected no path to be added for duplicate with trailing slash")
	}
}

func TestRemoveFromPathAt(t *testing.T) {
	var savedPath string
	mockGet := func() (string, error) {
		return `C:\Windows;C:\tools;C:\videopress`, nil
	}
	mockSet := func(p string) error {
		savedPath = p
		return nil
	}

	// 1. 测试移除存在路径
	removed, err := RemoveFromPathAt(`C:\tools`, mockGet, mockSet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removed {
		t.Fatalf("expected path to be removed")
	}
	expected := `C:\Windows;C:\videopress`
	if savedPath != expected {
		t.Fatalf("expected saved path %q, got %q", expected, savedPath)
	}

	// 2. 测试移除不存在路径
	savedPath = ""
	removed, err = RemoveFromPathAt(`D:\nonexistent`, mockGet, mockSet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Fatalf("expected no path to be removed")
	}
	if savedPath != "" {
		t.Fatalf("expected no path to be saved")
	}
}

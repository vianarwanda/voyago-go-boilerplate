package entity

import (
	"strings"
	"testing"
	"voyago/core-api/internal/modules/product/entity"
)

func TestLocalizedScan_Success(t *testing.T) {
	input := []byte(`{"en-US":"Phone","id-ID":"Telepon"}`)

	var l entity.Localized
	err := l.Scan(input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l["en-US"] != "Phone" {
		t.Fatal("wrong value en-US")
	}

	if l["id-ID"] != "Telepon" {
		t.Fatal("wrong value id-ID")
	}
}

func TestLocalizedScan_Null(t *testing.T) {
	var l entity.Localized

	err := l.Scan(nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l != nil {
		t.Fatal("expected nil map")
	}
}

func TestLocalizedScan_InvalidType(t *testing.T) {
	var l entity.Localized

	err := l.Scan(123) // bukan []byte

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLocalizedValue(t *testing.T) {
	l := entity.Localized{
		"en-US": "Phone",
	}

	val, err := l.Value()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bytes := val.([]byte)

	if !strings.Contains(string(bytes), "Phone") {
		t.Fatal("json missing value")
	}
}

package phone

import (
	"testing"
)

func TestNormalizeCountry(t *testing.T) {
	tests := []struct {
		in   string
		want string
		err  bool
	}{
		{"PH", "PH", false},
		{"ph", "PH", false},
		{"Philippines", "PH", false},
		{"Philippine", "PH", false},
		{"VN", "VN", false},
		{"Vietnam", "VN", false},
		{"XX", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := NormalizeCountry(tt.in)
		if (err != nil) != tt.err {
			t.Fatalf("NormalizeCountry(%q) err=%v wantErr=%v", tt.in, err, tt.err)
		}
		if got != tt.want {
			t.Errorf("NormalizeCountry(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidate_Philippines(t *testing.T) {
	_, err := Validate("PH", "+639171234567")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	canonical, err := Validate("PH", "639171234567")
	if err != nil || canonical != "639171234567" {
		t.Fatalf("got %q err=%v", canonical, err)
	}
	canonical, err = Validate("PH", "09171234567")
	if err != nil || canonical != "639171234567" {
		t.Fatalf("national 09… want 639171234567 got %q err=%v", canonical, err)
	}
	canonical, err = Validate("PH", "0639171234567")
	if err != nil || canonical != "639171234567" {
		t.Fatalf("0 + international remainder got %q err=%v", canonical, err)
	}
	canonical, err = Validate("PH", "9171234567")
	if err != nil || canonical != "639171234567" {
		t.Fatalf("national without 0: got %q err=%v", canonical, err)
	}
	canonical, err = Validate("PH", "+84961234567")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if want := "6384961234567"; canonical != want {
		t.Fatalf("digits normalized with selected country CC: got %q want %q", canonical, want)
	}
}

func TestValidate_TrunkZero_VN_TH_SG(t *testing.T) {
	c, err := Validate("VN", "0912345678")
	if err != nil {
		t.Fatalf("VN: %v", err)
	}
	if want := "84912345678"; c != want {
		t.Fatalf("VN: got %q want %q", c, want)
	}
	c, err = Validate("TH", "0812345678")
	if err != nil {
		t.Fatalf("TH: %v", err)
	}
	if want := "66812345678"; c != want {
		t.Fatalf("TH: got %q want %q", c, want)
	}
	c, err = Validate("SG", "081234567")
	if err != nil {
		t.Fatalf("SG: %v", err)
	}
	if want := "6581234567"; c != want {
		t.Fatalf("SG: got %q want %q", c, want)
	}
}

package account

import "testing"

func TestNormalizeTheme(t *testing.T) {
	if _, err := NormalizeTheme("blue"); err == nil {
		t.Fatal("expected invalid theme")
	}
	got, err := NormalizeTheme("dark")
	if err != nil || got != "dark" {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestClampSlots(t *testing.T) {
	if ClampSlots(3) != 7 || ClampSlots(20) != 14 || ClampSlots(9) != 9 {
		t.Fatal("slot clamp mismatch")
	}
}

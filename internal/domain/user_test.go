package domain

import (
	"testing"
	"time"
)

func TestUserValidate(t *testing.T) {
	if err := (&User{PhoneNumber: "   "}).Validate(); err == nil {
		t.Fatal("expected blank phone number to be rejected")
	}
	if err := (&User{PhoneNumber: "13800000000"}).Validate(); err != nil {
		t.Fatalf("expected valid phone number, got %v", err)
	}
}

func TestUserStatusAndExpiry(t *testing.T) {
	u := User{Status: UserStatusBlocked, ValidTime: time.Now().Add(-time.Minute)}
	if !u.IsBlocked() || !u.IsExpired() {
		t.Fatal("expected blocked and expired user")
	}
}

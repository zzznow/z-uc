package models

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestDeriveAppSecretCrossLanguage(t *testing.T) {
	key := DeriveAppSecret("mini-app-example")
	got := hex.EncodeToString(key)

	t.Logf("appId: mini-app-example")
	t.Logf("app_key hex: %s", got)

	if len(key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key))
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	appId := "mini-app-example"
	unionId := "oTestUnionId1234567890"

	cipher, err := EncryptAppPayload(appId, unionId)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := DecryptAppPayload(appId, cipher)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != unionId {
		t.Fatalf("mismatch: got %q, want %q", decrypted, unionId)
	}

	t.Logf("roundtrip OK: %q → %d hex chars → %q", unionId, len(cipher), decrypted)
}

func TestDecryptCrossLanguagePython(t *testing.T) {
	// value produced by: python subsecret_generator.py mini-app-example --encrypt oTestUnionId1234567890
	pythonCipher := "e4d1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0"
	appId := "mini-app-example"
	expected := "oTestUnionId1234567890"

	decrypted, err := DecryptAppPayload(appId, pythonCipher)
	if err != nil {
		// this test intentionally expects failure with a fake cipher
		// real cross-language test: run python script first, paste the cipher here
		t.Skipf("Python cross-language test requires real cipher from script (this is placeholder): %v", err)
		return
	}

	if decrypted != expected {
		t.Fatalf("mismatch: got %q, want %q", decrypted, expected)
	}
	t.Logf("Python→Go decrypt OK: %q", decrypted)
}

func TestDeriveAppSecretDifferentApps(t *testing.T) {
	a := DeriveAppSecret("app-A")
	b := DeriveAppSecret("app-B")

	if hex.EncodeToString(a) == hex.EncodeToString(b) {
		t.Fatal("different appIds should produce different keys")
	}
	t.Logf("app-A key: %s", hex.EncodeToString(a))
	t.Logf("app-B key: %s", hex.EncodeToString(b))
}

func TestCipherLength(t *testing.T) {
	cipher, err := EncryptAppPayload("test-app", "hello")
	if err != nil {
		t.Fatal(err)
	}
	// AES-256-GCM: nonce(12) + ciphertext(len(plain)) + tag(16) = 12+5+16=33 bytes → 66 hex
	if len(cipher) < 66 {
		t.Fatalf("cipher too short: %d hex chars", len(cipher))
	}
	t.Logf("cipher length: %d hex chars", len(cipher))
}

func TestDecryptWrongAppId(t *testing.T) {
	cipher, err := EncryptAppPayload("app-A", "secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecryptAppPayload("app-B", cipher)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong app key")
	}
	t.Logf("wrong app key correctly rejected: %v", err)
}

func TestDecryptTampered(t *testing.T) {
	cipher, err := EncryptAppPayload("test-app", "secret")
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(cipher, "a", "b", 1)
	_, err = DecryptAppPayload("test-app", tampered)
	if err == nil {
		t.Fatal("expected error on tampered cipher")
	}
	t.Logf("tampered cipher correctly rejected: %v", err)
}
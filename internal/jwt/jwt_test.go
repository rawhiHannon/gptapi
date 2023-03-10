package jwt

import (
	"errors"
	"gptapi/internal/uniqid"
	"log"
	"testing"
)

type mockCacheManager struct {
	dataMap map[string]map[string]interface{}
}

func (m *mockCacheManager) HSet(key string, field string, value interface{}) error {
	if m.dataMap == nil {
		m.dataMap = make(map[string]map[string]interface{})
	}
	if m.dataMap[key] == nil {
		m.dataMap[key] = make(map[string]interface{})
	}
	m.dataMap[key][field] = value
	return nil
}

func (m *mockCacheManager) HGet(key string, field string) (string, error) {
	if m.dataMap == nil {
		m.dataMap = make(map[string]map[string]interface{})
	}
	if hash, exists := m.dataMap[key]; exists {
		if val, hasField := hash[field]; hasField {
			return val.(string), nil
		}
	}
	return "", errors.New("not found")
}

func TestJWT(t *testing.T) {
	cache := &mockCacheManager{}
	jwt := New(cache, "abcde")

	// Test creating a token
	data := map[string]interface{}{
		"name": "John",
		"age":  30.5,
	}
	tokenPayload, err := jwt.CreateToken("test", uniqid.NextId(), data)
	if err != nil {
		t.Fatalf("Unexpected error creating token: %v", err)
	}
	if tokenPayload.AccessId == 0 {
		t.Errorf("Expected non-zero AccessId in token payload")
	}
	if tokenPayload.Token == "" {
		t.Errorf("Expected non-empty Token in token payload")
	}
	if !tokenPayload.Exists {
		t.Errorf("Expected Exists to be true in token payload")
	}
	if tokenPayload.Expire == 0 {
		t.Errorf("Expected non-zero Expire in token payload")
	}
	if tokenPayload.Device != "" {
		t.Errorf("Expected empty Device in token payload")
	}
	if len(tokenPayload.Data) != len(data) {
		t.Errorf("Expected Data to have %d items, but got %d", len(data), len(tokenPayload.Data))
	}
	for key, value := range data {
		if tokenPayload.Data[key] != value {
			t.Errorf("Expected Data[%s] to be %s, but got %s", key, value, tokenPayload.Data[key])
		}
	}
	if _, err := cache.HGet("test:metadata", "jwt"); err != nil {
		t.Errorf("Expected token to be stored in cache, but got error: %v", err)
	}

	// Test validating a token
	existingTokenStr := tokenPayload.Token
	log.Println(existingTokenStr)
	tokenPayload, err = jwt.ValidateToken(existingTokenStr)
	if err != nil {
		t.Fatalf("Unexpected error validating token: %v", err)
	}
	if tokenPayload.Token != existingTokenStr {
		t.Errorf("Expected token string to match, but got %s and %s", tokenPayload.Token, existingTokenStr)
	}
	if len(tokenPayload.Data) != len(data) {
		t.Errorf("Expected Data to have %d items, but got %d", len(data), len(tokenPayload.Data))
	}
	for key, value := range data {
		if tokenPayload.Data[key] != value {
			t.Errorf("Expected Data[%s] to be %s, but got %s", key, value, tokenPayload.Data[key])
		}
	}
}

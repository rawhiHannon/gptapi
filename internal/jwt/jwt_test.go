package jwt

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
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

func TestNewJWT(t *testing.T) {
	cache := &mockCacheManager{}
	jwt := New(cache)
	if jwt.cache != cache {
		t.Error("Failed to initialize cache")
	}
}

func TestVerifyToken(t *testing.T) {
	os.Setenv("SERVER_SECRET", "secret")
	_jwt := &JWT{}
	token, err := _jwt.verifyToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NVdWlkIjoiOTBiNmJjYmItYjEwZi00ZWE2LWJkZmItNWI4OGI4YjU5ZDZjIiwiZXhwaXJlIjozMjI4NTUzNTM0LCJpZGVudGlmaWVyIjoidGVzdHVzZXIiLCJwZXJtaXNzaW9uIjoidGVzdF9wZXJtaXNzaW9uIn0.ddPkj9euDbkSURdkoJPFdqTxIyQsvdQQ_EUb8jIe1bs")
	if err != nil {
		t.Error(err)
	}
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		t.Error("Unexpected signing method")
	}
}

func TestExtractTokenMetadata(t *testing.T) {
	os.Setenv("SERVER_SECRET", "secret")
	jwt := &JWT{}
	payload, err := jwt.extractTokenMetadata("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NVdWlkIjoiOTBiNmJjYmItYjEwZi00ZWE2LWJkZmItNWI4OGI4YjU5ZDZjIiwiZXhwaXJlIjozMjI4NTUzNTM0LCJpZGVudGlmaWVyIjoidGVzdHVzZXIiLCJwZXJtaXNzaW9uIjoidGVzdF9wZXJtaXNzaW9uIn0.ddPkj9euDbkSURdkoJPFdqTxIyQsvdQQ_EUb8jIe1bs")
	if err != nil {
		t.Error(err)
	}
	if payload.Identifier != "testuser" {
		t.Error("Invalid identifier")
	}
	if payload.Permission != "test_permission" {
		t.Error("Invalid permission")
	}
}

func TestValidateToken(t *testing.T) {
	os.Setenv("SERVER_SECRET", "secret")
	cache := &mockCacheManager{}
	jwt := New(cache)
	token, err := jwt.CreateToken("test_user", "test_permission")
	payload, err := jwt.ValidateToken(token.Token)
	if err != nil {
		t.Error(err)
	}
	if payload.Identifier != "test_user" {
		t.Error("Invalid identifier")
	}
	if payload.Permission != "test_permission" {
		t.Error("Invalid permission")
	}
}

func TestCreateToken(t *testing.T) {
	os.Setenv("SERVER_SECRET", "secret")
	cache := &mockCacheManager{}
	jwt := New(cache)

	token, err := jwt.CreateToken("testuser", "test_permission")
	if err != nil {
		t.Error(err)
	}

	if token.AccessUuid == "" {
		t.Error("Failed to generate accessUuid", token.AccessUuid)
	}
	if token.Permission != "test_permission" {
		t.Error("Invalid permission", token.Permission)
	}

	if token.Expire <= time.Now().Unix() {
		t.Error("Invalid expiration time", token.Expire)
	}

	token, err = jwt.CreateToken("testuser", "test_permission")
	if err != nil {
		t.Error(err)
	}
	if !token.Exists {
		log.Println(token)
		t.Error("Failed to find existing token")
	}
}

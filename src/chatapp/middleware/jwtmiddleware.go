package middleware

import (
	"errors"
	"net/http"

	// "os"
	"log"
	"strings"
	// "github.com/dgrijalva/jwt-go"
)

type UserManager interface {
	ValidateToken(string) error
	ValidateAdminToken(string) error
}

type JWTMiddleware struct {
	userManager UserManager
}

func NewJWTMiddleware(userManager UserManager) *JWTMiddleware {
	jwtm := &JWTMiddleware{}
	jwtm.userManager = userManager
	return jwtm
}

func (this *JWTMiddleware) extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func (this *JWTMiddleware) PreProcessor(w http.ResponseWriter, r *http.Request) error {
	var tokenStr string
	if tokenStr = this.extractToken(r); tokenStr == "" {
		return errors.New("no authorization header")
	}
	if err := this.userManager.ValidateToken(tokenStr); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (this *JWTMiddleware) PostProcessor(resultChan map[string]chan struct{}, resultStr string) error {
	return nil
}

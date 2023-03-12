package jwt

import (
	"errors"
	"gptapi/pkg/models"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Identifier string
	Password   string
	Data       map[string]string
}

type JWT struct {
	cache  models.CacheManager
	secret string
}

type TokenPayload struct {
	AccessId   uint64
	Identifier string
	Token      string
	Exists     bool
	Expire     int64
	Data       map[string]interface{}
}

func New(cache models.CacheManager, secret string) *JWT {
	instance := &JWT{}
	instance.cache = cache
	return instance
}

func (j *JWT) verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method") //, token.Header["alg"]
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (j *JWT) extractTokenMetadata(tokenString string) (*TokenPayload, error) {
	token, err := j.verifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	var ok bool
	var accessId float64
	var identifier string
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessId, ok = claims["accessId"].(float64)
		if !ok {
			return nil, errors.New("accessId token not valid")
		}
		identifier, ok = claims["identifier"].(string)
		if !ok {
			return nil, errors.New("identifier token not valid")
		}
		dataMap, ok := claims["data"].(map[string]interface{})
		if !ok {
			return nil, errors.New("data token not valid")
		}
		return &TokenPayload{
			AccessId:   uint64(accessId),
			Identifier: identifier,
			Data:       dataMap,
			Token:      tokenString,
		}, nil
	}
	return nil, err
}

func (j *JWT) ValidateToken(tokenStr string) (*TokenPayload, error) {
	var tokenPayload *TokenPayload
	var jwtToken string
	var err error
	if tokenPayload, err = j.extractTokenMetadata(tokenStr); err != nil {
		return nil, err
	}
	if jwtToken, err = j.cache.HGet(tokenPayload.Identifier+":metadata", "jwt"); jwtToken != tokenStr || err != nil {
		return nil, errors.New("token is not valid")
	}
	return tokenPayload, nil
}

func (j *JWT) CreateToken(identifier string, accessId uint64, data map[string]interface{}) (*TokenPayload, error) {
	tokenPayload := &TokenPayload{}
	storedJwtToken, _ := j.cache.HGet(identifier+":metadata", "jwt")
	if storedJwtToken != "" {
		tokenPayload.Token = storedJwtToken
		tokenPayload.Exists = true
		return tokenPayload, nil
	}
	expire := time.Now().Add(time.Hour * time.Duration(24)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accessId":   accessId,
		"identifier": identifier,
		"data":       data,
		"exp":        expire,
	})
	jwtToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return nil, err
	}
	if err := j.cache.HSet(identifier+":metadata", "jwt", jwtToken); err != nil {
		return nil, err
	}
	return &TokenPayload{
		AccessId:   accessId,
		Identifier: identifier,
		Token:      jwtToken,
		Exists:     true,
		Expire:     expire,
		Data:       data,
	}, nil
}

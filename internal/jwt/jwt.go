package jwt

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
)

type CacheManager interface {
	HSet(string, string, interface{}) error
	HGet(string, string) (string, error)
}

type User struct {
	Identifier string
	Password   string
	Permission string
}

type JWT struct {
	cache CacheManager
}

type TokenPayload struct {
	AccessUuid string
	Identifier string
	Token      string
	Exists     bool
	Expire     int64
	Permission string
	Device     string
}

func New(cache CacheManager) *JWT {
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
		return []byte(os.Getenv("SERVER_SECRET")), nil
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
	var accessUuid string
	var identifier string
	var permission string
	var device string
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok = claims["accessUuid"].(string)
		if !ok {
			return nil, errors.New("token not valid")
		}
		identifier, ok = claims["identifier"].(string)
		if !ok {
			return nil, errors.New("token not valid")
		}
		permission, ok = claims["permission"].(string)
		if !ok {
			return nil, errors.New("token not valid")
		}
		return &TokenPayload{
			AccessUuid: accessUuid,
			Permission: permission,
			Identifier: identifier,
			Device:     device,
		}, nil
	}
	return nil, err
}

func (j *JWT) ValidateToken(tokenStr string) (*TokenPayload, error) {
	var tokenPayload *TokenPayload
	var jwtToken string
	var err error
	if tokenPayload, err = j.extractTokenMetadata(tokenStr); err != nil {
		return nil, errors.New("token is not valid, extractin, metadata error")
	}
	if jwtToken, err = j.cache.HGet(tokenPayload.Identifier+":metadata", "jwt"); jwtToken != tokenStr || err != nil {
		return nil, errors.New("token is not valid")
	}
	return tokenPayload, nil
}

func (j *JWT) CreateToken(identifier string, permission string) (*TokenPayload, error) {
	tokenPayload := &TokenPayload{}
	storedJwtToken, _ := j.cache.HGet(identifier+":metadata", "jwt")
	if storedJwtToken != "" {
		tokenPayload.Token = storedJwtToken
		tokenPayload.Exists = true
		log.Println(tokenPayload)
		return tokenPayload, nil
	}
	tokenPayload.Expire = time.Now().Add(time.Hour * 24 * 360 * 50).Unix()
	tokenPayload.AccessUuid = uuid.NewV4().String()
	tokenPayload.Permission = permission
	var err error
	var accessToken string
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["accessUuid"] = tokenPayload.AccessUuid
	atClaims["permission"] = tokenPayload.Permission
	atClaims["identifier"] = identifier
	atClaims["expire"] = tokenPayload.Expire
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err = at.SignedString([]byte(os.Getenv("SERVER_SECRET")))
	if err != nil {
		return nil, err
	}
	tokenPayload.Token = accessToken
	//cache token in redis
	if j.cache != nil {
		j.cache.HSet(identifier+":metadata", "jwt", accessToken)
	}
	return tokenPayload, nil
}

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// new type to differentiate between different token types
type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy-access"
	// TokenTypeRefresh -
	TokenTypeRefresh TokenType = "chirpy-refresh"
)

var ErrNoAuthHeaderIncluded = errors.New("not auth header included in request")

// HashPassword - generate hash using bcrypt
func HashPassword(password string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

// CheckPasswordHash - compare hash of passwords
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// MakeJWT - creates new jwt based on userID & tokenSecret
func MakeJWT(userID int, tokenSecret string, expiresIn time.Duration, tokenType TokenType) (string, error) {
	// key written in "jwt.env" file
	signingKey := []byte(tokenSecret)
	// create new token using jwt library, specifying signing method and claims
	// registered claims are standardized values to external libraries
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		// which application issued the JWT?
		Issuer: string(tokenType),
		// when was the JWT issued?
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		// until what date/time can the JWT be accepted?
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		// who is the user subject of the JWT?
		Subject: fmt.Sprintf("%d", userID),
	})
	// sign the token with secret key
	return token.SignedString(signingKey)
}

// RefreshToken - generate a new token based on the refresh token
func RefreshToken(tokenString, tokenSecret string) (string, error) {
	// use claims of received jwt to check if it's same as local data
	claimsStruct := jwt.RegisteredClaims{}
	// retrieve token using library function with appropriate params
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}
	// get the subject field value from request auth header
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	// get issuer field value from request auth header
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return "", err
	}
	// check if issuer is for refresh token and not access token
	if issuer != string(TokenTypeRefresh) {
		return "", errors.New("invalid issuer")
	}
	// prepare to generate new jwt with same userID
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		return "", err
	}
	// generate a new jwt based on the same userID and tokenSecret, but at different timing
	newToken, err := MakeJWT(
		userID,
		tokenSecret,
		time.Hour,
		TokenTypeAccess,
	)
	if err != nil {
		return "", err
	}
	// all checks passed, return the newly generate token
	return newToken, nil
}

// ValidateJWT - check if jwt token satisfies proper formatting, return user id if valid
func ValidateJWT(tokenString, tokenSecret string) (string, error) {
	// use claims of received jwt to check if it's same as local data
	claimsStruct := jwt.RegisteredClaims{}
	// retrieve token using library function with appropriate params
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}
	// get the subject field value from request auth header
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	// get issuer field value from request auth header
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return "", err
	}
	// check if issuer is for access token and not refresh token
	if issuer != string(TokenTypeAccess) {
		return "", errors.New("invalid issuer")
	}
	// all checks passed, return embedded user ID
	return userIDString, nil
}

// GetBearerToken - returns the token within request header
func GetBearerToken(headers http.Header) (string, error) {
	// get header in field of "authorization"
	authHeader := headers.Get("Authorization")
	// throw error if missing such field in headers
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	// check if auth header is in correct format, "Bearer: <Token>"
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}
	// all checks passed, return the token
	return splitAuth[1], nil
}

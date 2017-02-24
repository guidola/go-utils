package security

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"crypto/rsa"
)


type (
	// JWTConfig defines the config for JWT auth middleware.
	JWTConfig struct {
		// Signing key to validate token.
		// Required.
		//in this case it is a rsa private key therefore  not being
		SigningKey *rsa.PublicKey `json:"signing_key"`

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string `json:"signing_method"`

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string `json:"context_key"`

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		TokenLookup string `json:"token_lookup"`
	}

	jwtExtractor func(echo.Context) (string, error)
)

const (
	bearer = "jwt$"
)

// Algorithims
const (
	AlgorithmRS512 = "RS512"
)

var (
	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		SigningMethod: AlgorithmRS512,
		ContextKey:    "user_id",
		TokenLookup:   "header:" + echo.HeaderAuthorization,
	}
)

// JWT returns a JSON Web Token (JWT) auth middleware.
// The key parameter with default configuration has to be a rsa_public_key
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it sends "401 - Unauthorized" response.
// For empty or invalid `Authorization` header, it sends "400 - Bad Request".
//
// See: https://jwt.io/introduction
func RSA_JWT(key *rsa.PublicKey) echo.MiddlewareFunc {
	DefaultJWTConfig.SigningKey = key
	c := DefaultJWTConfig
	return jwtWithConfig(c)
}

// JWTWithConfig returns a JWT auth middleware from config.
// See: `JWT()`.
func jwtWithConfig(config JWTConfig) echo.MiddlewareFunc {
	// Defaults
	if config.SigningKey == nil {
		panic("jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := JwtFromHeader(parts[1])

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if NonAuthenticationRequired(c.Request().URL.Path) {
				//if its an open query point bypass jwp auth
				return next(c)
			}
			auth, err := extractor(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			token, err := jwt.Parse(auth, func(t *jwt.Token) (interface{}, error) {
				// Check the signing method
				if t.Method.Alg() != config.SigningMethod {
					return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return config.SigningKey, nil

			})
			if err == nil && token.Valid && IsJWTValid(*token) {
				// Store user information from token into context.
				c.Set(config.ContextKey, token.Claims.(jwt.MapClaims)["sub"])
				return next(c)
			}

			return echo.ErrUnauthorized
		}
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the provided
// request header.
func JwtFromHeader(header string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		l := len(bearer)
		if len(auth) > l && auth[:l] == bearer {
			return auth[l:], nil
		}
		return "", errors.New("empty or invalid jwt in authorization header")
	}
}

//not allowing authentication in query parameter otfp
// jwtFromQuery returns a `jwtExtractor` that extracts token from the provided query
// parameter.
/*func jwtFromQuery(param string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", errors.New("empty jwt in query param")
		}
		return token, nil
	}
}*/

package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"math/rand"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
	log "github.com/sirupsen/logrus"
)

var expirationDelta = time.Hour * 12

type AuthType int

const (
	AuthTypeUnknown AuthType = iota
	AuthTypeBasic
	AuthTypeToken
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	// JWT
	Token    string
	AuthType AuthType `json:"-"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"-"`
}

type Authenticator interface {
	// indicates whether authentication is enabled
	Enabled() bool
	Authenticate(req *AuthRequest) (*AuthResponse, error)
	GenerateToken(u User) (*AuthResponse, error)
}

func New(opts *Opts, issuers map[string]Issuer) *DefaultAuthenticator {
	if len(opts.Secret) == 0 {
		opts.Secret = []byte(randStringRunes(23))
	}

	return &DefaultAuthenticator{
		opts:    opts,
		secret:  opts.Secret,
		issuers: issuers,
	}
}

type Opts struct {
	// Basic auth
	Username string
	Password string

	// Secret used to sign JWT tokens
	Secret []byte
}

type Issuer struct {
	Jwks          string
	Name          string
	UsernameClaim string
}

type DefaultAuthenticator struct {
	issuers map[string]Issuer
	opts    *Opts

	secret []byte
}

var (
	ErrUnauthorized = errors.New("unauthorized")
)

func (a *DefaultAuthenticator) Enabled() bool {
	return true
}

func (a *DefaultAuthenticator) Authenticate(req *AuthRequest) (*AuthResponse, error) {

	switch req.AuthType {
	case AuthTypeToken:
		user, err := a.parseToken(req.Token)
		if err != nil {
			return nil, err
		}
		return &AuthResponse{
			User:  *user,
			Token: req.Token,
		}, nil
	case AuthTypeBasic:
		// ok
	default:
		return nil, fmt.Errorf("unknown auth type")
	}

	// TODO: set as env var
	if a.opts.Username == "" && a.opts.Password == "" {
		// if basic auth not set - authenticating as guest
		return a.GenerateToken(User{Username: "guest"})
	}

	if req.Username != a.opts.Username || req.Password != a.opts.Password {
		return nil, ErrUnauthorized
	}

	return a.GenerateToken(User{Username: req.Username})
}

type User struct {
	Username string
	Roles    []string
}

func (a *DefaultAuthenticator) GenerateToken(u User) (*AuthResponse, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"username": "admin",
		"exp":      time.Now().Add(expirationDelta).Unix(),
		"iat":      time.Now().Unix(),
		"iss":      "quilla",
		"roles":    []string{"admin"},
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token, error: %s, s: %s", err, string(a.secret))
	}

	return &AuthResponse{
		Token: tokenString,
	}, nil
}

func parseThirdPartyToken(tokenString string, jwks string) (*jwt.Token, error) {
	keyset, err := jwk.Fetch(context.Background(), jwks)
	if err != nil {
		return nil, fmt.Errorf("error fetching remote jwks")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		keys, ok := keyset.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("key %v not found", kid)
		}

		publicKey := &rsa.PublicKey{}
		err := keys.Raw(publicKey)
		if err != nil {
			return nil, fmt.Errorf("could not parse public key")
		}

		return publicKey, nil
	})

	return token, nil
}

func (a *DefaultAuthenticator) parseToken(tokenString string) (*User, error) {
	var issuer Issuer
	unverifiedToken, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("unable to parse token")
	}

	if claims, ok := unverifiedToken.Claims.(jwt.MapClaims); ok {
		if claims["iss"] != nil {
			issuer, ok = a.issuers[claims["iss"].(string)]
			if !ok {
				return nil, fmt.Errorf("issuer doesnt exist")
			}
		} else {
			return nil, fmt.Errorf("issuer is not valid")
		}
	}

	var token *jwt.Token
	if issuer.Name == "Quilla" {
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return a.secret, nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		token, err = parseThirdPartyToken(tokenString, issuer.Jwks)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &User{}
		user.Username = parseString(claims, issuer.UsernameClaim)
		if user.Username == "" {
			log.WithFields(log.Fields{
				"token": tokenString,
				"error": "token is missing account username field",
			}).Warn("authenticator: malformed token")
			return nil, fmt.Errorf("malformed token")
		}

		user.Roles = append(user.Roles, parseGroups(claims, "groups")...)
		user.Roles = append(user.Roles, parseGroups(claims, "roles")...)
		// returning
		return user, nil

	}
	return nil, fmt.Errorf("invalid token")

}

func parseGroups(meta map[string]interface{}, key string) []string {
	val, ok := meta[key]
	if !ok {
		return []string{}
	}

	var arr []string
	for _, s := range val.([]interface{}) {
		arr = append(arr, s.(string))
	}

	return arr
}

func parseString(meta map[string]interface{}, key string) string {
	val, ok := meta[key]
	if !ok {
		return ""
	}

	s, ok := val.(string)
	if ok {
		return s
	}

	return ""
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

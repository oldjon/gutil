package gjwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type errorHandler func(w http.ResponseWriter, r *http.Request, err string)

type Options struct {
	KeyGetter     jwt.Keyfunc
	NewClaimsFunc func() jwt.Claims
	SigningMethod jwt.SigningMethod
	ErrorHandler  errorHandler
}

func OnError(w http.ResponseWriter, r *http.Request, err string) {
	http.Error(w, err, http.StatusUnauthorized)
}

type Middleware struct {
	Options Options
}

func New(opts Options) *Middleware {
	if opts.ErrorHandler == nil {
		opts.ErrorHandler = OnError
	}
	if opts.NewClaimsFunc == nil {
		opts.NewClaimsFunc = func() jwt.Claims {
			return jwt.MapClaims{}
		}
	}
	return &Middleware{opts}
}

func (m *Middleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nr, err := m.CheckJWT(w, r)

		// CheckJWT will write http status if there is error
		if err != nil {
			return
		}

		h.ServeHTTP(w, nr)
	})
}

func (m *Middleware) CheckJWT(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	token, err := FromAuthHeader(r)
	if err != nil {
		m.Options.ErrorHandler(w, r, err.Error())
		return nil, err
	}

	claims := m.Options.NewClaimsFunc()
	pToken, err := jwt.ParseWithClaims(token, claims, m.Options.KeyGetter)
	if err != nil {
		m.Options.ErrorHandler(w, r, err.Error())
		return nil, err
	}

	if m.Options.SigningMethod != nil {
		ealg := m.Options.SigningMethod.Alg()
		galg := pToken.Header["alg"]
		if ealg != galg {
			msg := fmt.Sprintf("expected %s signing method but token specificed %s", ealg, galg)
			m.Options.ErrorHandler(w, r, msg)
			return r, errors.New(msg)
		}
	}

	if !pToken.Valid {
		err := errors.New("invalid token")
		m.Options.ErrorHandler(w, r, err.Error())
		return nil, err
	}

	nr := r.WithContext(WithClaims(r.Context(), pToken.Claims))

	return nr, nil
}

func FromAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("Authorization header must be Bearer {token}")
	}

	return parts[1], nil
}

func (m *Middleware) CheckJWTTCP(token string) (jwt.Claims, error) {
	claims := m.Options.NewClaimsFunc()
	pToken, err := jwt.ParseWithClaims(token, claims, m.Options.KeyGetter)
	if err != nil {
		return nil, err
	}

	if m.Options.SigningMethod != nil {
		eAlg := m.Options.SigningMethod.Alg()
		gAlg := pToken.Header["alg"]
		if eAlg != gAlg {
			msg := fmt.Sprintf("expected %s signing method but token specificed %s", eAlg, gAlg)
			return nil, errors.New(msg)
		}
	}

	if !pToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

type key string

const claimsKey key = "jwt_claims"

func FromContext(ctx context.Context) (jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	return claims, ok
}

func WithClaims(ctx context.Context, claims jwt.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

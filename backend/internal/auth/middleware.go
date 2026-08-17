package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/jwks"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/gin-gonic/gin"
)

// JWTChecker is the small portion of the Auth0 middleware used by Gin. Keeping
// it injectable lets tests validate routing without contacting a real JWKS URL.
type JWTChecker interface {
	CheckJWT(http.Handler) http.Handler
}

func New(domain, audience string) (gin.HandlerFunc, error) {
	domain = strings.TrimSpace(domain)
	audience = strings.TrimSpace(audience)
	if domain == "" {
		return nil, errors.New("AUTH0_DOMAIN is required")
	}
	if audience == "" {
		return nil, errors.New("AUTH0_AUDIENCE is required")
	}

	issuerURL, err := url.Parse("https://" + strings.TrimSuffix(domain, "/") + "/")
	if err != nil || issuerURL.Host == "" {
		return nil, fmt.Errorf("invalid AUTH0_DOMAIN")
	}

	provider, err := jwks.NewCachingProvider(jwks.WithIssuerURL(issuerURL))
	if err != nil {
		return nil, fmt.Errorf("create Auth0 JWKS provider: %w", err)
	}

	jwtValidator, err := validator.New(
		validator.WithKeyFunc(provider.KeyFunc),
		validator.WithAlgorithm(validator.RS256),
		validator.WithIssuer(issuerURL.String()),
		validator.WithAudience(audience),
	)
	if err != nil {
		return nil, fmt.Errorf("create Auth0 JWT validator: %w", err)
	}

	checker, err := jwtmiddleware.New(
		jwtmiddleware.WithValidator(jwtValidator),
		jwtmiddleware.WithErrorHandler(unauthorized),
		// Browser preflight contains no credentials and is handled by CORS.
		jwtmiddleware.WithValidateOnOptions(false),
	)
	if err != nil {
		return nil, fmt.Errorf("create Auth0 JWT middleware: %w", err)
	}

	return GinMiddleware(checker), nil
}

func GinMiddleware(checker JWTChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		passed := false
		next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			passed = true
			c.Request = r
		})

		checker.CheckJWT(next).ServeHTTP(c.Writer, c.Request)
		if !passed {
			c.Abort()
			return
		}
		c.Next()
	}
}

func unauthorized(w http.ResponseWriter, _ *http.Request, _ error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}

func Subject(r *http.Request) (string, error) {
	claims, err := jwtmiddleware.GetClaims[*validator.ValidatedClaims](r.Context())
	if err != nil || claims == nil || claims.RegisteredClaims.Subject == "" {
		return "", errors.New("validated subject is unavailable")
	}
	return claims.RegisteredClaims.Subject, nil
}

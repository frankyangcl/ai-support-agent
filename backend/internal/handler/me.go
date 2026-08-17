package handler

import (
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	claims, err := jwtmiddleware.GetClaims[*validator.ValidatedClaims](c.Request.Context())
	if err != nil || claims == nil || claims.RegisteredClaims.Subject == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sub": claims.RegisteredClaims.Subject})
}

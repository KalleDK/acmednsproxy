package httphandlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v4/challenge"
)

type messageRaw struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"keyAuth"`
}

type ProviderBackend interface {
	Get(domain string) (challenge.Provider, error)
}

func getBasicAuth(c *gin.Context) (user string, pass string, err error) {

	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Basic ") {
		return "", "", errors.New("missing auth")
	}

	decodedAuthValue, err := base64.StdEncoding.DecodeString(h[6:])
	if err != nil {
		return "", "", err
	}

	parts := bytes.SplitN(decodedAuthValue, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", err
	}

	return string(parts[0]), string(parts[1]), nil
}

func verifyPermission(a auth.UserAuthenticator) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, pass, _ := getBasicAuth(c)

		var json messageRaw
		if err := c.ShouldBindJSON(&json); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := a.VerifyPermissions(user, pass, json.Domain); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("message", json)
	}
}

func presentHandler(provider challenge.Provider) func(c *gin.Context) {
	return func(c *gin.Context) {
		json := c.MustGet("message").(messageRaw)

		if err := provider.Present(json.Domain, json.Token, json.KeyAuth); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func cleanuptHandler(provider challenge.Provider) func(c *gin.Context) {
	return func(c *gin.Context) {
		json := c.MustGet("message").(messageRaw)

		if err := provider.CleanUp(json.Domain, json.Token, json.KeyAuth); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func NewHandler(a auth.UserAuthenticator, p challenge.Provider) (handler http.Handler, err error) {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/present", verifyPermission(a), presentHandler(p))
	router.POST("/cleanup", verifyPermission(a), cleanuptHandler(p))

	return router, nil
}

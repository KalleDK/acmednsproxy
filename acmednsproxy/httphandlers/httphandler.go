package httphandlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type messageRaw struct {
	Domain  string
	Token   string
	KeyAuth string
}

type messageDefault struct {
	FQDN  string
	Value string
}

type combinedMessage struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"keyAuth"`
	FQDN    string `json:"fqdn"`
	Value   string `json:"value"`
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

		var json combinedMessage
		if err := c.ShouldBindJSON(&json); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.Domain == "" && json.FQDN == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no domain"})
			return
		}

		if json.Domain != "" {
			if err := a.VerifyPermissions(user, pass, json.Domain); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			c.Set("message", messageRaw{
				Domain:  json.Domain,
				Token:   json.Token,
				KeyAuth: json.KeyAuth,
			})
			return
		}

		if err := a.VerifyPermissions(user, pass, dns01.UnFqdn(json.FQDN)); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("message", messageDefault{
			FQDN:  json.FQDN,
			Value: json.Value,
		})
	}
}

func presentHandler(provider providers.ProviderSolved) func(c *gin.Context) {
	return func(c *gin.Context) {
		jsonraw := c.MustGet("message")

		switch json := jsonraw.(type) {
		case messageRaw:

			if err := providers.Present(provider, json.Domain, json.Token, json.KeyAuth); err != nil {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok", "value": json.Token})
		case messageDefault:
			if err := provider.CreateRecord(json.FQDN, json.Value); err != nil {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok", "value": json.Value})

		}

	}
}

func cleanuptHandler(provider providers.ProviderSolved) func(c *gin.Context) {
	return func(c *gin.Context) {
		jsonraw := c.MustGet("message")

		switch json := jsonraw.(type) {
		case messageRaw:
			if err := providers.CleanUp(provider, json.Domain, json.Token, json.KeyAuth); err != nil {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok", "value": json.Token})
		case messageDefault:
			if err := provider.RemoveRecord(json.FQDN, json.Value); err != nil {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok", "value": json.Value})

		}
	}
}

func NewHandler(a auth.UserAuthenticator, p providers.ProviderSolved) (handler http.Handler, err error) {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"pong": time.Now().String(),
		})
	})
	router.POST("/present", verifyPermission(a), presentHandler(p))
	router.POST("/cleanup", verifyPermission(a), cleanuptHandler(p))

	return router, nil
}

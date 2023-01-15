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

type message struct {
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

func (c combinedMessage) is_raw() bool {
	return c.Domain != "" && c.Token != "" && c.KeyAuth != ""
}

func (c combinedMessage) is_default() bool {
	return c.FQDN != "" && c.Value != ""
}

func (c combinedMessage) as_msg() (message, error) {
	if c.is_default() {
		return message{c.FQDN, c.Value}, nil
	}

	if c.is_raw() {
		fqdn, value := dns01.GetRecord(c.Domain, c.KeyAuth)
		return message{fqdn, value}, nil
	}

	return message{}, errors.New("is not a valid request")
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

func parseRequest(c *gin.Context) {
	var combinedMsg combinedMessage
	if err := c.ShouldBindJSON(&combinedMsg); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := combinedMsg.as_msg()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Set("message", msg)
}

func verifyPermission(a auth.Authenticator) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, pass, err := getBasicAuth(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		msg := c.MustGet("message").(message)

		if err := a.VerifyPermissions(user, pass, msg.FQDN); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
	}
}

func presentHandler(provider providers.DNSProvider) func(c *gin.Context) {
	return func(c *gin.Context) {
		msg := c.MustGet("message").(message)

		if err := provider.CreateRecord(msg.FQDN, msg.Value); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "value": msg.Value})

	}
}

func cleanuptHandler(provider providers.DNSProvider) func(c *gin.Context) {
	return func(c *gin.Context) {
		json := c.MustGet("message").(message)

		if err := provider.RemoveRecord(json.FQDN, json.Value); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "value": json.Value})
	}
}

func pong(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pong": time.Now().String(),
	})
}

func NewHandler(a auth.Authenticator, p providers.DNSProvider) (handler http.Handler, err error) {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/ping", pong)
	router.POST("/present", parseRequest, verifyPermission(a), presentHandler(p))
	router.POST("/cleanup", parseRequest, verifyPermission(a), cleanuptHandler(p))

	return router, nil
}

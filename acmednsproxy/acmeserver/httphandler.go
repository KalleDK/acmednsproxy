package acmeserver

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

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

func (c combinedMessage) as_record() (acmeservice.Record, error) {
	if c.is_default() {
		return acmeservice.Record{
			FQDN:  c.FQDN,
			Value: c.Value,
		}, nil
	}

	if c.is_raw() {
		fqdn, value := dns01.GetRecord(c.Domain, c.KeyAuth)
		return acmeservice.Record{
			FQDN:  fqdn,
			Value: value,
		}, nil
	}

	return acmeservice.Record{}, errors.New("is not a valid request")
}

func getBasicAuth(c *gin.Context) {

	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Basic ") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing auth"})
		return
	}

	decodedAuthValue, err := base64.StdEncoding.DecodeString(h[6:])
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parts := bytes.SplitN(decodedAuthValue, []byte(":"), 2)
	if len(parts) != 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid auth header"})
		return
	}

	c.Set("auth", acmeservice.Auth{
		Username: string(parts[0]),
		Password: string(parts[1]),
	})
}

func getRecord(c *gin.Context) {
	var combinedMsg combinedMessage
	if err := c.ShouldBindJSON(&combinedMsg); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := combinedMsg.as_record()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Set("record", msg)
}

func presentHandler(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := c.MustGet("auth").(acmeservice.Auth)
		record := c.MustGet("record").(acmeservice.Record)

		if err := proxy.Authenticate(auth, record); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if err := proxy.Present(record); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "value": record.Value})

	}
}

func cleanupHandler(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := c.MustGet("auth").(acmeservice.Auth)
		record := c.MustGet("record").(acmeservice.Record)

		if err := proxy.Authenticate(auth, record); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if err := proxy.Cleanup(record); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "value": record.Value})
	}
}

func pong(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pong": time.Now().String(),
	})
}

func NewHandler(p *acmeservice.DNSProxy) (handler http.Handler, err error) {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/ping", pong)
	router.POST("/present", getBasicAuth, getRecord, presentHandler(p))
	router.POST("/cleanup", getBasicAuth, getRecord, cleanupHandler(p))

	return router, nil
}

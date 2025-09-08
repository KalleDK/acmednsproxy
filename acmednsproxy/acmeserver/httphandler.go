package acmeserver

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type domainTest struct {
	Domain string `json:"domain"`
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

func (c combinedMessage) as_record() (providers.Record, error) {
	if c.is_default() {
		return providers.Record{
			Fqdn:  c.FQDN,
			Value: c.Value,
		}, nil
	}

	if c.is_raw() {
		fqdn, value := dns01.GetRecord(c.Domain, c.KeyAuth)
		return providers.Record{
			Fqdn:  fqdn,
			Value: value,
		}, nil
	}

	return providers.Record{}, errors.New("is not a valid request")
}

func getBasicAuth(c *gin.Context) {

	const authPrefix = "Basic "

	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, authPrefix) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing auth"})
		return
	}

	decodedAuthValue, err := base64.StdEncoding.DecodeString(h[len(authPrefix):])
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parts := bytes.SplitN(decodedAuthValue, []byte(":"), 2)
	if len(parts) != 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid auth header"})
		return
	}

	c.Set("auth", auth.Credentials{
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

func getDomain(c *gin.Context) {
	var domainMsg domainTest
	if err := c.ShouldBindJSON(&domainMsg); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if domainMsg.Domain == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing domain"})
		return
	}
	c.Set("domain", domainMsg.Domain)
}

func presentHandler(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		cred := c.MustGet("auth").(auth.Credentials)
		record := c.MustGet("record").(providers.Record)

		log.Printf("Presenting %s for %s", record.Value, record.Fqdn)

		if err := proxy.Authenticate(cred, record.Fqdn); err != nil {
			log.Printf("Failed to authenticate %s for %s because %v", cred.Username, record.Fqdn, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if err := proxy.Present(record); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"FQDN": record.Fqdn, "Value": record.Value})

	}
}

func cleanupHandler(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		a := c.MustGet("auth").(auth.Credentials)
		record := c.MustGet("record").(providers.Record)

		if err := proxy.Authenticate(a, record.Fqdn); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if err := proxy.Cleanup(record); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"FQDN": record.Fqdn, "Value": record.Value})
	}
}

func reloadHandler(proxy *acmeservice.DNSProxy, cert *TLSService) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := proxy.Reload(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "value": fmt.Errorf("service: %w", err).Error()})
		}
		if err := cert.Reload(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "value": fmt.Errorf("cert: %w", err).Error()})
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "value": "reloaded"})
	}
}

func testAuth(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		cred := c.MustGet("auth").(auth.Credentials)
		domain := c.MustGet("domain").(string)

		if err := proxy.Authenticate(cred, "_acme-challenge."+domain+"."); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func pong(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pong": time.Now().String(),
	})
}

func NewHandler(p *acmeservice.DNSProxy, cert *TLSService) (handler http.Handler, err error) {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/ping", pong)
	router.POST("/domain", getBasicAuth, getDomain, testAuth(p))
	router.POST("/present", getBasicAuth, getRecord, presentHandler(p))
	router.POST("/cleanup", getBasicAuth, getRecord, cleanupHandler(p))
	router.POST("/reload", reloadHandler(p, cert))

	return router, nil
}

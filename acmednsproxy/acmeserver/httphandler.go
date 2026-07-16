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

// domainTest is the JSON body expected by the /domain endpoint.
type domainTest struct {
	Domain string `json:"domain"`
}

// combinedMessage is the JSON body accepted by /present and /cleanup.
// It supports two modes:
//   - RAW mode: Domain + Token + KeyAuth are provided; the FQDN and TXT value
//     are derived using dns01.GetRecord.
//   - Default mode: FQDN + Value are provided directly (e.g. from lego's
//     httpreq provider in default mode).
type combinedMessage struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"keyAuth"`
	FQDN    string `json:"fqdn"`
	Value   string `json:"value"`
}

// isRaw reports whether the message is in RAW mode.
func (c combinedMessage) isRaw() bool {
	return c.Domain != "" && c.Token != "" && c.KeyAuth != ""
}

// isDefault reports whether the message is in default (pre-computed) mode.
func (c combinedMessage) isDefault() bool {
	return c.FQDN != "" && c.Value != ""
}

// asRecord converts the message into a providers.Record, computing the FQDN
// and value from the raw fields if necessary.
func (c combinedMessage) asRecord() (providers.Record, error) {
	if c.isDefault() {
		return providers.Record{
			Fqdn:  c.FQDN,
			Value: c.Value,
		}, nil
	}

	if c.isRaw() {
		fqdn, value := dns01.GetRecord(c.Domain, c.KeyAuth)
		return providers.Record{
			Fqdn:  fqdn,
			Value: value,
		}, nil
	}

	return providers.Record{}, errors.New("is not a valid request")
}

// getBasicAuth extracts and decodes the HTTP Basic-Auth credentials from the
// Authorization header.
func getBasicAuth(c *gin.Context) (auth.Credentials, error) {
	const authPrefix = "Basic "

	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, authPrefix) {
		return auth.Credentials{}, errors.New("missing auth")
	}

	decodedAuthValue, err := base64.StdEncoding.DecodeString(h[len(authPrefix):])
	if err != nil {
		return auth.Credentials{}, err
	}

	parts := bytes.SplitN(decodedAuthValue, []byte(":"), 2)
	if len(parts) != 2 {
		return auth.Credentials{}, errors.New("invalid auth header")
	}

	return auth.Credentials{
		Username: string(parts[0]),
		Password: string(parts[1]),
	}, nil
}

// verifyAuth is a gin middleware that reads the Basic-Auth credentials from
// the request and verifies them against the proxy's authenticator for the
// domain stored in the gin context by getDomain or getRecord.
func verifyAuth(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		domain := c.MustGet("domain").(string)

		cred, err := getBasicAuth(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if err := proxy.Authenticate(cred, domain); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("auth", cred)
	}
}

// getRecord is a gin middleware that parses the request body as a
// combinedMessage, converts it to a providers.Record, validates the ACME
// challenge domain format, and stores both "domain" and "record" in the gin
// context for subsequent handlers.
func getRecord(c *gin.Context) {
	var combinedMsg combinedMessage
	if err := c.ShouldBindJSON(&combinedMsg); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := combinedMsg.asRecord()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := msg.Fqdn
	if !strings.HasPrefix(domain, "_acme-challenge.") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid challenge domain %s missing prefix", domain)})
		return
	}
	domain = strings.TrimPrefix(domain, "_acme-challenge.")

	if !strings.HasSuffix(domain, ".") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid challenge domain %s missing . at end", domain)})
		return
	}
	domain = strings.TrimSuffix(domain, ".")

	c.Set("domain", domain)
	c.Set("record", msg)
}

// getDomain is a gin middleware that parses the request body as a domainTest
// and stores the domain in the gin context.
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

// presentHandler is the gin handler for POST /present.  It reads the record
// from the context (set by getRecord) and calls proxy.Present.
func presentHandler(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		record := c.MustGet("record").(providers.Record)

		log.Printf("Presenting %s for %s", record.Value, record.Fqdn)

		if err := proxy.Present(record); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"FQDN": record.Fqdn, "Value": record.Value})
	}
}

// cleanupHandler is the gin handler for POST /cleanup.  It reads the record
// from the context (set by getRecord) and calls proxy.Cleanup.
func cleanupHandler(proxy *acmeservice.DNSProxy) func(c *gin.Context) {
	return func(c *gin.Context) {
		record := c.MustGet("record").(providers.Record)

		if err := proxy.Cleanup(record); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"FQDN": record.Fqdn, "Value": record.Value})
	}
}

// reloadHandler is the gin handler for POST /reload.  It triggers an in-place
// reload of the DNS proxy configuration and the TLS certificate.
// Note: this endpoint is unauthenticated; restrict access via network policy
// or firewall rules as appropriate.
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

// testAuth is the gin handler for POST /domain.  It responds with the
// authenticated user and domain, confirming that the credentials are valid.
func testAuth(c *gin.Context) {
	domain := c.MustGet("domain").(string)
	cred := c.MustGet("auth").(auth.Credentials)

	c.JSON(http.StatusOK, gin.H{"status": "ok", "domain": domain, "user": cred.Username})
}

// pong is the gin handler for GET /ping.  It returns the current server time
// and can be used as a liveness probe.
func pong(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pong": time.Now().String(),
	})
}

// NewHandler constructs the gin router with all routes configured.
//
// Routes:
//
//	GET  /ping    – liveness probe
//	POST /domain  – verify credentials for a domain
//	POST /present – publish a DNS-01 TXT record
//	POST /cleanup – remove a DNS-01 TXT record
//	POST /reload  – reload proxy config and TLS certificate (unauthenticated)
func NewHandler(p *acmeservice.DNSProxy, cert *TLSService) (handler http.Handler, err error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/ping", pong)
	router.POST("/domain", getDomain, verifyAuth(p), testAuth)
	router.POST("/present", getRecord, verifyAuth(p), presentHandler(p))
	router.POST("/cleanup", getRecord, verifyAuth(p), cleanupHandler(p))
	router.POST("/reload", reloadHandler(p, cert))

	return router, nil
}

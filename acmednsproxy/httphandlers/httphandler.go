package httphandlers

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/gin-gonic/gin"
)

type messageRaw struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"keyAuth"`
}

func getAuth(c *gin.Context) (string, string) {
	h := c.GetHeader("Authorization")
	d, _ := base64.StdEncoding.DecodeString(h[6:])
	ss := string(d)
	sss := strings.SplitN(ss, ":", 2)
	return sss[0], sss[1]
}

func verifyPermission(a auth.UserAuthenticator) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, pass := getAuth(c)

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

func presentHandler(p providers.ProviderBackend) func(c *gin.Context) {
	return func(c *gin.Context) {
		json := c.MustGet("message").(messageRaw)

		provider, err := p.GetProvider(json.Domain)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		if err := provider.Present(json.Domain, json.Token, json.KeyAuth); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func cleanuptHandler(p providers.ProviderBackend) func(c *gin.Context) {
	return func(c *gin.Context) {
		json := c.MustGet("message").(messageRaw)

		provider, err := p.GetProvider(json.Domain)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		if err := provider.CleanUp(json.Domain, json.Token, json.KeyAuth); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func NewHandler(a auth.UserAuthenticator, p providers.ProviderBackend) (handler http.Handler, err error) {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/present", verifyPermission(a), presentHandler(p))
	router.POST("/cleanup", verifyPermission(a), cleanuptHandler(p))

	return router, nil
}

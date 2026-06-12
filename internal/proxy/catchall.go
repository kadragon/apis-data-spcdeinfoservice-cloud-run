package proxy

import (
	"net/http"
	"path"
	"regexp"

	"github.com/gin-gonic/gin"
)

// catchAllPathRe restricts forwarded paths to ≥2 segments of URL-safe
// characters, blocking traversal and root-level probes (e.g. /favicon.ico).
var catchAllPathRe = regexp.MustCompile(`^(/[A-Za-z0-9._~-]+){2,}$`)

func NewCatchAllHandler(baseURL, serviceKey string, client *http.Client) gin.HandlerFunc {
	if client == nil {
		panic("NewCatchAllHandler: client must not be nil")
	}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
				"error":   "Method Not Allowed",
				"message": "Only GET is supported",
			})
			return
		}

		p := c.Request.URL.Path
		// path.Clean equality rejects "." / ".." segments the regexp's
		// character class would otherwise admit.
		if !catchAllPathRe.MatchString(p) || path.Clean(p) != p {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		q := c.Request.URL.Query()
		q.Set("serviceKey", serviceKey)
		target := baseURL + p + "?" + q.Encode()

		proxyTarget(c, client, target)
	}
}

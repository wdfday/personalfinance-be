package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName = "refresh_token"
)

// getRefreshTokenFromCookie gets the refresh token from cookie
func (h *AuthHandler) getRefreshTokenFromCookie(c *gin.Context) (string, error) {
	return c.Cookie(refreshTokenCookieName)
}

// setRefreshTokenCookie sets the refresh token in an HTTP-only cookie
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	// Parse SameSite mode
	sameSite := http.SameSiteLaxMode
	switch h.config.Auth.CookieSameSite {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	default:
		sameSite = http.SameSiteLaxMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		refreshTokenCookieName,       // name
		refreshToken,                 // value
		h.config.Auth.CookieMaxAge,   // maxAge (seconds)
		"/",                          // path
		h.config.Auth.CookieDomain,   // domain
		h.config.Auth.CookieSecure,   // secure
		h.config.Auth.CookieHTTPOnly, // httpOnly
	)
}

// clearRefreshTokenCookie clears the refresh token cookie
func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	// Use same SameSite mode for consistency
	sameSite := http.SameSiteLaxMode
	switch h.config.Auth.CookieSameSite {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	default:
		sameSite = http.SameSiteLaxMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		refreshTokenCookieName,
		"",
		-1,                           // maxAge -1 deletes the cookie
		"/",                          // path
		h.config.Auth.CookieDomain,   // domain
		h.config.Auth.CookieSecure,   // secure
		h.config.Auth.CookieHTTPOnly, // httpOnly
	)
}

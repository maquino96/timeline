package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const sessionCookie = "timeline_session"
const sessionTTL = 30 * 24 * time.Hour

func adminPasswordHash() string {
	return os.Getenv("ADMIN_PASSWORD")
}

func secretKey() string {
	k := os.Getenv("SECRET_KEY")
	if k == "" {
		k = "change-me"
	}
	return k
}

// authEnabled reports whether a management password is configured. When no
// ADMIN_PASSWORD is set, auth is disabled and all routes stay open (useful for
// local development).
func authEnabled() bool {
	return adminPasswordHash() != ""
}

func verifyPassword(password string) bool {
	want := adminPasswordHash()
	if want == "" {
		return false
	}
	sum := sha256.Sum256([]byte(password))
	got := hex.EncodeToString(sum[:])
	return hmac.Equal([]byte(got), []byte(strings.ToLower(want)))
}

func sign(value string) string {
	mac := hmac.New(sha256.New, []byte(secretKey()))
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func makeToken() string {
	exp := strconv.FormatInt(time.Now().Add(sessionTTL).Unix(), 10)
	return exp + "." + sign(exp)
}

func validToken(token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	exp, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(sign(exp)), []byte(sig)) {
		return false
	}
	expUnix, err := strconv.ParseInt(exp, 10, 64)
	if err != nil {
		return false
	}
	return time.Now().Unix() < expUnix
}

func isAuthed(r *http.Request) bool {
	if !authEnabled() {
		return true
	}
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	return validToken(c.Value)
}

func setSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    makeToken(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// requireAuth wraps a handler, rejecting unauthenticated requests with 401.
func (a *API) requireAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthed(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return
		}
		h(w, r)
	}
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !verifyPassword(body.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid password"})
		return
	}
	setSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": false})
}

func (a *API) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{
		"authenticated": isAuthed(r),
		"auth_required": authEnabled(),
	})
}

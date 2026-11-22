package httpapi

import (
	"errors"
	"net/http"
	"time"

	"establishment/v1/internal/auth"
)

const sessionCookieName = "session_id"

type registrationJSON struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type credentialsJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loggedInUserJSON struct {
	Login string `json:"login"`
}

type accountHandlers struct {
	accounts      accountManager
	secureCookies bool
}

func (h accountHandlers) register(w http.ResponseWriter, r *http.Request) {
	var input registrationJSON
	if !decodeJSON(w, r, &input) {
		return
	}
	registration := auth.Registration{Login: input.Login, Email: input.Email, Password: input.Password}
	if err := h.accounts.Register(r.Context(), registration); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h accountHandlers) login(w http.ResponseWriter, r *http.Request) {
	var input credentialsJSON
	if !decodeJSON(w, r, &input) {
		return
	}
	session, err := h.accounts.Login(r.Context(), input.Login, input.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	http.SetCookie(w, h.sessionCookie(session.ID, session.ExpiresAt))
	w.WriteHeader(http.StatusOK)
}

func (h accountHandlers) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, auth.ErrSessionNotFound)
		return
	}
	if err := h.accounts.Logout(r.Context(), cookie.Value); err != nil {
		writeError(w, err)
		return
	}
	http.SetCookie(w, h.sessionCookie("", time.Unix(0, 0)))
	w.WriteHeader(http.StatusOK)
}

func (h accountHandlers) checkSession(w http.ResponseWriter, r *http.Request) {
	user, ok := loggedInUser(r.Context())
	if !ok {
		writeError(w, errors.New("check-session reached without requireSession middleware"))
		return
	}
	writeJSON(w, http.StatusOK, loggedInUserJSON{Login: user.Login})
}

func (h accountHandlers) sessionCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
}

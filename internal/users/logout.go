package users

import (
	"net/http"
	"time"
)

func (u *UserHandle) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionToken, err := SessionTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized status", http.StatusUnauthorized)
		return
	}

	if err := u.Queries.DeleteSessionByTokenHash(ctx, HashSessionToken(sessionToken)); err != nil {
		u.Logger.Error("Error trying to delete session: "+err.Error(), nil)
		http.Error(w, "There was a problem logging you out", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookieName,
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Now(),
	})
	WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
	u.Logger.Info("Logged out successfully", nil)
}

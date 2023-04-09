package commands

import (
	"net/http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
)

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "chess-session", Path: "/", MaxAge: -1})
	core.WriteOK(w, r, nil)
}

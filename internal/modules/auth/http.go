package main

import (
	"http"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
)

type AuthHTTPHandler struct {

}

func (h *AuthHTTPHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	core.WriteOK(w, r, nil)
}

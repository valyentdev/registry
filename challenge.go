package main

import (
	"fmt"
	"net/http"
)

type challenge struct {
	err error
}

func (ch challenge) SetHeaders(r *http.Request, w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic")
}

func (ch challenge) Error() string {
	return fmt.Sprintf("basic authentication challenge: %s", ch.err)
}

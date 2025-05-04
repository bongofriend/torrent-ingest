package api

import "net/http"

const (
	unauthorizedMessage        string = "Unauthorized"
	internalServerErrorMessage string = "Internal Server Error"
	badRequestMessage          string = "Bad Request"
)

func unauthorized(w http.ResponseWriter) {
	http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, internalServerErrorMessage, http.StatusInternalServerError)
}

func badRequest(w http.ResponseWriter) {
	http.Error(w, badRequestMessage, http.StatusBadRequest)
}

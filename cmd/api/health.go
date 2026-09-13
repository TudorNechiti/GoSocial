package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

type echoPayload struct {
	Message string `json:"message"`
}

func (app *application) echoHandler(w http.ResponseWriter, r *http.Request) {
	var payload echoPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	app.jsonResponse(w, http.StatusOK, payload)
}

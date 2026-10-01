package main

import (
	"net/http"
)

// healthCheckHandler godoc
//
//	@Summary		Health check
//	@Description	Reports service status, environment, and version
//	@Tags			ops
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	ErrorResponse
//	@Router			/health [get]
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

// echoHandler is not currently wired to a route; no swagger annotations
// since there's no path to document.
func (app *application) echoHandler(w http.ResponseWriter, r *http.Request) {
	var payload echoPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	app.jsonResponse(w, http.StatusOK, payload)
}

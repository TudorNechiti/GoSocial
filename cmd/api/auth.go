package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/nechititudorr/GoSocial/internal/mailer"
	"github.com/nechititudorr/GoSocial/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Password string `json:"password" validate:"required,min=3,max=72"`
	Email    string `json:"email" validate:"required,email,max=255" `
}

// registerUserHandler godoc
//
//	@Summary		Register a user
//	@Description	Creates an inactive user account and emails them an activation invitation
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	RegisterUserPayload	true	"Registration payload"
//	@Success		201
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	// plainToken is what gets emailed to the user; the store layer hashes it
	// before persisting, mirroring how password hashing is centralized in
	// password.Set rather than done at the handler layer.
	plainToken := uuid.New().String()

	// store the user
	err := app.store.Users.CreateAndInvite(ctx, user, plainToken, app.config.mail.expiry)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	activationURL := fmt.Sprintf("%s/v1/users/activate/%s", app.config.apiUrl, plainToken)

	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	if err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv); err != nil {
		app.logger.Errorw("failed to send welcome email", "error", err)

		// the invitation was already persisted; roll it back so the user
		// isn't left with an unusable, un-emailed invite
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("failed to rollback user after failed mail send", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

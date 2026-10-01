package main

import (
	"net/http"

	"github.com/nechititudorr/GoSocial/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetch the user's feed
//	@Description	Fetches posts from the user and the users they follow, with pagination, sorting, search and tag filters
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int		false	"Number of posts to return"	default(20)
//	@Param			offset	query		int		false	"Number of posts to skip"	default(0)
//	@Param			sort	query		string	false	"Sort direction"			Enums(asc, desc)	default(desc)
//	@Param			search	query		string	false	"Search term matched against title/content"
//	@Param			tags	query		string	false	"Comma-separated list of tags to filter by"
//	@Success		200		{array}		store.PostWithMetadata
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// pagination, filters
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(42), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}

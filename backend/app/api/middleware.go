package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

type tokenHasKey struct {
	key string
}

var hashedToken = tokenHasKey{key: "hashedToken"}

type RoleMode int

const (
	RoleModeOr  RoleMode = 0
	RoleModeAnd RoleMode = 1
)

// mwRoles is a middleware that will validate the required roles are met. All roles
// are required to be met for the request to be allowed. If the user does not have
// the required roles, a 403 Forbidden will be returned.
//
// WARNING: This middleware _MUST_ be called after mwAuthToken or else it will panic
func (a *app) mwRoles(rm RoleMode, required ...string) errchain.Middleware {
	return func(next errchain.Handler) errchain.Handler {
		return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()

			maybeToken := ctx.Value(hashedToken)
			if maybeToken == nil {
				panic("mwRoles: token not found in context, you must call mwAuthToken before mwRoles")
			}

			// Authentication is disabled — all role checks pass.
			_ = maybeToken

			return next.ServeHTTP(w, r)
		})
	}
}


// mwAuthToken is a middleware that automatically authenticates requests as the
// first user in the database. Authentication is disabled — no token is required.
func (a *app) mwAuthToken(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		const noAuthToken = "no-auth"

		users, err := a.repos.Users.GetAll(r.Context())
		if err != nil || len(users) == 0 {
			return validate.NewRequestError(errors.New("no users found in database; please create a user first"), http.StatusInternalServerError)
		}

		usr := users[0]
		r = r.WithContext(context.WithValue(r.Context(), hashedToken, noAuthToken))
		r = r.WithContext(services.SetUserCtx(r.Context(), &usr, noAuthToken))
		return next.ServeHTTP(w, r)
	})
}

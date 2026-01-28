// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/github"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// OAuthStateCookie holds the state value that must come back unchanged from the provider.
const OAuthStateCookie = "_mut_oauth_state"

// GitHubOAuthStartAction redirects the browser to GitHub's authorize URL.
func (a *API) GitHubOAuthStartAction(w http.ResponseWriter, r *http.Request) {
	state, err := util.GenerateSecureToken(24)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to generate oauth state")
		redirectOAuthError(w, r, "github")
		return
	}

	setOAuthStateCookie(w, state)

	http.Redirect(w, r, a.GitHub.AuthorizeURL(state), http.StatusFound)
}

// GitHubOAuthCallbackAction completes GitHub sign-in and starts a session for the matching user.
func (a *API) GitHubOAuthCallbackAction(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	expectedState := util.GetCookie(r, OAuthStateCookie)

	util.DeleteCookie(w, OAuthStateCookie)

	token, err := a.GitHub.Exchange(r.Context(), code, state, expectedState)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth exchange failed")
		redirectOAuthError(w, r, "github")
		return
	}

	user, err := a.GitHub.User(r.Context(), token.AccessToken)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth user fetch failed")
		redirectOAuthError(w, r, "github")
		return
	}

	emails, err := a.GitHub.Emails(r.Context(), token.AccessToken)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth emails fetch failed")
		redirectOAuthError(w, r, "github")
		return
	}

	result, err := a.Auth.LoginWithOAuth(r.Context(), &module.OAuthIdentity{
		Provider:       db.UserProviderGithub,
		ProviderUserID: strconv.FormatInt(user.ID, 10),
		Email:          github.PrimaryEmail(emails, user.Email),
		Name:           lo.Ternary(lo.IsNotEmpty(user.Name), user.Name, user.Login),
	})
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth login failed")
		redirectOAuthError(w, r, "github")
		return
	}

	util.SetCookie(w, "_mut_session", result.Session.Token, result.CookieOptions)

	err = a.Invite.AttachPending(result.User)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", result.User.Id.String()).
			Msg("Failed to attach pending workspace invites")
	}

	http.Redirect(w, r, util.AppURL("/login?oauth=github"), http.StatusFound)
}

// GoogleOAuthStartAction redirects the browser to Google's authorize URL.
func (a *API) GoogleOAuthStartAction(w http.ResponseWriter, r *http.Request) {
	state, err := util.GenerateSecureToken(24)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to generate oauth state")
		redirectOAuthError(w, r, "google")
		return
	}

	setOAuthStateCookie(w, state)

	http.Redirect(w, r, a.Google.AuthorizeURL(state), http.StatusFound)
}

// GoogleOAuthCallbackAction completes Google sign-in and starts a session for the matching user.
func (a *API) GoogleOAuthCallbackAction(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	expectedState := util.GetCookie(r, OAuthStateCookie)

	util.DeleteCookie(w, OAuthStateCookie)

	token, err := a.Google.Exchange(r.Context(), code, state, expectedState)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Google oauth exchange failed")
		redirectOAuthError(w, r, "google")
		return
	}

	user, err := a.Google.User(r.Context(), token.AccessToken)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Google oauth user fetch failed")
		redirectOAuthError(w, r, "google")
		return
	}

	if !user.EmailVerified {
		log.Info().
			Str("email", user.Email).
			Msg("Google oauth rejected for unverified email")
		redirectOAuthError(w, r, "google")
		return
	}

	result, err := a.Auth.LoginWithOAuth(r.Context(), &module.OAuthIdentity{
		Provider:       db.UserProviderGoogle,
		ProviderUserID: user.Sub,
		Email:          user.Email,
		Name:           user.Name,
	})
	if err != nil {
		log.Error().
			Err(err).
			Msg("Google oauth login failed")
		redirectOAuthError(w, r, "google")
		return
	}

	util.SetCookie(w, "_mut_session", result.Session.Token, result.CookieOptions)

	err = a.Invite.AttachPending(result.User)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", result.User.Id.String()).
			Msg("Failed to attach pending workspace invites")
	}

	http.Redirect(w, r, util.AppURL("/login?oauth=google"), http.StatusFound)
}

// setOAuthStateCookie stores the state for ten minutes, sent on the provider's top-level redirect.
func setOAuthStateCookie(w http.ResponseWriter, state string) {
	opts := lo.Ternary(
		strings.HasPrefix(util.AppURL(""), "https://"),
		util.SecureCookieOptions(),
		util.DefaultCookieOptions(),
	)
	opts.SameSite = http.SameSiteLaxMode
	opts.MaxAge = int((10 * time.Minute) / time.Second)

	util.SetCookie(w, OAuthStateCookie, state, opts)
}

// redirectOAuthError sends the browser back to the login page with the provider's error flag.
func redirectOAuthError(w http.ResponseWriter, r *http.Request, provider string) {
	http.Redirect(w, r, util.AppURL("/login?oauth_error="+provider), http.StatusFound)
}

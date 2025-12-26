// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// CreateUserAction creates a user (admin only).
func (a *API) CreateUserAction(w http.ResponseWriter, r *http.Request) {
	var req module.CreateUserRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("email", req.Email).
		Msg("New create user request")

	user, err := a.User.CreateUser(&req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrUserEmailAlreadyExists, http.StatusConflict, locale.TR(r, "user_email_already_exists")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to create user")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_create_user"),
		})

		return
	}

	a.WriteJSON(w, http.StatusCreated, map[string]any{
		"user":           userResponse(user),
		"successMessage": locale.TR(r, "user_created_successfully"),
	})
}

// ListUsersAction returns a page of users (admin only).
func (a *API) ListUsersAction(w http.ResponseWriter, r *http.Request) {
	limit, offset := a.ParsePagination(r)

	result, err := a.User.ListUsers(limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list users")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_users"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"users": lo.Map(result.Users, func(user *db.User, _ int) map[string]any {
			return userResponse(user)
		}),
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetUserAction returns a user (admin only).
func (a *API) GetUserAction(w http.ResponseWriter, r *http.Request) {
	userId := db.Id(chi.URLParam(r, "userId"))

	user, err := a.User.GetUser(userId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrUserNotFound, http.StatusNotFound, locale.TR(r, "user_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_user"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"user": userResponse(user),
	})
}

// UpdateUserAction updates a user (admin only).
func (a *API) UpdateUserAction(w http.ResponseWriter, r *http.Request) {
	userId := db.Id(chi.URLParam(r, "userId"))

	var req module.UpdateUserRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("userId", userId.String()).
		Msg("New update user request")

	user, err := a.User.UpdateUser(userId, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrUserNotFound, http.StatusNotFound, locale.TR(r, "user_not_found")},
			{module.ErrUserEmailAlreadyExists, http.StatusConflict, locale.TR(r, "user_email_already_exists")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to update user")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_update_user"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"user":           userResponse(user),
		"successMessage": locale.TR(r, "user_updated_successfully"),
	})
}

// DeleteUserAction deletes a user (admin only).
func (a *API) DeleteUserAction(w http.ResponseWriter, r *http.Request) {
	userId := db.Id(chi.URLParam(r, "userId"))
	current, _ := a.GetUser(r)

	log.Info().
		Str("userId", userId.String()).
		Msg("New delete user request")

	err := a.User.DeleteUser(userId, current)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrUserNotFound, http.StatusNotFound, locale.TR(r, "user_not_found")},
			{module.ErrCannotDeleteSelf, http.StatusBadRequest, locale.TR(r, "cannot_delete_own_account")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to delete user")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_delete_user"),
		})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// userResponse shapes a user for API responses.
func userResponse(user *db.User) map[string]any {
	return map[string]any{
		"id":          user.Id,
		"name":        user.Name,
		"email":       user.Email,
		"role":        user.Role,
		"isActive":    user.IsActive,
		"avatar":      util.GravatarURL(user.Email, 80),
		"language":    user.Language,
		"theme":       user.Theme,
		"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

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

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// ListChatsAction returns the user's assistant chats in a workspace.
func (a *API) ListChatsAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	chats, err := a.Chat.ListChats(workspaceId, user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list chats")

		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})
		return
	}

	items := chats

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"chats": items,
	})
}

// CreateChatAction starts an empty assistant chat in a workspace.
func (a *API) CreateChatAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	chat, err := a.Chat.CreateChat(workspaceId, user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to create chat")

		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("chatId", chat.Id).
		Msg("Chat created")

	a.WriteJSON(w, http.StatusCreated, chat)
}

// GetChatAction returns one of the user's chats with its messages.
func (a *API) GetChatAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	chatId := db.Id(chi.URLParam(r, "chatId"))

	chat, err := a.Chat.GetChat(workspaceId, user.Id, chatId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrChatNotFound, http.StatusNotFound, locale.TR(r, "chat_not_found")},
			{module.ErrChatNotConfigured, http.StatusServiceUnavailable, locale.TR(r, "chat_not_configured")},
			{module.ErrChatBusy, http.StatusConflict, locale.TR(r, "chat_busy")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("chatId", chatId.String()).
			Msg("Failed to get chat")

		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, chat)
}

// DeleteChatAction removes one of the user's chats and its messages.
func (a *API) DeleteChatAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	chatId := db.Id(chi.URLParam(r, "chatId"))

	err := a.Chat.DeleteChat(workspaceId, user.Id, chatId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrChatNotFound, http.StatusNotFound, locale.TR(r, "chat_not_found")},
			{module.ErrChatNotConfigured, http.StatusServiceUnavailable, locale.TR(r, "chat_not_configured")},
			{module.ErrChatBusy, http.StatusConflict, locale.TR(r, "chat_busy")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("chatId", chatId.String()).
			Msg("Failed to delete chat")

		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("chatId", chatId.String()).
		Msg("Chat deleted")

	w.WriteHeader(http.StatusNoContent)
}

// SendChatMessageAction stores a question and streams the assistant's answer as server-sent events.
func (a *API) SendChatMessageAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	chatId := db.Id(chi.URLParam(r, "chatId"))

	var req module.SendChatMessageRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	chat, err := a.Chat.SendMessage(workspaceId, user.Id, chatId, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrChatNotFound, http.StatusNotFound, locale.TR(r, "chat_not_found")},
			{module.ErrChatNotConfigured, http.StatusServiceUnavailable, locale.TR(r, "chat_not_configured")},
			{module.ErrChatBusy, http.StatusConflict, locale.TR(r, "chat_busy")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("chatId", chatId.String()).
			Msg("Failed to send chat message")

		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})
		return
	}

	// An answer can outlive the server write timeout, so the stream gets its own deadline.
	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("chatId", chatId.String()).
		Msg("Chat answer started")

	a.Chat.Reply(chat, a.EventStream(w, module.ChatReplyTimeout+time.Minute))
}

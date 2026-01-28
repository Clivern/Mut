// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	// ChatStatusIdle means the assistant is not answering the chat.
	ChatStatusIdle = "idle"
	// ChatStatusRunning means the assistant is answering the chat now.
	ChatStatusRunning = "running"
	// ChatStatusFailed means the last answer did not finish.
	ChatStatusFailed = "failed"

	// ChatMessageKindUser marks a message the user wrote.
	ChatMessageKindUser = "user"
	// ChatMessageKindAssistant marks a message the assistant wrote.
	ChatMessageKindAssistant = "assistant"
)

// Chat is one assistant conversation a user holds in a workspace. ClusterId is
type Chat struct {
	Id          Id
	WorkspaceId Id
	UserId      Id
	ClusterId   string
	Title       string
	Status      string
	Preview     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ChatMessage is one message in a chat.
type ChatMessage struct {
	Id        Id
	ChatId    Id
	Kind      string
	Content   string
	CreatedAt time.Time
}

// ChatRepository is the interface for assistant chats and their messages.
type ChatRepository interface {
	CreateChat(chat *Chat) error
	GetChat(workspaceId, userId, id Id) (*Chat, error)
	ListChats(workspaceId, userId Id, limit, offset int) ([]*Chat, error)
	UpdateChat(chat *Chat) error
	DeleteChat(workspaceId, userId, id Id) (bool, error)
	CreateMessage(message *ChatMessage) error
	ListMessages(chatId Id) ([]*ChatMessage, error)
}

// ChatRepositoryPostgres implements ChatRepository for PostgreSQL
type ChatRepositoryPostgres struct {
	db *sql.DB
}

// NewChatRepository returns the chat repository for the current driver
func NewChatRepository(db *sql.DB) ChatRepository {
	return &ChatRepositoryPostgres{db: db}
}

// scanChat maps a row into a chat. Column order must match the select list.
func scanChat(chat *Chat) []any {
	return []any{
		&chat.Id,
		&chat.WorkspaceId,
		&chat.UserId,
		&chat.ClusterId,
		&chat.Title,
		&chat.Status,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	}
}

// scanChatPreview maps a row with the latest message preview. Column order must match the select list.
func scanChatPreview(chat *Chat) []any {
	return []any{
		&chat.Id,
		&chat.WorkspaceId,
		&chat.UserId,
		&chat.ClusterId,
		&chat.Title,
		&chat.Status,
		&chat.Preview,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	}
}

// scanChatMessage maps a row into a chat message. Column order must match the select list.
func scanChatMessage(message *ChatMessage) []any {
	return []any{
		&message.Id,
		&message.ChatId,
		&message.Kind,
		&message.Content,
		&message.CreatedAt,
	}
}

// CreateChat inserts a chat and sets its Id and timestamps
func (r *ChatRepositoryPostgres) CreateChat(chat *Chat) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	chat.Id = id

	return r.db.QueryRow(
		`INSERT INTO chats
		(id, workspace_id, user_id, cluster_id, title, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		chat.Id.String(),
		chat.WorkspaceId.String(),
		chat.UserId.String(),
		chat.ClusterId,
		chat.Title,
		chat.Status,
	).Scan(
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
}

// GetChat returns a chat that belongs to the user in the workspace, or nil
func (r *ChatRepositoryPostgres) GetChat(workspaceId, userId, id Id) (*Chat, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, user_id, cluster_id, title, status, created_at, updated_at
		FROM chats
		WHERE id = $1 AND workspace_id = $2 AND user_id = $3`,
		scanChat,
		id.String(),
		workspaceId.String(),
		userId.String(),
	)
}

// ListChats returns a page of a user's chats in a workspace, most recently updated first, with a preview
func (r *ChatRepositoryPostgres) ListChats(workspaceId, userId Id, limit, offset int) ([]*Chat, error) {
	return QueryMany(
		r.db,
		`SELECT
			c.id, c.workspace_id, c.user_id, c.cluster_id, c.title, c.status,
			(
				SELECT m.content
				FROM chat_messages m
				WHERE m.chat_id = c.id
				ORDER BY m.created_at DESC, m.id DESC
				LIMIT 1
			),
			c.created_at, c.updated_at
		FROM chats c
		WHERE c.workspace_id = $1 AND c.user_id = $2
		ORDER BY c.updated_at DESC
		LIMIT $3 OFFSET $4`,
		scanChatPreview,
		workspaceId.String(),
		userId.String(),
		limit,
		offset,
	)
}

// UpdateChat saves a chat's title, status and cluster and marks it active now
func (r *ChatRepositoryPostgres) UpdateChat(chat *Chat) error {
	return r.db.QueryRow(
		`UPDATE chats
		SET cluster_id = $2, title = $3, status = $4, updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		WHERE id = $1
		RETURNING updated_at`,
		chat.Id.String(),
		chat.ClusterId,
		chat.Title,
		chat.Status,
	).Scan(&chat.UpdatedAt)
}

// DeleteChat removes a user's chat in a workspace and its messages, reporting whether one was removed
func (r *ChatRepositoryPostgres) DeleteChat(workspaceId, userId, id Id) (bool, error) {
	result, err := r.db.Exec(
		`DELETE FROM chats WHERE id = $1 AND workspace_id = $2 AND user_id = $3`,
		id.String(),
		workspaceId.String(),
		userId.String(),
	)
	if err != nil {
		return false, err
	}

	removed, err := result.RowsAffected()

	return removed > 0, err
}

// CreateMessage inserts a message and sets its Id and timestamp
func (r *ChatRepositoryPostgres) CreateMessage(message *ChatMessage) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	message.Id = id

	return r.db.QueryRow(
		`INSERT INTO chat_messages
		(id, chat_id, kind, content)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		message.Id.String(),
		message.ChatId.String(),
		message.Kind,
		message.Content,
	).Scan(&message.CreatedAt)
}

// ListMessages returns a chat's messages, oldest first
func (r *ChatRepositoryPostgres) ListMessages(chatId Id) ([]*ChatMessage, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, chat_id, kind, content, created_at
		FROM chat_messages
		WHERE chat_id = $1
		ORDER BY created_at ASC, id ASC`,
		scanChatMessage,
		chatId.String(),
	)
}

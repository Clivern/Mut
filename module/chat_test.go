// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/llm"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubChatRepo keeps chats and messages in memory for unit tests.
type stubChatRepo struct {
	chats    map[db.Id]*db.Chat
	messages []*db.ChatMessage
}

func (s *stubChatRepo) CreateChat(chat *db.Chat) error {
	chat.Id = db.Id(fmt.Sprintf("chat-%d", len(s.chats)+1))
	chat.CreatedAt = time.Now().UTC()
	chat.UpdatedAt = chat.CreatedAt
	s.chats[chat.Id] = chat

	return nil
}

func (s *stubChatRepo) GetChat(workspaceId, userId, id db.Id) (*db.Chat, error) {
	chat, ok := s.chats[id]
	if !ok || chat.WorkspaceId != workspaceId || chat.UserId != userId {
		return nil, nil
	}

	return chat, nil
}

func (s *stubChatRepo) ListChats(workspaceId, userId db.Id, limit, offset int) ([]*db.Chat, error) {
	return nil, nil
}

func (s *stubChatRepo) UpdateChat(chat *db.Chat) error {
	stored := s.chats[chat.Id]
	stored.Title = chat.Title
	stored.Status = chat.Status
	stored.ClusterId = chat.ClusterId

	return nil
}

func (s *stubChatRepo) DeleteChat(workspaceId, userId, id db.Id) (bool, error) {
	chat, err := s.GetChat(workspaceId, userId, id)
	if err != nil || chat == nil {
		return false, err
	}

	delete(s.chats, id)

	return true, nil
}

func (s *stubChatRepo) CreateMessage(message *db.ChatMessage) error {
	message.Id = db.Id(fmt.Sprintf("message-%d", len(s.messages)+1))
	message.CreatedAt = time.Now().UTC()
	s.messages = append(s.messages, message)

	return nil
}

func (s *stubChatRepo) ListMessages(chatId db.Id) ([]*db.ChatMessage, error) {
	var out []*db.ChatMessage
	for _, message := range s.messages {
		if message.ChatId == chatId {
			out = append(out, message)
		}
	}

	return out, nil
}

// stubModel plans no tool calls and streams a fixed answer in two pieces.
type stubModel struct {
	streamed []llm.Message
}

func (s *stubModel) Complete(ctx context.Context, messages []llm.Message) (string, llm.Usage, error) {
	return `{"calls": []}`, llm.Usage{}, nil
}

func (s *stubModel) Stream(ctx context.Context, messages []llm.Message, onDelta func(string)) (string, llm.Usage, error) {
	s.streamed = messages
	onDelta("The cluster ")
	onDelta("is healthy.")

	return "The cluster is healthy.", llm.Usage{}, nil
}

func newTestChat(repo *stubChatRepo, model *stubModel) *Chat {
	return NewChat(repo, NewChatTools(nil), func() ChatModel { return model })
}

func TestUnitChatSendAndReply(t *testing.T) {
	viper.Set("app.ai.api_key", "test-key")
	defer viper.Set("app.ai.api_key", "")

	repo := &stubChatRepo{chats: map[db.Id]*db.Chat{}}
	model := &stubModel{}
	c := newTestChat(repo, model)

	created, err := c.CreateChat("w-1", "u-1")
	require.NoError(t, err)

	chat, err := c.SendMessage("w-1", "u-1", db.Id(created.Id), &SendChatMessageRequest{Content: "Is the cluster healthy?"})
	require.NoError(t, err)
	assert.Equal(t, db.ChatStatusRunning, chat.Status)

	var events []string
	c.Reply(chat, func(event string, data any) { events = append(events, event) })

	assert.Equal(t, []string{ChatEventChat, ChatEventDelta, ChatEventDelta, ChatEventDone}, events)
	assert.Equal(t, "Is the cluster healthy?", chat.Title)
	assert.Equal(t, db.ChatStatusIdle, chat.Status)

	thread, err := c.GetChat("w-1", "u-1", db.Id(created.Id))
	require.NoError(t, err)
	require.Len(t, thread.Messages, 2)
	assert.Equal(t, db.ChatMessageKindUser, thread.Messages[0].Kind)
	assert.Equal(t, db.ChatMessageKindAssistant, thread.Messages[1].Kind)
	assert.Equal(t, "The cluster is healthy.", thread.Messages[1].Content)
}

func TestUnitChatSendBusy(t *testing.T) {
	viper.Set("app.ai.api_key", "test-key")
	defer viper.Set("app.ai.api_key", "")

	repo := &stubChatRepo{chats: map[db.Id]*db.Chat{}}
	c := newTestChat(repo, &stubModel{})

	created, err := c.CreateChat("w-1", "u-1")
	require.NoError(t, err)

	_, err = c.SendMessage("w-1", "u-1", db.Id(created.Id), &SendChatMessageRequest{Content: "First"})
	require.NoError(t, err)

	_, err = c.SendMessage("w-1", "u-1", db.Id(created.Id), &SendChatMessageRequest{Content: "Second"})
	assert.ErrorIs(t, err, ErrChatBusy)
}

func TestUnitChatSendNotConfigured(t *testing.T) {
	viper.Set("app.ai.api_key", "")

	repo := &stubChatRepo{chats: map[db.Id]*db.Chat{}}
	c := newTestChat(repo, &stubModel{})

	created, err := c.CreateChat("w-1", "u-1")
	require.NoError(t, err)

	_, err = c.SendMessage("w-1", "u-1", db.Id(created.Id), &SendChatMessageRequest{Content: "Hi"})
	assert.ErrorIs(t, err, ErrChatNotConfigured)
}

func TestUnitChatOtherUserCannotRead(t *testing.T) {
	repo := &stubChatRepo{chats: map[db.Id]*db.Chat{}}
	c := newTestChat(repo, &stubModel{})

	created, err := c.CreateChat("w-1", "u-1")
	require.NoError(t, err)

	_, err = c.GetChat("w-1", "u-2", db.Id(created.Id))
	assert.ErrorIs(t, err, ErrChatNotFound)

	err = c.DeleteChat("w-2", "u-1", db.Id(created.Id))
	assert.ErrorIs(t, err, ErrChatNotFound)
}

func TestUnitChatToolWithoutClusterReportsError(t *testing.T) {
	repo := &stubChatRepo{chats: map[db.Id]*db.Chat{}}
	c := newTestChat(repo, &stubModel{})

	chat := &db.Chat{Id: "chat-1", WorkspaceId: "w-1", UserId: "u-1"}

	result, err := c.Call(context.Background(), chat, c.Tools, ChatToolCall{Tool: "clusterSummary"})
	require.NoError(t, err)
	assert.True(t, strings.Contains(result, ErrNoCluster.Error()))
}

func TestUnitChatTitleLimit(t *testing.T) {
	long := strings.Repeat("é", 80)

	assert.Len(t, []rune(chatTitle(long)), ChatTitleLength)
	assert.Equal(t, "short", chatTitle("short"))
}

// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/mut/conf"
	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/llm"

	"github.com/rs/zerolog/log"
)

const (
	// ChatMaxCalls caps how many tools one question can run.
	ChatMaxCalls = 3
	// ChatOutputLength caps the tool output the model sees for one call.
	ChatOutputLength = 8000
	// ChatTitleLength is the longest title taken from a first question.
	ChatTitleLength = 60
	// ChatListLimit caps how many chats the list returns, most recent first.
	ChatListLimit = 50
	// ChatReplyTimeout bounds one answer, including its tool calls.
	ChatReplyTimeout = 3 * time.Minute

	// ChatEventChat opens the stream with the chat as it is while the assistant answers.
	ChatEventChat = "chat"
	// ChatEventDelta carries one piece of the answer.
	ChatEventDelta = "delta"
	// ChatEventDone closes the stream with the finished chat.
	ChatEventDone = "done"
	// ChatEventError closes the stream when the answer failed.
	ChatEventError = "error"
)

var (
	// ErrChatNotFound is returned when a chat does not exist or belongs to another user or workspace.
	ErrChatNotFound = errors.New("chat not found")
	// ErrChatNotConfigured is returned when no AI API key is set.
	ErrChatNotConfigured = errors.New("assistant is not configured")
	// ErrChatBusy is returned when the chat is still being answered.
	ErrChatBusy = errors.New("assistant is still answering")
	// ErrFailedChatRequest is returned when a chat request fails for another reason.
	ErrFailedChatRequest = errors.New("failed chat request")
)

// ChatModel is the language model the assistant plans with and answers with.
type ChatModel interface {
	Complete(ctx context.Context, messages []llm.Message) (string, llm.Usage, error)
	Stream(ctx context.Context, messages []llm.Message, onDelta func(string)) (string, llm.Usage, error)
}

// ChatEmit sends one named event with its data to the user while the assistant answers.
type ChatEmit func(event string, data any)

// SendChatMessageRequest is the body for asking the assistant a question. ClusterId
type SendChatMessageRequest struct {
	Content   string `json:"content" validate:"required,max=4000"`
	ClusterId string `json:"clusterId"`
}

// ChatPlan is the set of tool calls the model picked for a question.
type ChatPlan struct {
	Calls []ChatToolCall `json:"calls"`
}

// ChatToolCall is one tool the model asked for, with its string arguments.
type ChatToolCall struct {
	Tool      string            `json:"tool"`
	Arguments map[string]string `json:"arguments"`
}

// ChatResponse is one chat for API responses. Messages are set only for a single chat.
type ChatResponse struct {
	Id        string                 `json:"id"`
	Title     string                 `json:"title"`
	Status    string                 `json:"status"`
	ClusterId string                 `json:"clusterId"`
	Preview   string                 `json:"preview"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
	Messages  []*ChatMessageResponse `json:"messages,omitempty"`
}

// ChatMessageResponse is one message in a chat for API responses.
type ChatMessageResponse struct {
	Id        string `json:"id"`
	Kind      string `json:"kind"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// ChatThread is a chat with its messages, oldest first.
type ChatThread struct {
	Chat     *db.Chat
	Messages []*db.ChatMessage
}

// Chat serves the workspace assistant. Each chat belongs to one user, so
type Chat struct {
	ChatRepository db.ChatRepository
	Tools          []*ChatTool
	NewModel       func() ChatModel
}

// NewChat creates the assistant module with its tools and model.
func NewChat(chats db.ChatRepository, tools []*ChatTool, newModel func() ChatModel) *Chat {
	return &Chat{
		ChatRepository: chats,
		Tools:          tools,
		NewModel:       newModel,
	}
}

// ListChats returns the user's chats in a workspace, most recent first.
func (c *Chat) ListChats(workspaceId, userId db.Id) ([]*ChatResponse, error) {
	chats, err := c.ChatRepository.ListChats(workspaceId, userId, ChatListLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	responses := make([]*ChatResponse, 0, len(chats))
	for _, chat := range chats {
		responses = append(responses, ToChatResponse(chat, nil))
	}

	return responses, nil
}

// CreateChat starts an empty chat for the user in a workspace.
func (c *Chat) CreateChat(workspaceId, userId db.Id) (*ChatResponse, error) {
	chat := &db.Chat{
		WorkspaceId: workspaceId,
		UserId:      userId,
		Status:      db.ChatStatusIdle,
	}

	err := c.ChatRepository.CreateChat(chat)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	return ToChatResponse(chat, []*db.ChatMessage{}), nil
}

// GetChat returns a chat with its messages, oldest first.
func (c *Chat) GetChat(workspaceId, userId, id db.Id) (*ChatResponse, error) {
	thread, err := c.loadThread(workspaceId, userId, id)
	if err != nil {
		return nil, err
	}

	return ToChatResponse(thread.Chat, thread.Messages), nil
}

// DeleteChat removes a chat and its messages.
func (c *Chat) DeleteChat(workspaceId, userId, id db.Id) error {
	removed, err := c.ChatRepository.DeleteChat(workspaceId, userId, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	if !removed {
		return ErrChatNotFound
	}

	return nil
}

// SendMessage stores the question and marks the chat as running. Reply answers it.
func (c *Chat) SendMessage(workspaceId, userId, id db.Id, req *SendChatMessageRequest) (*db.Chat, error) {
	if !llm.IsConfigured() {
		return nil, ErrChatNotConfigured
	}

	chat, err := c.ChatRepository.GetChat(workspaceId, userId, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	if chat == nil {
		return nil, ErrChatNotFound
	}

	if chat.Status == db.ChatStatusRunning {
		return nil, ErrChatBusy
	}

	content := strings.TrimSpace(req.Content)

	err = c.ChatRepository.CreateMessage(&db.ChatMessage{
		ChatId:  chat.Id,
		Kind:    db.ChatMessageKindUser,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	if chat.Title == "" {
		chat.Title = chatTitle(content)
	}

	chat.ClusterId = req.ClusterId
	chat.Status = db.ChatStatusRunning

	err = c.ChatRepository.UpdateChat(chat)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	return chat, nil
}

// Reply streams the answer and its outcome to emit, and stores the answer even if the user leaves mid-stream.
func (c *Chat) Reply(chat *db.Chat, emit ChatEmit) {
	ctx, cancel := context.WithTimeout(context.Background(), ChatReplyTimeout)
	defer cancel()

	started, err := c.GetChat(chat.WorkspaceId, chat.UserId, chat.Id)
	if err != nil {
		emit(ChatEventError, map[string]string{"status": db.ChatStatusFailed})
		return
	}

	emit(ChatEventChat, started)

	log.Info().
		Str("workspaceId", chat.WorkspaceId.String()).
		Str("chatId", chat.Id.String()).
		Str("clusterId", chat.ClusterId).
		Msg("Assistant answering chat")

	err = c.Answer(ctx, chat, emit)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", chat.WorkspaceId.String()).
			Str("chatId", chat.Id.String()).
			Msg("Failed to answer chat")
	}

	chat.Status = db.ChatStatusIdle
	if err != nil {
		chat.Status = db.ChatStatusFailed
	}

	err = c.ChatRepository.UpdateChat(chat)
	if err != nil {
		log.Error().
			Err(err).
			Str("chatId", chat.Id.String()).
			Msg("Failed to update chat")
	}

	finished, err := c.GetChat(chat.WorkspaceId, chat.UserId, chat.Id)
	if err != nil {
		emit(ChatEventError, map[string]string{"status": db.ChatStatusFailed})
		return
	}

	log.Info().
		Str("chatId", chat.Id.String()).
		Str("status", finished.Status).
		Msg("Assistant finished chat")

	if finished.Status == db.ChatStatusFailed {
		emit(ChatEventError, finished)
		return
	}

	emit(ChatEventDone, finished)
}

// Answer plans one round of tools, runs them, then streams and stores the answer. Tool results are never stored.
func (c *Chat) Answer(ctx context.Context, chat *db.Chat, emit ChatEmit) error {
	messages, err := c.ChatRepository.ListMessages(chat.Id)
	if err != nil {
		return err
	}

	model := c.NewModel()
	history := ChatHistory(messages)

	plan, err := c.Plan(ctx, model, history, c.Tools, nil)
	if err != nil {
		return err
	}

	results := []string{}
	for _, call := range plan.Calls {
		result, err := c.Call(ctx, chat, c.Tools, call)
		if err != nil {
			return err
		}

		results = append(results, result)
	}

	// Looked-up data goes in the instructions, not the chat, so the model answers from it instead of narrating it.
	descriptions := make([]string, 0, len(c.Tools))
	for _, tool := range c.Tools {
		descriptions = append(descriptions, tool.Description)
	}

	prompt := conf.AssistantAnswer(descriptions, results)

	answer, _, err := model.Stream(ctx, append([]llm.Message{{Role: "system", Content: prompt}}, history...), func(delta string) {
		emit(ChatEventDelta, map[string]string{"content": delta})
	})
	if err != nil {
		return err
	}

	return c.ChatRepository.CreateMessage(&db.ChatMessage{
		ChatId:  chat.Id,
		Kind:    db.ChatMessageKindAssistant,
		Content: strings.TrimSpace(answer),
	})
}

// Plan asks the model which tools to run next; a reply that is not a plan means no calls.
func (c *Chat) Plan(ctx context.Context, model ChatModel, history []llm.Message, tools []*ChatTool, results []string) (*ChatPlan, error) {
	if len(tools) == 0 {
		return &ChatPlan{}, nil
	}

	choices, _ := json.Marshal(tools)

	reply, _, err := model.Complete(ctx, append(
		append([]llm.Message{{Role: "system", Content: conf.AssistantPlanPrompt}}, history...),
		llm.Message{Role: "user", Content: fmt.Sprintf(
			"Tools you may run: %s\nData already looked up for this message:\n%s",
			choices,
			ternary(len(results) == 0, "none", strings.Join(results, "\n")),
		)},
	))
	if err != nil {
		return nil, err
	}

	plan := &ChatPlan{}
	start, end := strings.Index(reply, "{"), strings.LastIndex(reply, "}")
	if start == -1 || end < start {
		return plan, nil
	}

	_ = json.Unmarshal([]byte(reply[start:end+1]), plan)

	allowed := map[string]bool{}
	for _, tool := range tools {
		allowed[tool.Name] = true
	}

	calls := make([]ChatToolCall, 0, len(plan.Calls))
	for _, call := range plan.Calls {
		if allowed[call.Tool] {
			calls = append(calls, call)
		}
	}

	if len(calls) > ChatMaxCalls {
		calls = calls[:ChatMaxCalls]
	}

	plan.Calls = calls

	return plan, nil
}

// Call runs one tool for the chat's cluster and returns its outcome as a line of cluster data for the model.
func (c *Chat) Call(ctx context.Context, chat *db.Chat, tools []*ChatTool, call ChatToolCall) (string, error) {
	var tool *ChatTool
	for _, candidate := range tools {
		if candidate.Name == call.Tool {
			tool = candidate
		}
	}

	arguments := call.Arguments
	if arguments == nil {
		arguments = map[string]string{}
	}

	input, _ := json.Marshal(arguments)

	result, err := tool.Run(ctx, chat.WorkspaceId, db.Id(chat.ClusterId), arguments)
	if err != nil {
		log.Info().
			Err(err).
			Str("tool", call.Tool).
			Str("chatId", chat.Id.String()).
			Msg("Assistant tool could not be read")

		return fmt.Sprintf("%s %s: could not be read (%s)", call.Tool, input, err), nil
	}

	output, _ := json.Marshal(result)

	return fmt.Sprintf("%s %s: %s", call.Tool, input, truncate(string(output), ChatOutputLength)), nil
}

// loadThread loads a user's chat with its messages, rejecting one that is missing or belongs to someone else.
func (c *Chat) loadThread(workspaceId, userId, id db.Id) (*ChatThread, error) {
	chat, err := c.ChatRepository.GetChat(workspaceId, userId, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	if chat == nil {
		return nil, ErrChatNotFound
	}

	messages, err := c.ChatRepository.ListMessages(chat.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	if messages == nil {
		messages = []*db.ChatMessage{}
	}

	return &ChatThread{Chat: chat, Messages: messages}, nil
}

// ChatHistory turns stored messages into model messages.
func ChatHistory(messages []*db.ChatMessage) []llm.Message {
	history := make([]llm.Message, 0, len(messages))
	for _, message := range messages {
		history = append(history, llm.Message{Role: message.Kind, Content: message.Content})
	}

	return history
}

// ToChatResponse shapes a chat and, when given, its messages for API responses.
func ToChatResponse(chat *db.Chat, messages []*db.ChatMessage) *ChatResponse {
	response := &ChatResponse{
		Id:        chat.Id.String(),
		Title:     chat.Title,
		Status:    chat.Status,
		ClusterId: chat.ClusterId,
		CreatedAt: chat.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: chat.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if chat.Preview != nil {
		response.Preview = *chat.Preview
	}

	if messages != nil {
		response.Messages = make([]*ChatMessageResponse, 0, len(messages))
		for _, message := range messages {
			response.Messages = append(response.Messages, &ChatMessageResponse{
				Id:        message.Id.String(),
				Kind:      message.Kind,
				Content:   message.Content,
				CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
	}

	return response
}

// chatTitle names a chat after its first question.
func chatTitle(content string) string {
	return truncate(content, ChatTitleLength)
}

// truncate cuts text to a number of runes.
func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit])
}

func ternary[T any](condition bool, yes, no T) T {
	if condition {
		return yes
	}

	return no
}

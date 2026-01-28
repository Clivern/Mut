// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package llm

import (
	"context"
	"fmt"
	"strings"

	orsdk "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/optionalnullable"
	"github.com/spf13/viper"
)

// Message is a chat message passed to Complete and Stream.
type Message struct {
	Role    string
	Content string
}

// ChatClient calls oroute chat completion APIs for a single model tier.
type ChatClient struct {
	api   *orsdk.OpenRouter
	model string
}

// IsConfigured reports whether an AI API key is set for the assistant.
func IsConfigured() bool {
	return viper.GetString("app.ai.api_key") != ""
}

// NewChatClient returns the chat LLM client loaded from app.ai.llm config.
func NewChatClient() *ChatClient {
	return &ChatClient{
		api: orsdk.New(
			orsdk.WithSecurity(viper.GetString("app.ai.api_key")),
		),
		model: viper.GetString("app.ai.llm.model"),
	}
}

// Complete sends messages using this client's model and returns the reply and usage.
func (c *ChatClient) Complete(ctx context.Context, messages []Message) (string, Usage, error) {
	items := make([]components.ChatMessages, 0, len(messages))

	for _, message := range messages {
		var item components.ChatMessages
		switch message.Role {
		case "system":
			item = components.CreateChatMessagesSystem(components.ChatSystemMessage{
				Role:    components.ChatSystemMessageRoleSystem,
				Content: components.CreateChatSystemMessageContentStr(message.Content),
			})
		case "assistant":
			assistantContent := components.CreateChatAssistantMessageContentStr(message.Content)
			item = components.CreateChatMessagesAssistant(components.ChatAssistantMessage{
				Role:    components.ChatAssistantMessageRoleAssistant,
				Content: optionalnullable.From(&assistantContent),
			})
		case "user", "":
			item = components.CreateChatMessagesUser(components.ChatUserMessage{
				Role:    components.ChatUserMessageRoleUser,
				Content: components.CreateChatUserMessageContentStr(message.Content),
			})
		default:
			return "", Usage{}, fmt.Errorf("ai chat: unsupported role %q", message.Role)
		}

		items = append(items, item)
	}

	response, err := c.api.Chat.Send(ctx, components.ChatRequest{
		Model:    orsdk.Pointer(c.model),
		Messages: items,
	}, nil)
	if err != nil {
		return "", Usage{}, fmt.Errorf("ai chat: %w", err)
	}

	if response == nil || response.ChatResult == nil {
		return "", Usage{}, ErrNoCompletion
	}

	choices := response.ChatResult.GetChoices()
	if len(choices) == 0 {
		return "", Usage{}, ErrNoCompletion
	}

	assistantMessage := choices[0].GetMessage()
	content, ok := assistantMessage.GetContent().GetOrZero()
	if !ok {
		return "", Usage{}, ErrNoCompletion
	}

	usage := Usage{}
	if chatUsage := response.ChatResult.GetUsage(); chatUsage != nil {
		usage = Usage{
			PromptTokens: chatUsage.GetPromptTokens(),
			TotalTokens:  chatUsage.GetTotalTokens(),
		}
		if amount, ok := chatUsage.GetCost().GetOrZero(); ok {
			usage.Cost = CostFromUSD(amount)
		}
	}

	if content.Type == components.ChatAssistantMessageContentTypeStr && content.Str != nil {
		return *content.Str, usage, nil
	}

	return "", usage, ErrNoCompletion
}

// Stream sends messages and calls onDelta with each piece of the reply as it arrives.
func (c *ChatClient) Stream(ctx context.Context, messages []Message, onDelta func(string)) (string, Usage, error) {
	items := make([]components.ChatMessages, 0, len(messages))

	for _, message := range messages {
		var item components.ChatMessages
		switch message.Role {
		case "system":
			item = components.CreateChatMessagesSystem(components.ChatSystemMessage{
				Role:    components.ChatSystemMessageRoleSystem,
				Content: components.CreateChatSystemMessageContentStr(message.Content),
			})
		case "assistant":
			assistantContent := components.CreateChatAssistantMessageContentStr(message.Content)
			item = components.CreateChatMessagesAssistant(components.ChatAssistantMessage{
				Role:    components.ChatAssistantMessageRoleAssistant,
				Content: optionalnullable.From(&assistantContent),
			})
		case "user", "":
			item = components.CreateChatMessagesUser(components.ChatUserMessage{
				Role:    components.ChatUserMessageRoleUser,
				Content: components.CreateChatUserMessageContentStr(message.Content),
			})
		default:
			return "", Usage{}, fmt.Errorf("ai chat: unsupported role %q", message.Role)
		}

		items = append(items, item)
	}

	response, err := c.api.Chat.Send(ctx, components.ChatRequest{
		Model:    orsdk.Pointer(c.model),
		Messages: items,
		Stream:   orsdk.Pointer(true),
	}, nil)
	if err != nil {
		return "", Usage{}, fmt.Errorf("ai chat: %w", err)
	}

	events := response.EventStream
	defer events.Close()

	var reply strings.Builder

	for events.Next() {
		chunk := events.Value().Data
		if chunk.Error != nil {
			return "", Usage{}, fmt.Errorf("ai chat: %s", chunk.Error.Message)
		}

		for _, choice := range chunk.Choices {
			delta, _ := choice.Delta.Content.GetOrZero()
			if delta != "" {
				reply.WriteString(delta)
				onDelta(delta)
			}
		}
	}

	if err := events.Err(); err != nil {
		return "", Usage{}, fmt.Errorf("ai chat: %w", err)
	}

	if reply.Len() == 0 {
		return "", Usage{}, ErrNoCompletion
	}

	return reply.String(), Usage{}, nil
}

package ai_models

import (
	"context"
	"fmt"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

type DoubaoAI struct {
	ModelKey string
	ApiKey   string
	BaseUrl  string
}

func NewDoubaoAI(modelKey, apiKey string) *DoubaoAI {
	return &DoubaoAI{
		ModelKey: modelKey,
		ApiKey:   apiKey,
		BaseUrl:  "https://ark.cn-beijing.volces.com/api/v3",
	}
}

func (d *DoubaoAI) GetAiResponse(ctx context.Context, systemRolePrompt, userRolePrompt string) (string, error) {
	client := arkruntime.NewClientWithApiKey(d.ApiKey,
		arkruntime.WithBaseUrl(d.BaseUrl),
	)
	req := model.ChatCompletionRequest{
		Model: d.ModelKey,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(systemRolePrompt),
				},
			},
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(userRolePrompt),
				},
			},
		},
		Temperature: 0.3,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("standard chat error: %v", err)
	}
	resStrPtr := resp.Choices[0].Message.Content.StringValue
	if resStrPtr == nil {
		return "", fmt.Errorf("standard chat error: %v", err)
	}
	return *resp.Choices[0].Message.Content.StringValue, nil
}

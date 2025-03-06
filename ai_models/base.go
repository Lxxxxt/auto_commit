package ai_models

import (
	"context"
	"log"
	"strings"
)

type AI interface {
	GetAiResponse(ctx context.Context, systemRolePrompt, userRolePrompt string) (string, error)
}

func GetAiModel(modelKey, apiKey, platform string) AI {
	switch {
	case strings.Contains(strings.ToLower(platform), "doubao"):
		return NewDoubaoAI(modelKey, apiKey)
	case strings.Contains(strings.ToLower(platform), "kimi"):
		return NewKimiAI(modelKey, apiKey)
	default:
		log.Fatalf("不支持的ai平台类型: %s", platform)
		return nil
	}
}

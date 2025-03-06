package ai_models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tidwall/gjson"
)

type KimiAI struct {
	ModelKey string
	ApiKey   string
	baseUrl  string
}

func NewKimiAI(modelKey, apiKey string) *KimiAI {
	return &KimiAI{
		ModelKey: modelKey,
		ApiKey:   apiKey,
		baseUrl:  "https://api.moonshot.cn/v1/chat/completions",
	}
}

// moonshot-v1-8k
func (k *KimiAI) GetAiResponse(ctx context.Context, systemRolePrompt, userRolePrompt string) (string, error) {
	reqBody := map[string]interface{}{}
	reqBody["model"] = k.ModelKey
	reqBody["messages"] = []map[string]string{
		{
			"role":    "system",
			"content": systemRolePrompt,
		},
		{
			"role":    "user",
			"content": userRolePrompt,
		},
	}
	reqBody["temperature"] = 0.3

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal(fmt.Sprintf("json marshal 失败:%s", err.Error()))
	}
	// 创建请求体的字节缓冲区
	requestBodyBuffer := bytes.NewBuffer(jsonData)

	// 构建HTTP请求
	//
	req, err := http.NewRequest(http.MethodPost, k.baseUrl, requestBodyBuffer)
	if err != nil {
		log.Fatal(fmt.Sprintf("创建请求体失败:%s", err.Error()))

	}
	// 设置请求头
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", k.ApiKey))

	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(fmt.Sprintf("发送请求失败:%s", err.Error()))
	}
	defer resp.Body.Close()

	// 读取响应数据
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(fmt.Sprintf("读取响应数据失败:%s", err.Error()))
	}
	content := gjson.Get(string(respBody), "choices.0.message.content")
	if content.String() == "" {
		log.Fatal(fmt.Sprintf("提取kimi resp失败:%s", string(respBody)))
	}
	// 打印响应数据
	return content.String(), nil

}

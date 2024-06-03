package main

import (
	"autoc/log"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"net/http"
	"os/exec"

	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
)

const systemRolePrompt = "你是一名资深的golang开发者和代码审核员,并有着对自己的git commit记录的简洁性和准确性有着近乎疯狂的追求.你会根据用户发送的git diff内容,为用户生成一段简洁的,准确的,概括性的,不超过20个中文文字的commit message,帮助用户在提交commit时归档改动内容.你的回答需要满足特定格式,格式为{\"response\":\"<你的答案>\"}."
const userRolePrompt = "我的git diff的内容:{%s}, 请你帮我生成commit message"

var kimiKey string

var cmd = struct {
	customMessage *string
	whetherPush   *bool
	setKey        *string
}{}

func init() {
	cmd.customMessage = pflag.StringP("message", "m", "", "commit message")
	cmd.whetherPush = pflag.BoolP("push", "p", false, "push to remote")
	cmd.setKey = pflag.StringP("set-key", "k", "", "set kimi key")
	pflag.Parse()
}

func main() {
	if cmd.setKey != nil && *cmd.setKey != "" {
		setKimiKey(*cmd.setKey)
		return
	}
	kimiKey = getKimiKey()
	if kimiKey == "" {
		log.Fatal("未配置kimi key")
	}
	var commitMessage string
	if cmd.customMessage != nil && *cmd.customMessage != "" {
		commitMessage = *cmd.customMessage
	} else {
		commitMessage = getCommitMessage()
	}
	gitCommit(commitMessage)
	if cmd.whetherPush != nil && *cmd.whetherPush {
		gitPush()
	}

}
func setKimiKey(key string) {
	fullPath := getConfigPath()
	// 文件存在，打开文件进行后续操作
	file, err := os.OpenFile(fullPath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 这里可以添加代码来操作文件，比如读取或写入内容
	// 例如，写入文件内容
	_, err = file.WriteString(key)
	if err != nil {
		log.Fatal(err)
	}

	// 保存并关闭文件
	err = file.Sync()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("kimi key 已设置")

}
func getConfigPath() (fullPath string) {
	// 定义文件路径
	path := filepath.Join(".autoc", "config.json")

	// 获取当前用户的主目录路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// 构建完整的文件路径
	fullPath = filepath.Join(homeDir, path)

	// 检查目录是否存在，如果不存在则创建
	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 检查文件是否存在，如果不存在则创建
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		file, err := os.Create(fullPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}
	return
}

func getKimiKey() string {
	fullPath := getConfigPath()
	// 文件存在，打开文件进行后续操作
	file, err := os.OpenFile(fullPath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 这里可以添加代码来操作文件，比如读取或写入内容
	// 例如，读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

// 执行git diff ':!kitex_gen' 命令并保存输出
func gitDiff() string {
	cmd := exec.Command("git", "diff", "--", ":(exclude)kitex_gen/*", ":(exclude)go.sum")
	gitDiffOut, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(fmt.Sprintf("git diff 失败:%s", err.Error()))

	}
	gitDiffOutput := string(gitDiffOut)
	if len(gitDiffOutput) > 10000 {
		log.Fatal("git diff 输出过长")
	}
	return gitDiffOutput
}
func gitCommit(commitMessage string) {
	// 添加所有更改的文件并提交
	addCmd := exec.Command("git", "add", ".")
	// 执行git add
	err := addCmd.Run()
	if err != nil {
		log.Fatal(fmt.Sprintf("git add 失败:%s", err.Error()))
	}
	commitCmd := exec.Command("git", "commit", "-a", "-m", commitMessage)
	// 执行git commit
	err = commitCmd.Run()
	if err != nil {
		log.Fatal(fmt.Sprintf("git commit 失败:%s", err.Error()))
	}
}
func gitPush() {
	// git push (如果需要的话)
	pushCmd := exec.Command("git", "push")
	err := pushCmd.Run()
	if err != nil {
		log.Fatal(fmt.Sprintf("git push 失败:%s", err.Error()))
	}
}
func getAiResponse(gitDiff string) string {
	reqBody := map[string]interface{}{}
	reqBody["model"] = "moonshot-v1-8k"
	reqBody["messages"] = []map[string]string{
		{
			"role":    "system",
			"content": systemRolePrompt,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf(userRolePrompt, gitDiff),
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
	req, err := http.NewRequest(http.MethodPost, "https://api.moonshot.cn/v1/chat/completions", requestBodyBuffer)
	if err != nil {
		log.Fatal(fmt.Sprintf("创建请求体失败:%s", err.Error()))

	}
	// 设置请求头
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", kimiKey))

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
	// 打印响应数据
	return string(respBody)

}
func getCommitMessage() string {
	gitDiff := gitDiff()
	if gitDiff == "" {
		return "init commit"
	}
	response := getAiResponse(gitDiff)
	commitMessage := extractCommitMessage(response)
	return commitMessage

}

// extractCommitMessage 从Kimi的回答中提取commit message
func extractCommitMessage(response string) string {
	content := gjson.Get(response, "choices.0.message.content")
	if content.String() == "" {
		log.Fatal(fmt.Sprintf("提取kimi resp失败:%s", response))
	}
	commitMessage := gjson.Get(content.String(), "response")
	if commitMessage.String() == "" {
		log.Fatal(fmt.Sprintf("提取commit message失败:%s", content.String()))
	}
	return commitMessage.String()

}

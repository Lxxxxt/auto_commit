package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"os/exec"

	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

const systemRolePrompt = "你是一名资深的golang开发者和代码审核员,目前主要做的业务是小程序开放平台的研发,你有着对自己的git commit记录的简洁性和准确性有着近乎疯狂的追求.你会根据用户发送的git diff内容,为用户生成一段简洁的,准确的,概括性的,不超过20个中文文字的commit message,帮助用户在提交commit时归档改动内容.你的回答需要满足特定格式,格式为{\"response\":\"<你的答案>\"}."
const userRolePrompt = "我的git diff的内容:{%s}, 请你帮我生成commit message"

var modelKey string = "ep-20240707145511-q7dqb"
var ARK_API_KEY string 

var cmd = struct {
	customMessage *string
	whetherPush   *bool
	setKey        *string
}{}

func init() {

	cmd.customMessage = pflag.StringP("message", "m", "", "commit message")
	cmd.whetherPush = pflag.BoolP("push", "p", false, "push to remote")
	cmd.setKey = pflag.StringP("set-key", "k", "", "set api key")
	pflag.Parse()
}

func main() {
	if cmd.setKey != nil && *cmd.setKey != "" {
		setApiKey(*cmd.setKey)
		return
	}
	ARK_API_KEY = getApiKey()
	if ARK_API_KEY == "" {
		log.Fatal("未配置api key")
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
func setApiKey(key string) {
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
	fmt.Println("api key 已设置")

}
func getConfigPath() (fullPath string) {
	// 定义文件路径
	path := filepath.Join(".autoc", "autoc_config.json")

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

func getApiKey() string {
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
		log.Fatalf("git diff 失败:%s", err.Error())

	}
	gitDiffOutput := string(gitDiffOut)
	if len(gitDiffOutput) > 30000 {
		log.Fatal("git diff 输出过长")
	}
	fmt.Println("git diff 输出:", gitDiffOutput)
	fmt.Println("git diff 输出长度:", len(gitDiffOutput))
	return gitDiffOutput
}
func gitCommit(commitMessage string) {
	// 添加所有更改的文件并提交
	addCmd := exec.Command("git", "add", ".")
	// 执行git add
	err := addCmd.Run()
	if err != nil {
		log.Fatalf("git add 失败:%s", err.Error())
	}
	commitCmd := exec.Command("git", "commit", "-a", "-m", commitMessage)
	// 执行git commit
	err = commitCmd.Run()
	if err != nil {
		log.Fatalf("git commit 失败:%s", err.Error())
	}
}
func gitPush() {
	// 检查远端是否存在这个分支 git branch -r |grep $(git_current_branch)
	branchCmd := exec.Command("git", "branch", "--remote", "|", "grep", "$(git_current_branch)")
	branchOut, err := branchCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("git branch 失败:%s", err.Error())
	}
	branchOutStr := string(branchOut)
	//，如果不存在执行 git push --set-upstream origin $(git_current_branch)
	if strings.Contains(branchOutStr, "origin/master") {
		pushCmd := exec.Command("git", "push")
		err := pushCmd.Run()
		if err != nil {
			log.Fatalf("git push 失败:%s", err.Error())
		}
	} else {
		pushCmd := exec.Command("git", "push", "--set-upstream", "origin", "master")
		err := pushCmd.Run()
		if err != nil {
			log.Fatalf("git push 失败:%s", err.Error())
		}
	}
}

func getAiResponseDoubao(ctx context.Context, gitDiff string) (string, error) {
	client := arkruntime.NewClientWithApiKey(ARK_API_KEY,
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)
	req := model.ChatCompletionRequest{
		Model: modelKey,
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
					StringValue: volcengine.String(fmt.Sprintf(userRolePrompt, gitDiff)),
				},
			},
		},
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

func getCommitMessage() string {
	gitDiff := gitDiff()
	if gitDiff == "" {
		return "init commit"
	}
	aiResp, err := getAiResponseDoubao(context.Background(), gitDiff)
	if err != nil {
		log.Fatalf("获取commit message失败:%s", err.Error())
	}
	return extractCommitMessage(aiResp)

}

// extractCommitMessage 从Kimi的回答中提取commit message
func extractCommitMessage(response string) string {
	commitMessage := gjson.Get(response, "response")
	if commitMessage.String() == "" {
		log.Fatalf("提取commit message失败:%s", response)
	}
	return commitMessage.String()

}

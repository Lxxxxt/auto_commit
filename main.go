package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"os/exec"

	"github.com/Lxxxxt/auto_commit/ai_models"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
)

const systemRolePrompt = "你是一名资深的golang开发者和代码审核员,目前主要做的业务是小程序开放平台的研发,你有着对自己的git commit记录的简洁性和准确性有着近乎疯狂的追求.你会根据用户发送的git diff内容,为用户生成一段简洁的,准确的,概括性的,不超过20个中文文字的commit message,帮助用户在提交commit时归档改动内容.你的回答需要满足特定格式,格式为{\"response\":\"<你的答案>\"}."
const userRolePrompt = "我的git diff的内容:{%s}, 请你帮我生成commit message"

var cmd = struct {
	customMessage *string
	whetherPush   *bool
	setKey        *string
	setPlatform   *string
	setModelKey   *string
	getVersion    *bool
}{}

func init() {

	cmd.customMessage = pflag.StringP("message", "m", "", "commit message")
	cmd.whetherPush = pflag.BoolP("push", "p", false, "push to remote")
	cmd.setKey = pflag.StringP("set-key", "", "", "set api key")
	cmd.setPlatform = pflag.StringP("set-platform", "", "", "set platform")
	cmd.setModelKey = pflag.StringP("set-model-key", "", "", "set model key")
	cmd.getVersion = pflag.BoolP("version", "v", false, "get version")
	pflag.Parse()
}

func main() {
	if cmd.getVersion != nil && *cmd.getVersion {
		fmt.Println("version: 1.0.0 2025-03-05")
		return
	}
	if cmd.setKey != nil && *cmd.setKey != "" {
		setApiKey(*cmd.setKey)
		return
	}
	if cmd.setPlatform != nil && *cmd.setPlatform != "" {
		setPlatform(*cmd.setPlatform)
		return
	}
	if cmd.setModelKey != nil && *cmd.setModelKey != "" {
		setModelKey(*cmd.setModelKey)
		return
	}
	apiKey := getApiKey()
	if apiKey == "" {
		log.Fatal("未配置api key")
	}
	modelKey := getModelKey()
	if modelKey == "" {
		log.Fatal("未配置model key")
	}
	platform := getPlatform()
	if platform == "" {
		log.Fatal("未配置平台")
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

func getCommitMessage() string {
	gitDiff := gitDiff()
	if gitDiff == "" {
		return "init commit"
	}
	aiResp, err := ai_models.GetAiModel(getModelKey(), getApiKey(), getPlatform()).GetAiResponse(context.Background(), systemRolePrompt, userRolePrompt)
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

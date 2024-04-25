package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

func getDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	diff, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot get diff: %v", err)
	}
	if len(diff) == 0 {
		return "", errors.New("empty diff")
	}
	return string(diff), nil
}

func fetch(url string, prompt string) (string, error) {
	safePrompt, _ := json.Marshal(prompt)
	body := strings.NewReader(fmt.Sprintf(`
	{
		"messages": [
			{
				"content": %s,
				"role": "user"
			}
		],
		"id": "",
		"previewToken": null,
		"userId": "",
		"codeModelMode": true,
		"agentMode": {},
		"trendingAgentMode": {},
		"isMicMode": false
	}
	`, safePrompt))
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", fmt.Errorf("error creating a request: %v", err)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Referer", "https://www.blackbox.ai/")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "https://www.blackbox.ai")
	req.Header.Add("Alt-Used", "www.blackbox.ai")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error with network request: statuscode: %v %v", res.StatusCode, err)
	}
	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}
	defer res.Body.Close()
	return string(resBytes), nil
}

func getCommitMsg(diff string) (string, error) {
	commitmsg, err := fetch("https://www.blackbox.ai/api/chat", fmt.Sprintf("For the following git diff, give me a good commit message (only reply with the commit message, dont say anything else):\n%s", diff))
	if err != nil {
		return "", fmt.Errorf("cannot fetch commit message: %v", err)
	}
	return commitmsg, nil
}

func main() {
	diff, err := getDiff()
	var commitmsg string
	if err != nil {
		commitmsg = fmt.Sprintf("CommitBuddy Failed 😕 (empty diff): %v", err)
	} else {
		commitmsg, err = getCommitMsg(diff)
		if err != nil {
			commitmsg = fmt.Sprintf("CommitBuddy Failed 😕 (could not generate a commit message): %v", err)
		}
	}
	fmt.Println(commitmsg)
}

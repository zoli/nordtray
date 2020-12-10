package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func execCmd(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.Output()
	res := cleanCliOutPut(string(out))
	if err != nil {
		if strings.Contains(res, noNetErrStr) {
			return "", ErrNoNet
		}
		return "", fmt.Errorf("non zero exit code: %s: %s", err, res)
	}
	if ctx.Err() == context.DeadlineExceeded {
		return "", err
	}
	return res, nil
}

func cleanCliOutPut(out string) string {
	r := strings.NewReplacer("-\r", "", "|\r", "", "/\r", "", "\\\r", "")
	return strings.ReplaceAll(r.Replace(out), "\r  \r", "\r")
}

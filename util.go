package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func execCmd(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("non zero exit code: %s", err)
	}
	if ctx.Err() == context.DeadlineExceeded {
		return "", err
	}
	return string(out), nil
}

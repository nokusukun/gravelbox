package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("No .execute path specified")
		os.Exit(1)
	}
	execPath := os.Args[1]
	execDir := path.Dir(execPath)

	execS, err := ioutil.ReadFile(execPath)
	if err != nil {
		fmt.Println("Cannot read .execute")
		os.Exit(1)
	}

	err = os.Chdir(execDir)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to change directory %v", execDir))
		os.Exit(1)
	}

	executor := ExecuteFile{}
	err = json.Unmarshal(execS, &executor)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to parse .execute: %v", err))
		os.Exit(1)
	}

	for _, command := range executor.Commands {
		var timeout time.Duration
		if command.Timeout != "" {
			var err error
			timeout, err = time.ParseDuration(command.Timeout)
			if err != nil {
				fmt.Println(fmt.Sprintf("Failed to parse timeout duration: %v", command.Timeout))
				os.Exit(1)
			}
		}

		var ctx context.Context
		var cancel context.CancelFunc
		if timeout != 0 {
			ctx, cancel = context.WithTimeout(context.Background(), timeout)
			defer cancel()
		} else {
			ctx = context.Background()
		}
		fmt.Println("---.executor---")
		cmd := exec.CommandContext(ctx, command.Command, command.Args...)
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			switch ctx.Err() {
			case context.DeadlineExceeded:
				err = fmt.Errorf("script execution timeout")
			}
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Timeout string   `json:"timeout"`
}

type ExecuteFile struct {
	Commands []Command `json:"commands"`
}

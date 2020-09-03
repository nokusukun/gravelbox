package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
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

	exports := []Export{}

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
		if !executor.NoParse && !executor.ExportJSON {
			fmt.Println("---.executor---")
		}
		cmd := exec.CommandContext(ctx, command.Command, command.Args...)

		if !executor.ExportJSON {
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
		}

		if command.Env != nil {
			cmd.Env = append(os.Environ(), command.Env...)
		}
		var err error
		var outputB []byte
		var output string

		if executor.ExportJSON {
			outputB, err = cmd.CombinedOutput()
			output = string(outputB)
		} else {
			err = cmd.Run()
		}

		if err != nil {
			switch ctx.Err() {
			case context.DeadlineExceeded:
				err = fmt.Errorf("script execution timeout")
				output = fmt.Sprintf("%v (Error: script execution timeout)", output)
			}
			if !executor.ExportJSON {
				fmt.Println(err)
			}
		}
		exports = append(exports, Export{
			Command: command,
			Output:  strings.TrimSpace(output),
		})
	}
	if executor.ExportJSON {
		out, err := json.Marshal(exports)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ExecutorJSON:%v", string(out))
	}
}

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Timeout string   `json:"timeout"`
	Env     []string `json:"env"`
}

type ExecuteFile struct {
	Commands   []Command `json:"commands"`
	NoParse    bool      `json:"no_parse"`
	ExportJSON bool      `json:"export_json"`
}

type Export struct {
	Command Command `json:"command"`
	Output  string  `json:"output"`
}

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/nokusukun/stemp"
	"github.com/rs/xid"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func GetDockerVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cfgDocker.Key("command").String(), "--version")
	x, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(x)), nil
}

func BuildAtom(name string) (string, error) {
	// docker build --tag atom-2 .
	atomName := fmt.Sprintf("atom-%v", name)
	atomSource := path.Join(".", "atom", ".")

	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		cfgDocker.Key("command").String(), "build",
		"--tag", atomName,
		atomSource)
	x, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug(cmd.Args)
		log.Errorf("Failed to run command: %v\n%v", err, strings.TrimSpace(string(x)))
		return "", err
	}

	log.Debug(string(x))
	return strings.TrimSpace(string(x)), nil
}

type Atom struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageID    string `json:"image_id"`
	Created    string `json:"created"`
	Size       string `json:"size"`
}

func ListAtoms() ([]Atom, error) {
	// docker images --filter "label=source=gravelbox"
	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		cfgDocker.Key("command").String(), "images",
		"--filter", "label=source=gravelbox",
		"--format", "{{.Repository}}@@{{.Tag}}@@{{.ID}}@@{{.CreatedAt}}@@{{.Size}}")
	x, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug(cmd.Args)
		log.Errorf("Failed to run command: ", err)
	}

	var atoms []Atom
	lines := strings.TrimSpace(string(x))
	for _, line := range strings.Split(lines, "\n") {
		elems := strings.Split(line, "@@")
		atoms = append(atoms, Atom{
			Repository: elems[0],
			Tag:        elems[1],
			ImageID:    elems[2],
			Created:    elems[3],
			Size:       elems[4],
		})
	}

	return atoms, nil
}

type Binary struct {
	// Base64 encoded data
	Data string `json:"data" binding:"required"`
	// Name of the fine
	Name string `json:"name" binding:"required"`
	// Resolve the {path} inside of the code
	Resolve bool `json:"resolve"`
}

type Executor struct {
	// Files to send to the sandbox
	Binaries []Binary `json:"binaries"`

	// Entry point
	Command []string `json:"command" binding:"required"`
	// Script execution timeout
	Timeout  string `json:"timeout" binding:"required"`
	Atom     string `json:"atom" binding:"required"`
	Network  bool   `json:"network"`
	ReadOnly bool   `json:"read_only"`

	sbx string
}

func (e *Executor) RuntineID() string {
	return e.sbx
}

func (e *Executor) Start() (string, error) {
	mountdir := cfg.Section("gravelbox").Key("mountdir").String()

	timeout, err := time.ParseDuration(e.Timeout)
	if err != nil {
		return "", fmt.Errorf("%v is not a valid timeout duration", e.Timeout)
	}

	// Sandbox path
	workingpath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}

	sbxfolder := xid.New().String()
	e.sbx = sbxfolder
	mountpath := path.Join(workingpath, mountdir)
	sbxpath := path.Join(mountpath, sbxfolder)
	err = os.MkdirAll(sbxpath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create sandbox directory '%v': %v", sbxpath, err)
	}

	for _, b := range e.Binaries {
		// Save to payload
		binary, err := base64.StdEncoding.DecodeString(b.Data)
		if err != nil {
			return "", fmt.Errorf("cannot parse binary: %v", err)
		}

		if b.Resolve {
			binary = []byte(stemp.Compile(string(binary), map[string]interface{}{
				"path": fmt.Sprintf("/mnt/%v", sbxfolder),
			}))
		}

		binaryPath := path.Join(sbxpath, b.Name)
		err = ioutil.WriteFile(binaryPath, binary, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to write binary: %v", err)
		}

	}

	// docker run --rm --network none -it -v absolute_path:/mnt atom-2 python3 /mnt/sample.py
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	arguments := []string{"run", "--rm"}
	if e.ReadOnly {
		arguments = append(arguments, "--read-only")
	}
	if !e.Network {
		arguments = append(arguments, "--network", "none")
	}

	arguments = append(arguments,
		"-v", fmt.Sprintf("%v:/mnt", mountpath),
		e.Atom,)

	// replace format strings in command
	for i := range e.Command {
		//if strings.Contains(e.Command[i], "%v") {
		//	e.Command[i] = fmt.Sprintf(e.Command[i], sbxfolder)
		//}
		e.Command[i] = stemp.Compile(e.Command[i], map[string]interface{}{
			"path": fmt.Sprintf("/mnt/%v", sbxfolder),
		})
	}

	arguments = append(arguments, e.Command...)
	log.Verbose("Executing command", arguments)

	cmd := exec.CommandContext(ctx, cfgDocker.Key("command").String(), arguments...)
	x, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug(cmd.Args, ctx.Err())
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return "", fmt.Errorf("script execution timeout")
		}
		return strings.TrimSpace(string(x)), err
	}

	return strings.TrimSpace(string(x)), nil
}

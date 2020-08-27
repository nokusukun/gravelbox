package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	// docker image rm atom-marcus
	atomName := fmt.Sprintf("atom-%v", name)
	atomSource := path.Join(".", cfg.Section("atom").Key("path").String(), ".")

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
		err = fmt.Errorf("%v: %v", strings.Split(string(x), "\n")[0], err)
	}
	return strings.TrimSpace(string(x)), err
}

func DeleteAtom(name string) error {
	// docker build --tag atom-2 .
	// docker image rm atom-marcus
	atomName := fmt.Sprintf("atom-%v", name)

	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		cfgDocker.Key("command").String(), "image",
		"rm", atomName)
	x, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug(cmd.Args)
		log.Errorf("Failed to run command: %v\n%v", err, strings.TrimSpace(string(x)))
	}
	return err
}

type Atom struct {
	Name       string `json:"name"`
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
		if len(elems) == 5 {
			atoms = append(atoms, Atom{
				Name:       strings.ReplaceAll(elems[0], "atom-", ""),
				Repository: elems[0],
				Tag:        elems[1],
				ImageID:    elems[2],
				Created:    elems[3],
				Size:       elems[4],
			})
		} else {
			return atoms, fmt.Errorf("no atoms detected")
		}

	}

	return atoms, nil
}

type Binary struct {
	// data string, when sending payloads, encode to base64 and set the `decode_b64` flag option to true.
	Data interface{} `json:"data" binding:"required"`
	// Name of the fine
	Name string `json:"name" binding:"required"`
	// Resolve the {path} inside of the code
	Resolve bool `json:"resolve"`
	// DecodeB64 treat the string as a base64 string
	DecodeB64  bool `json:"decode_b64"`
	DecodeJSON bool `json:"decode_json"`
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

/*
mcs -out:{path}/test.exe {path}/test.cs
mono {path}/test.exe
*/

func (e *Executor) RuntineID() string {
	return e.sbx
}

func (e *Executor) Start() (string, error) {
	start := time.Now()
	defer func() {
		log.Verbose("Execution time: ", time.Now().Sub(start))
	}()

	mountdir := cfg.Section("gravelbox").Key("mountdir").String()

	timeout, err := time.ParseDuration(e.Timeout)
	if err != nil {
		return "", fmt.Errorf("%v is not a valid timeout duration", e.Timeout)
	}

	// Sandbox path
	workingPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}

	sbxFolder := xid.New().String()
	e.sbx = sbxFolder
	mountPath := path.Join(workingPath, mountdir)
	sbxPath := path.Join(mountPath, sbxFolder)
	err = os.MkdirAll(sbxPath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create sandbox directory '%v': %v", sbxPath, err)
	}

	resolves := map[string]interface{}{
		"path":    fmt.Sprintf("/mnt/%v", sbxFolder),
		"runtime": sbxFolder,
	}
	for _, b := range e.Binaries {
		// Save to payload
		decoded := false
		var binary []byte
		if b.DecodeB64 {
			binary, err = base64.StdEncoding.DecodeString(b.Data.(string))
			if err != nil {
				return "", fmt.Errorf("cannot parse binary: %v", err)
			}
			decoded = true
		}

		if b.DecodeJSON {
			binary, err = json.Marshal(b.Data)
			if err != nil {
				return "", fmt.Errorf("cannot parse binary: %v", err)
			}
			decoded = true
		}

		if !decoded {
			binary = []byte(b.Data.(string))
		}

		if b.Resolve {
			binary = []byte(stemp.Compile(string(binary), resolves))
		}

		binaryPath := path.Join(sbxPath, b.Name)
		err = ioutil.WriteFile(binaryPath, binary, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to write binary: %v", err)
		}

	}

	log.Verbose("Binary Saving: ", time.Now().Sub(start))

	// docker run --rm --network none -it -v absolute_path:/mnt atom-2 python3 /mnt/sample.py
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	arguments := []string{"run", "--rm", "--name", sbxFolder, "--workdir", fmt.Sprintf("/mnt/%v", sbxFolder)}
	if e.ReadOnly {
		arguments = append(arguments, "--read-only")
	}
	if !e.Network {
		arguments = append(arguments, "--network", "none")
	}

	arguments = append(arguments,
		"-v", fmt.Sprintf("%v:/mnt", mountPath),
		fmt.Sprintf("atom-%v", e.Atom))

	// replace format strings in command
	for i := range e.Command {
		e.Command[i] = stemp.Compile(e.Command[i], resolves)
	}

	arguments = append(arguments, e.Command...)
	log.Verbose("Executing command", arguments)
	cmd := exec.CommandContext(ctx, cfgDocker.Key("command").String(), arguments...)
	log.Verbose("Waiting for output: ", time.Now().Sub(start))

	x, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug(cmd.Args, ctx.Err())
		switch ctx.Err() {
		case context.DeadlineExceeded:
			go func() {
				err := KillAtomContainer(sbxFolder)
				if err != nil {
					log.Errorf("Failed to force kill a container: %v", err)
				}
			}()
			return "", fmt.Errorf("script execution timeout")
		}
		msg := string(x)
		log.Debug("Sending more info error")

		return strings.TrimSpace(msg), err
	}

	return strings.TrimSpace(string(x)), nil
}

func KillAtomContainer(name string) error {

	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		cfgDocker.Key("command").String(), "rm", "-f", name)
	x, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug(cmd.Args)
		log.Errorf("Failed to run command: %v\n%v", err, strings.TrimSpace(string(x)))
	}
	return err
}

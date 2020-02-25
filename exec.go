package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// errEmptyCommand identifies an error returned when a command is an invalid.
// The user should check the command input for errors before continuing.
var errInvalidCommand = errors.New("command is invalid")

// build runs the supplied command. This function will return after the command
// has finished. An error will be returned of the command failed.
func build(ctx context.Context, command string) error {
	if command == "" {
		return errInvalidCommand
	}

	fmt.Println("Building:", command)
	return shell(ctx, command).Run()
}

// run starts the supplied command. An error is returned if the operation fails.
func run(ctx context.Context, command string) (*exec.Cmd, error) {
	if command == "" {
		return nil, errInvalidCommand
	}

	fmt.Println("Running:", command)
	cmd := shell(ctx, command)
	return cmd, cmd.Start()
}

// shell takes a command string and splits it into a set of operators that is
// usable to the exec package, returning an exec.Cmd instance.
//
// It will attempt to pass the command to the bash/bourne shell if possible.
// Otherwise it will attempt to execute the command directly.
func shell(ctx context.Context, command string) *exec.Cmd {
	var cmd *exec.Cmd
	for _, t := range []string{"bash", "sh"} {
		sh, err := exec.LookPath(t)
		if err == nil {
			cmd = exec.CommandContext(ctx, sh, "-c", command)
			break
		}
	}
	if cmd == nil {
		cmd = &exec.Cmd{}
		args := strings.Split(command, " ")
		if len(args) > 0 {
			cmd = exec.CommandContext(ctx, args[0], args[1:]...)
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

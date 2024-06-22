package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/anton-martsiushkou/github_actions_multirun/commands"
	"github.com/anton-martsiushkou/github_actions_multirun/logger"
)

const (
	helpCommand = "help"
	toolName    = "multirun"
)

func init() {
	commands.RegisterCommand("run", commands.NewRunCommand())
}

func main() {
	if len(os.Args) < 2 {
		if err := commands.Usage(os.Stderr); err != nil {
			die(1, err.Error())
		}
		os.Exit(2)
	}

	flag.CommandLine.Usage = func() {
		die(2, "Run '%s help' for usage.", toolName)
	}

	flag.Parse()
	args := flag.Args()

	// help
	if args[0] == helpCommand {
		if err := help(args[1:], os.Stdout); err != nil {
			die(2, err.Error())
		}
	}

	l := logger.New()
	ctx := context.Background()
	logger.ToContext(ctx, l)

	// run command
	command := commands.GetCommand(args[0])
	if command == nil {
		die(2, "%s: unknown subcommand %q\nRun '%s help' for usage.", toolName, args[0], toolName)
	}

	if err := command.Run(ctx, args[1:], os.Stdout); err != nil {
		die(1, err.Error())
	}
}

func die(exitCode int, format string, args ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, format+"\n", args...); err != nil {
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func help(args []string, w io.Writer) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: %s help [command]\n\nToo many arguments given", toolName)
	}

	if len(args) == 0 {
		return commands.Usage(w)
	}

	command := commands.GetCommand(args[0])
	if command == nil {
		return fmt.Errorf(fmt.Sprintf("%s: unknown command %q\nRun 'multirun help' for usage", toolName, args[0]))
	}

	if _, err := fmt.Fprintf(w, "Usage: %s %s %s\n", toolName, args[0], command.UsageLine()); err != nil {
		return err
	}

	if fs := command.FlagSet(); fs != nil {
		fs.SetOutput(w)
		fs.PrintDefaults()
	}

	return command.HelpMessage(w)
}

package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"

	"github.com/anton-martsiushkou/github_actions_multirun/github"
	"github.com/anton-martsiushkou/github_actions_multirun/logger"
	"github.com/anton-martsiushkou/github_actions_multirun/multirun"
)

const branchParam = "branch"

type RunCommand struct {
	BaseCommand

	params paramsFlag
}

type paramsFlag map[string]interface{}

func (p *paramsFlag) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p *paramsFlag) Set(value string) error {
	kvPairs := strings.Split(value, ",")
	*p = make(map[string]interface{}, len(kvPairs))

	for _, kv := range kvPairs {
		split := strings.Split(kv, "=")
		if len(split) != 2 {
			return fmt.Errorf("invalid key-value pair: %s", kv)
		}

		(*p)[strings.TrimSpace(split[0])] = strings.TrimSpace(split[1])
	}
	return nil
}

func NewRunCommand() *RunCommand {
	rc := &RunCommand{}
	fs := &flag.FlagSet{}
	fs.Var(&rc.params, "params", `key-value pairs separated by comma (e.g. key1=value1,key2=value2)`)

	rc.BaseCommand = BaseCommand{
		Short: "run all workflows from workflow_manager_config for specified branch and waits for its completion",
		Flags: fs,
	}

	return rc
}

func (r *RunCommand) Run(ctx context.Context, args []string, stdout io.Writer) error {
	if err := r.FlagSet().Parse(args); err != nil {
		return CommandLineError(err.Error())
	}

	if err := r.checkFlags(); err != nil {
		return err
	}

	gh, err := github.InitGitHub(ctx)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize github client")
	}
	ghWithLog := multirun.NewGithubWithLog(gh, logger.FromContext(ctx))

	wmc, err := multirun.ReadWorkflowManagerConfig()
	if err != nil {
		return errors.WithMessage(err, "failed to read workflow_manager_config.yaml file")
	}

	wm := multirun.NewWorkflowManager(r.params, wmc, ghWithLog)

	return wm.Run(ctx)
}

func (r *RunCommand) checkFlags() error {
	if r.params[branchParam] == "" {
		return errNoBranch
	}

	return nil
}

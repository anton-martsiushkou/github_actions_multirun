package multirun

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type WorkflowManagerConfig struct {
	WorkersCount                                  int               `yaml:"workers_count"`
	MaxFailedCheckAttempts                        int               `yaml:"max_failed_check_attempts"`
	PrintResultsInterval                          time.Duration     `yaml:"print_results_interval"`
	Workflows                                     []*WorkflowConfig `yaml:"workflows"`
	DefaultInitialCheckInterval                   time.Duration     `yaml:"default_initial_check_interval"`
	DefaultFinalCheckInterval                     time.Duration     `yaml:"default_final_check_interval"`
	DefaultAttemptsThresholdForShortCheckInterval int               `yaml:"default_attempts_threshold_for_short_check_interval"`
	TimeToWaitRunIDs                              time.Duration     `yaml:"time_to_wait_run_ids"`
	MultirunTimeout                               time.Duration     `yaml:"multirun_timeout"`
	GracefulShutdownTimeout                       time.Duration     `yaml:"graceful_shutdown_timeout"`
	CancelInterval                                time.Duration     `yaml:"cancel_interval"`
	DispatchEventType                             string            `yaml:"dispatch_event_type"`
}

type WorkflowConfig struct {
	RepoName                               string                 `yaml:"repo_name"`
	OwnerName                              string                 `yaml:"owner_name"`
	WorkflowName                           string                 `yaml:"workflow_name"`
	Alias                                  string                 `yaml:"alias"`
	InitialCheckInterval                   time.Duration          `yaml:"initial_check_interval"`
	FinalCheckInterval                     time.Duration          `yaml:"final_check_interval"`
	AttemptsThresholdForShortCheckInterval int                    `yaml:"attempts_threshold_for_short_check_interval"`
	Params                                 map[string]interface{} `yaml:"params"`
}

func ReadWorkflowManagerConfig() (*WorkflowManagerConfig, error) {
	var (
		f   []byte
		err error
	)

	f, err = os.ReadFile("./workflow_manager_config.yaml")
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read workflow_manager_config.yaml file")
	}
	var cfg *WorkflowManagerConfig
	if err = yaml.Unmarshal(f, &cfg); err != nil {
		return nil, errors.WithMessage(err, "unable to unmarshal workflow_manager_config file")
	}

	cfg.Adjust()

	if err = cfg.Validate(); err != nil {
		return nil, errors.WithMessagef(err, "workflow_manager_config is invalid")
	}

	return cfg, nil
}

func (wmc *WorkflowManagerConfig) Adjust() {
	for _, wf := range wmc.Workflows {
		if wf.InitialCheckInterval == 0 {
			wf.InitialCheckInterval = wmc.DefaultInitialCheckInterval
		}
		if wf.FinalCheckInterval == 0 {
			wf.FinalCheckInterval = wmc.DefaultFinalCheckInterval
		}
		if wf.AttemptsThresholdForShortCheckInterval == 0 {
			wf.AttemptsThresholdForShortCheckInterval = wmc.DefaultAttemptsThresholdForShortCheckInterval
		}
	}
}

func (wmc *WorkflowManagerConfig) Validate() error {
	if wmc.WorkersCount == 0 {
		return errors.New("workers_count is required")
	}
	if wmc.MaxFailedCheckAttempts == 0 {
		return errors.New("max_failed_check_attempts is required")
	}
	if wmc.PrintResultsInterval == 0 {
		return errors.New("print_results_interval is required")
	}
	if len(wmc.Workflows) == 0 {
		return errors.New("workflows are required")
	}
	if wmc.DefaultInitialCheckInterval == 0 {
		return errors.New("default_initial_check_interval is required")
	}
	if wmc.DefaultFinalCheckInterval == 0 {
		return errors.New("default_final_check_interval is required")
	}
	if wmc.DefaultAttemptsThresholdForShortCheckInterval == 0 {
		return errors.New("default_attempts_threshold_for_short_check_interval is required")
	}
	if wmc.MultirunTimeout == 0 {
		return errors.New("multirun_timeout is required")
	}
	if wmc.GracefulShutdownTimeout == 0 {
		return errors.New("graceful_shutdown_timeout is required")
	}
	if wmc.CancelInterval == 0 {
		return errors.New("cancel_interval is required")
	}
	if wmc.DispatchEventType == "" {
		return errors.New("dispatch_event_type is required")
	}
	for _, wf := range wmc.Workflows {
		if wf.RepoName == "" {
			return errors.New("repo_name is required")
		}
		if wf.OwnerName == "" {
			return errors.New("owner_name is required")
		}
		if wf.WorkflowName == "" {
			return errors.New("workflow_name is required")
		}
	}
	return nil
}

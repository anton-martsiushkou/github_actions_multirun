package multirun

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowManagerConfig_Adjust(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *WorkflowManagerConfig
		inspect func(r *WorkflowManagerConfig, t *testing.T)
	}{
		{
			name: "workflow config has all values",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					Workflows: []*WorkflowConfig{
						{
							InitialCheckInterval:                   1,
							FinalCheckInterval:                     2,
							AttemptsThresholdForShortCheckInterval: 3,
							RepoName:                               "X",
						},
					},
					DefaultInitialCheckInterval:                   4,
					DefaultFinalCheckInterval:                     5,
					DefaultAttemptsThresholdForShortCheckInterval: 6,
				}
			},
			inspect: func(r *WorkflowManagerConfig, t *testing.T) {
				require.Equal(t, &WorkflowManagerConfig{
					Workflows: []*WorkflowConfig{
						{
							InitialCheckInterval:                   1,
							FinalCheckInterval:                     2,
							AttemptsThresholdForShortCheckInterval: 3,
							RepoName:                               "X",
						},
					},
					DefaultInitialCheckInterval:                   4,
					DefaultFinalCheckInterval:                     5,
					DefaultAttemptsThresholdForShortCheckInterval: 6,
				}, r)
			},
		},
		{
			name: "set default values",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					Workflows: []*WorkflowConfig{
						{
							RepoName: "X",
						},
					},
					DefaultInitialCheckInterval:                   4,
					DefaultFinalCheckInterval:                     5,
					DefaultAttemptsThresholdForShortCheckInterval: 6,
				}
			},
			inspect: func(r *WorkflowManagerConfig, t *testing.T) {
				require.Equal(t, &WorkflowManagerConfig{
					Workflows: []*WorkflowConfig{
						{
							InitialCheckInterval:                   4,
							FinalCheckInterval:                     5,
							AttemptsThresholdForShortCheckInterval: 6,
							RepoName:                               "X",
						},
					},
					DefaultInitialCheckInterval:                   4,
					DefaultFinalCheckInterval:                     5,
					DefaultAttemptsThresholdForShortCheckInterval: 6,
				}, r)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			receiver.Adjust()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func TestWorkflowManagerConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		init       func(t *testing.T) *WorkflowManagerConfig
		wantErr    bool
		inspectErr func(err error, t *testing.T)
	}{
		{
			name: "valid config",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
		},
		{
			name: "missing workers count",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "workers_count is required", err.Error())
			},
		},
		{
			name: "missing max failed check attempts",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:         1,
					PrintResultsInterval: 1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "max_failed_check_attempts is required", err.Error())
			},
		},
		{
			name: "missing print results interval",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "print_results_interval is required", err.Error())
			},
		},
		{
			name: "missing workflows",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:                                  1,
					MaxFailedCheckAttempts:                        1,
					PrintResultsInterval:                          1,
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "workflows are required", err.Error())
			},
		},
		{
			name: "missing default initial check interval",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "default_initial_check_interval is required", err.Error())
			},
		},
		{
			name: "missing default final check interval",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "default_final_check_interval is required", err.Error())
			},
		},
		{
			name: "missing default attempts threshold for short check interval",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval: 1,
					DefaultFinalCheckInterval:   1,
					MultirunTimeout:             1,
					GracefulShutdownTimeout:     1,
					CancelInterval:              1,
					DispatchEventType:           "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "default_attempts_threshold_for_short_check_interval is required", err.Error())
			},
		},
		{
			name: "missing multirun timeout",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "multirun_timeout is required", err.Error())
			},
		},
		{
			name: "missing graceful shutdown timeout",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "graceful_shutdown_timeout is required", err.Error())
			},
		},
		{
			name: "missing cancel interval",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "cancel_interval is required", err.Error())
			},
		},
		{
			name: "missing repo name",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "repo_name is required", err.Error())
			},
		},
		{
			name: "missing owner name",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "owner_name is required", err.Error())
			},
		},
		{
			name: "missing workflow name",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:  "X",
							OwnerName: "Y",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
					DispatchEventType:                             "multirun",
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "workflow_name is required", err.Error())
			},
		},
		{
			name: "missing dispatch event type",
			init: func(t *testing.T) *WorkflowManagerConfig {
				return &WorkflowManagerConfig{
					WorkersCount:           1,
					MaxFailedCheckAttempts: 1,
					PrintResultsInterval:   1,
					Workflows: []*WorkflowConfig{
						{
							RepoName:     "X",
							OwnerName:    "Y",
							WorkflowName: "Z",
						},
					},
					DefaultInitialCheckInterval:                   1,
					DefaultFinalCheckInterval:                     1,
					DefaultAttemptsThresholdForShortCheckInterval: 1,
					MultirunTimeout:                               1,
					GracefulShutdownTimeout:                       1,
					CancelInterval:                                1,
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.Equal(t, "dispatch_event_type is required", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			err := receiver.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("WorkflowManagerConfig.Validate error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

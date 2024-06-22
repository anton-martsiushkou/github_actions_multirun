package multirun

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anton-martsiushkou/github_actions_multirun/github"
)

func TestWorkflowManager_initQueue(t *testing.T) {
	tests := []struct {
		name         string
		init         func(t *testing.T) *WorkflowManager
		inspect      func(r *WorkflowManager, t *testing.T)
		commonParams map[string]interface{}
	}{
		{
			name: "default case",
			init: func(t *testing.T) *WorkflowManager {
				return &WorkflowManager{
					wmConfig: &WorkflowManagerConfig{
						Workflows: []*WorkflowConfig{
							{
								RepoName:                               "repoA",
								OwnerName:                              "ownerA",
								WorkflowName:                           "workflowA",
								FinalCheckInterval:                     time.Second,
								InitialCheckInterval:                   time.Minute,
								AttemptsThresholdForShortCheckInterval: 3,
								Alias:                                  "aliasA",
								Params: map[string]interface{}{
									"paramA": "valueA",
								},
							},
							{
								RepoName:     "repoB",
								OwnerName:    "ownerB",
								WorkflowName: "workflowB",
								Alias:        "aliasB",
								Params: map[string]interface{}{
									"paramB": "valueB",
								},
							},
							{
								RepoName:     "repoC",
								OwnerName:    "ownerC",
								WorkflowName: "workflowC",
							},
						},
					},
				}
			},
			commonParams: map[string]interface{}{
				"paramX": "valueX",
				"paramY": "valueY",
			},
			inspect: func(r *WorkflowManager, t *testing.T) {
				require.Equal(t, 3, len(r.queue))
				expWorkflowRunA := &workflowRun{
					repo:                                   "repoA",
					owner:                                  "ownerA",
					workflowName:                           "workflowA",
					initialCheckInterval:                   time.Minute,
					finalCheckInterval:                     time.Second,
					attemptsThresholdForShortCheckInterval: 3,
					alias:                                  "aliasA",
					params: map[string]interface{}{
						"paramA": "valueA",
						"paramX": "valueX",
						"paramY": "valueY",
					},
				}
				expWorkflowRunB := &workflowRun{
					repo:         "repoB",
					owner:        "ownerB",
					workflowName: "workflowB",
					alias:        "aliasB",
					params: map[string]interface{}{
						"paramB": "valueB",
						"paramX": "valueX",
						"paramY": "valueY",
					},
				}
				expWorkflowRunC := &workflowRun{
					repo:         "repoC",
					owner:        "ownerC",
					workflowName: "workflowC",
					params: map[string]interface{}{
						"paramX": "valueX",
						"paramY": "valueY",
					},
				}
				for wr := range r.queue {
					require.True(t, slices.Contains([]string{"repoA", "repoB", "repoC"}, wr.repo))
					if wr.repo == "repoA" {
						require.Equal(t, expWorkflowRunA, wr)
					}
					if wr.repo == "repoB" {
						require.Equal(t, expWorkflowRunB, wr)
					}
					if wr.repo == "repoC" {
						require.Equal(t, expWorkflowRunC, wr)
					}
				}
				require.EqualValues(t, []*workflowRun{expWorkflowRunA, expWorkflowRunB, expWorkflowRunC}, r.allWorkflows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			receiver.initQueue(tt.commonParams)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func TestWorkflowManager_launchWorkflow(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *WorkflowManager
		inspect func(r *WorkflowManager, t *testing.T)
		wr      *workflowRun
	}{
		{
			name: "golden path",
			init: func(t *testing.T) *WorkflowManager {
				ghMock := NewGithubMock(minimock.NewController(t))
				ghMock.GetLastWorkflowRunNumberMock.Expect(context.Background(), "ownerA", "repoA", "workflowA").Return(3, nil)
				ghMock.TriggerWorkflowMock.Expect(context.Background(), "ownerA", "repoA", "workflowA", "multirun", map[string]interface{}{"branch": "test"}).Return(nil)
				ghMock.GetWorkflowRunIDMock.Expect(context.Background(), "ownerA", "repoA", "workflowA", 3).Return(4, nil)
				return &WorkflowManager{
					gh: ghMock,
					wmConfig: &WorkflowManagerConfig{
						TimeToWaitRunIDs:  time.Millisecond,
						DispatchEventType: "multirun",
					},
					activeRuns: []*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
						params: map[string]interface{}{
							"branch": "test",
						},
					}},
				}
			},
			wr: &workflowRun{
				repo:         "repoA",
				owner:        "ownerA",
				workflowName: "workflowA",
				params: map[string]interface{}{
					"branch": "test",
				},
			},
			inspect: func(r *WorkflowManager, t *testing.T) {
				require.Empty(t, r.completedRuns)
				require.Equal(t, []*workflowRun{{
					repo:         "repoA",
					owner:        "ownerA",
					workflowName: "workflowA",
					params: map[string]interface{}{
						"branch": "test",
					},
				}}, r.activeRuns)
			},
		},
		{
			name: "failed to GetLastWorkflowRunNumber",
			init: func(t *testing.T) *WorkflowManager {
				ghMock := NewGithubMock(minimock.NewController(t))
				ghMock.GetLastWorkflowRunNumberMock.Expect(context.Background(), "ownerA", "repoA", "workflowA").Return(0, assert.AnError)
				return &WorkflowManager{
					gh: ghMock,
					activeRuns: []*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
					}},
				}
			},
			wr: &workflowRun{
				repo:         "repoA",
				owner:        "ownerA",
				workflowName: "workflowA",
			},
			inspect: func(r *WorkflowManager, t *testing.T) {
				require.Equal(
					t,
					[]*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
						completed:    true,
						errMsg:       "failed to get last workflow run number: assert.AnError general error for testing",
					}}, r.completedRuns)
				require.Empty(t, r.activeRuns)
			},
		},
		{
			name: "failed to TriggerWorkflow",
			init: func(t *testing.T) *WorkflowManager {
				ghMock := NewGithubMock(minimock.NewController(t))
				ghMock.GetLastWorkflowRunNumberMock.Expect(context.Background(), "ownerA", "repoA", "workflowA").Return(3, nil)
				ghMock.TriggerWorkflowMock.Expect(context.Background(), "ownerA", "repoA", "workflowA", "multirun", map[string]interface{}{"branch": "test"}).Return(assert.AnError)
				return &WorkflowManager{
					gh: ghMock,
					activeRuns: []*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
						params: map[string]interface{}{
							"branch": "test",
						},
					}},
					wmConfig: &WorkflowManagerConfig{
						DispatchEventType: "multirun",
					},
				}
			},
			wr: &workflowRun{
				repo:         "repoA",
				owner:        "ownerA",
				workflowName: "workflowA",
				params: map[string]interface{}{
					"branch": "test",
				},
			},
			inspect: func(r *WorkflowManager, t *testing.T) {
				require.Equal(
					t,
					[]*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
						completed:    true,
						params: map[string]interface{}{
							"branch": "test",
						},
						errMsg: "failed to trigger workflow: assert.AnError general error for testing",
					}}, r.completedRuns)
				require.Empty(t, r.activeRuns)
			},
		},
		{
			name: "failed to GetWorkflowRunID",
			init: func(t *testing.T) *WorkflowManager {
				ghMock := NewGithubMock(minimock.NewController(t))
				ghMock.GetLastWorkflowRunNumberMock.Expect(context.Background(), "ownerA", "repoA", "workflowA").Return(3, nil)
				ghMock.TriggerWorkflowMock.Expect(context.Background(), "ownerA", "repoA", "workflowA", "multirun", map[string]interface{}{"branch": "test"}).Return(nil)
				ghMock.GetWorkflowRunIDMock.Expect(context.Background(), "ownerA", "repoA", "workflowA", 3).Return(0, assert.AnError)
				return &WorkflowManager{
					gh: ghMock,
					wmConfig: &WorkflowManagerConfig{
						TimeToWaitRunIDs:  time.Millisecond,
						DispatchEventType: "multirun",
					},
					activeRuns: []*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
						params: map[string]interface{}{
							"branch": "test",
						},
					}},
				}
			},
			wr: &workflowRun{
				repo:         "repoA",
				owner:        "ownerA",
				workflowName: "workflowA",
				params: map[string]interface{}{
					"branch": "test",
				},
			},
			inspect: func(r *WorkflowManager, t *testing.T) {
				require.Equal(
					t,
					[]*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
						completed:    true,
						params: map[string]interface{}{
							"branch": "test",
						},
						errMsg: "failed to get workflow run ID: assert.AnError general error for testing",
					}}, r.completedRuns)
				require.Empty(t, r.activeRuns)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			receiver.launchWorkflow(context.Background(), tt.wr)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func TestWorkflowManager_addCompletedRun(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *WorkflowManager
		inspect func(r *WorkflowManager, t *testing.T)
		wr      *workflowRun
	}{
		{
			name: "default case",
			init: func(t *testing.T) *WorkflowManager {
				return &WorkflowManager{
					completedRuns: []*workflowRun{},
					activeRuns: []*workflowRun{{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
					}},
				}
			},
			wr: &workflowRun{
				repo:         "repoA",
				owner:        "ownerA",
				workflowName: "workflowA",
			},
			inspect: func(r *WorkflowManager, t *testing.T) {
				require.Equal(t, []*workflowRun{{
					repo:         "repoA",
					owner:        "ownerA",
					workflowName: "workflowA",
					completed:    true,
				}}, r.completedRuns)
				require.Empty(t, r.activeRuns)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			receiver.addCompletedRun(context.Background(), tt.wr)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func TestWorkflowManager_cancelActiveRuns(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *WorkflowManager
		inspect func(r *WorkflowManager, t *testing.T)
	}{
		{
			name: "default case",
			init: func(t *testing.T) *WorkflowManager {
				queue := make(chan *workflowRun, 1)
				queue <- &workflowRun{
					repo:         "repoC",
					owner:        "ownerC",
					workflowName: "workflowC",
				}
				close(queue)
				ghMock := NewGithubMock(minimock.NewController(t))
				ghMock.CancelWorkflowRunMock.When(context.Background(), "ownerA", "repoA", 1).Then(nil)
				ghMock.CancelWorkflowRunMock.When(context.Background(), "ownerB", "repoB", 2).Then(nil)
				ghMock.GetWorkflowRunStatusMock.When(context.Background(), "ownerA", "repoA", 1).Then(github.StatusCompleted, "cancelled", 0, nil)
				ghMock.GetWorkflowRunStatusMock.When(context.Background(), "ownerB", "repoB", 2).Then(github.StatusCompleted, "cancelled", 0, nil)
				return &WorkflowManager{
					activeRuns: []*workflowRun{
						{
							repo:         "repoA",
							owner:        "ownerA",
							workflowName: "workflowA",
							runID:        1,
						},
						{
							repo:         "repoB",
							owner:        "ownerB",
							workflowName: "workflowB",
							runID:        2,
						},
					},
					gh:    ghMock,
					queue: queue,
					wmConfig: &WorkflowManagerConfig{
						CancelInterval: time.Millisecond,
					},
				}
			},
			inspect: func(r *WorkflowManager, t *testing.T) {
				require.Empty(t, r.activeRuns)
				require.Equal(t, []*workflowRun{
					{
						repo:         "repoC",
						owner:        "ownerC",
						workflowName: "workflowC",
						completed:    true,
						errMsg:       "skipped because of cancel",
					},
					{
						repo:         "repoA",
						owner:        "ownerA",
						workflowName: "workflowA",
						completed:    true,
						runID:        1,
						errMsg:       "workflow was cancelled, conclusion: cancelled",
					},
					{
						repo:         "repoB",
						owner:        "ownerB",
						workflowName: "workflowB",
						completed:    true,
						runID:        2,
						errMsg:       "workflow was cancelled, conclusion: cancelled",
					},
				}, r.completedRuns)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			receiver.cancelActiveRuns(context.Background())

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

		})
	}
}

func TestWorkflowManager_handleResults(t *testing.T) {
	tests := []struct {
		name       string
		init       func(t *testing.T) *WorkflowManager
		wantErr    bool
		inspectErr func(err error, t *testing.T)
	}{
		{
			name: "no failed runs",
			init: func(t *testing.T) *WorkflowManager {
				return &WorkflowManager{
					completedRuns: []*workflowRun{
						{
							repo:         "repoA",
							owner:        "ownerA",
							workflowName: "workflowA",
							completed:    true,
						},
						{
							repo:         "repoB",
							owner:        "ownerB",
							workflowName: "workflowB",
							completed:    true,
						},
					},
				}
			},
		},
		{
			name: "failed runs",
			init: func(t *testing.T) *WorkflowManager {
				return &WorkflowManager{
					completedRuns: []*workflowRun{
						{
							repo:         "repoA",
							owner:        "ownerA",
							workflowName: "workflowA",
							completed:    true,
						},
						{
							repo:         "repoB",
							owner:        "ownerB",
							workflowName: "workflowB",
							completed:    true,
							errMsg:       "error",
						},
					},
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				require.ErrorContains(t, err, "some workflows failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			err := receiver.handleResults()

			if (err != nil) != tt.wantErr {
				t.Fatalf("WorkflowManager.handleResults error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

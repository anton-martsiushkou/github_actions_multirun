package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	gh "github.com/google/go-github/v46/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/anton-martsiushkou/github_actions_multirun/logger"
)

const (
	StatusCompleted   = "completed"
	ConclusionSuccess = "success"

	githubTokenEnvVar = "GITHUB_TOKEN"
)

type Github struct {
	client *gh.Client
}

func New(c *gh.Client) *Github {
	return &Github{client: c}
}

func InitGitHub(ctx context.Context) (*Github, error) {
	ghToken, ok := os.LookupEnv(githubTokenEnvVar)
	if !ok {
		return nil, errors.Errorf("github token should be set via %s env variable", githubTokenEnvVar)
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})

	return New(gh.NewClient(oauth2.NewClient(ctx, ts))), nil
}

func (g *Github) TriggerWorkflow(ctx context.Context, owner, repo, workflow, eventType string, params map[string]interface{}) error {
	jsonParams, err := json.Marshal(params)
	rm := json.RawMessage(jsonParams)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal params")
	}
	payload := gh.DispatchRequestOptions{
		EventType:     eventType,
		ClientPayload: &rm,
	}
	_, resp, err := g.client.Repositories.Dispatch(ctx, owner, repo, payload)
	if err != nil {
		return errors.WithMessagef(err, "failed to send request to trigger workflow %s", workflow)
	}

	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("failed to trigger workflow %s of repo %s for params %v, status code is %d", workflow, repo, params, resp.StatusCode)
	}

	return nil
}

func (g *Github) GetLastWorkflowRunNumber(ctx context.Context, owner, repo, workflowName string) (int, error) {
	wrs, resp, err := g.client.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, workflowName, &gh.ListWorkflowRunsOptions{
		ListOptions: gh.ListOptions{Page: 1, PerPage: 1},
	})
	if err != nil {
		return 0, errors.WithMessagef(err, "failed to list workflow for repo %s and workflow %s", repo, workflowName)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("failed to get last workflow for repo %s and workflow %s, status code is %d", repo, workflowName, resp.StatusCode)
	}

	if wrs.TotalCount == nil || *wrs.TotalCount == 0 {
		return 0, nil
	}

	return *wrs.WorkflowRuns[0].RunNumber, nil
}

func (g *Github) GetWorkflowRunID(ctx context.Context, owner, repo, workflowName string, minRunVersion int) (int64, error) {
	wrs, resp, err := g.client.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, workflowName, &gh.ListWorkflowRunsOptions{})
	if err != nil {
		return 0, errors.WithMessagef(err, "failed to list workflow runs for repo %s, workflow %s", repo, workflowName)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("failed to get last workflow run, status code is %d, repo %s, workflow %s", resp.StatusCode, repo, workflowName)
	}

	var runID int64
	for _, wr := range wrs.WorkflowRuns {
		if *wr.Status != StatusCompleted && *wr.RunNumber > minRunVersion {
			runID = *wr.ID
			break
		}
	}

	if runID == 0 {
		return 0, errors.New(fmt.Sprintf("workflow for repo %s was not started, workflow name %s", repo, workflowName))
	}

	return runID, nil
}

func (g *Github) GetWorkflowRunStatus(ctx context.Context, owner, repo string, runID int64) (string, string, time.Duration, error) {
	var (
		conclusion  string
		runDuration time.Duration
	)

	w, resp, err := g.client.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return "", "", runDuration, errors.WithMessagef(err, "failed to get workflow run status for run id %d", runID)
	}
	logger.FromContext(ctx).Debugf("remaining number of calls to github api is %d", resp.Rate.Remaining)

	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", "", runDuration, errors.Errorf("failed to get workflow status for run %d, status code is %d", runID, resp.StatusCode)
	}

	if w.Status == nil {
		return "", "", runDuration, errors.New(fmt.Sprintf("something went wrong and status for run %d is nil", runID))
	}

	if w.Conclusion != nil {
		conclusion = *w.Conclusion
	}

	if !w.UpdatedAt.IsZero() && !w.RunStartedAt.IsZero() {
		runDuration = w.UpdatedAt.Sub(w.RunStartedAt.Time)
	}

	return *w.Status, conclusion, runDuration, nil
}

func (g *Github) CancelWorkflowRun(ctx context.Context, owner, repo string, runID int64) error {
	resp, err := g.client.Actions.CancelWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return errors.WithMessagef(err, "failed to cancel workflow run %d", runID)
	}

	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("failed to cancel workflow run %d, status code is %d", runID, resp.StatusCode)
	}

	return nil
}

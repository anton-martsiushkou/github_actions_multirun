package multirun

// DO NOT EDIT!
// This code is generated with http://github.com/hexdigest/gowrap tool
// using ../templates/log.txt template

//go:generate gowrap gen -p multirun -i Github -t ../templates/log.txt -o github_with_log.go

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// GithubWithLog implements Github that is instrumented with logging
type GithubWithLog struct {
	_base Github
	_l    *logrus.Logger
}

// NewGithubWithLog instruments an implementation of the Github with simple logging
func NewGithubWithLog(base Github, l *logrus.Logger) GithubWithLog {
	return GithubWithLog{
		_base: base,
		_l:    l,
	}
}

// CancelWorkflowRun implements Github
func (_d GithubWithLog) CancelWorkflowRun(ctx context.Context, owner string, repo string, runID int64) (err error) {
	_d._l.
		WithField("ctx", ctx).
		WithField("owner", owner).
		WithField("repo", repo).
		WithField("runID", runID).
		Debug("GithubWithLog: CancelWorkflowRun called")
	defer func() {
		if err != nil {
			_d._l.
				WithField("err", err).
				Error("GithubWithLog: CancelWorkflowRun returned results")
		} else {
			_d._l.
				WithField("err", err).
				Debug("GithubWithLog: CancelWorkflowRun returned results")
		}
	}()
	return _d._base.CancelWorkflowRun(ctx, owner, repo, runID)
}

// GetLastWorkflowRunNumber implements Github
func (_d GithubWithLog) GetLastWorkflowRunNumber(ctx context.Context, owner string, repo string, workflowName string) (i1 int, err error) {
	_d._l.
		WithField("ctx", ctx).
		WithField("owner", owner).
		WithField("repo", repo).
		WithField("workflowName", workflowName).
		Debug("GithubWithLog: GetLastWorkflowRunNumber called")
	defer func() {
		if err != nil {
			_d._l.
				WithField("i1", i1).
				WithField("err", err).
				Error("GithubWithLog: GetLastWorkflowRunNumber returned results")
		} else {
			_d._l.
				WithField("i1", i1).
				WithField("err", err).
				Debug("GithubWithLog: GetLastWorkflowRunNumber returned results")
		}
	}()
	return _d._base.GetLastWorkflowRunNumber(ctx, owner, repo, workflowName)
}

// GetWorkflowRunID implements Github
func (_d GithubWithLog) GetWorkflowRunID(ctx context.Context, owner string, repo string, workflowName string, minRunVersion int) (i1 int64, err error) {
	_d._l.
		WithField("ctx", ctx).
		WithField("owner", owner).
		WithField("repo", repo).
		WithField("workflowName", workflowName).
		WithField("minRunVersion", minRunVersion).
		Debug("GithubWithLog: GetWorkflowRunID called")
	defer func() {
		if err != nil {
			_d._l.
				WithField("i1", i1).
				WithField("err", err).
				Error("GithubWithLog: GetWorkflowRunID returned results")
		} else {
			_d._l.
				WithField("i1", i1).
				WithField("err", err).
				Debug("GithubWithLog: GetWorkflowRunID returned results")
		}
	}()
	return _d._base.GetWorkflowRunID(ctx, owner, repo, workflowName, minRunVersion)
}

// GetWorkflowRunStatus implements Github
func (_d GithubWithLog) GetWorkflowRunStatus(ctx context.Context, owner string, repo string, runID int64) (s1 string, s2 string, d1 time.Duration, err error) {
	_d._l.
		WithField("ctx", ctx).
		WithField("owner", owner).
		WithField("repo", repo).
		WithField("runID", runID).
		Debug("GithubWithLog: GetWorkflowRunStatus called")
	defer func() {
		if err != nil {
			_d._l.
				WithField("s1", s1).
				WithField("s2", s2).
				WithField("d1", d1).
				WithField("err", err).
				Error("GithubWithLog: GetWorkflowRunStatus returned results")
		} else {
			_d._l.
				WithField("s1", s1).
				WithField("s2", s2).
				WithField("d1", d1).
				WithField("err", err).
				Debug("GithubWithLog: GetWorkflowRunStatus returned results")
		}
	}()
	return _d._base.GetWorkflowRunStatus(ctx, owner, repo, runID)
}

// TriggerWorkflow implements Github
func (_d GithubWithLog) TriggerWorkflow(ctx context.Context, owner string, repo string, workflow string, eventType string, params map[string]interface{}) (err error) {
	_d._l.
		WithField("ctx", ctx).
		WithField("owner", owner).
		WithField("repo", repo).
		WithField("workflow", workflow).
		WithField("eventType", eventType).
		WithField("params", params).
		Debug("GithubWithLog: TriggerWorkflow called")
	defer func() {
		if err != nil {
			_d._l.
				WithField("err", err).
				Error("GithubWithLog: TriggerWorkflow returned results")
		} else {
			_d._l.
				WithField("err", err).
				Debug("GithubWithLog: TriggerWorkflow returned results")
		}
	}()
	return _d._base.TriggerWorkflow(ctx, owner, repo, workflow, eventType, params)
}

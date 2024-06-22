package multirun

import (
	"fmt"
	"time"
)

type workflowRun struct {
	repo                                   string
	owner                                  string
	workflowName                           string
	errMsg                                 string
	runID                                  int64
	link                                   string
	failedCheckAttempts                    int
	currentCheckAttemptNumber              int
	completed                              bool
	initialCheckInterval                   time.Duration
	finalCheckInterval                     time.Duration
	attemptsThresholdForShortCheckInterval int
	alias                                  string
	params                                 map[string]interface{}
	duration                               time.Duration
}

func (wr *workflowRun) setLink() {
	wr.link = fmt.Sprintf("https://github.com/%s/%s/actions/runs/%d", wr.owner, wr.repo, wr.runID)
}

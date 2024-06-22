package multirun

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	log "github.com/sirupsen/logrus"

	"github.com/anton-martsiushkou/github_actions_multirun/github"
	"github.com/anton-martsiushkou/github_actions_multirun/logger"
)

const (
	StatePass       = "PASS"
	StateFail       = "FAIL"
	StateInProgress = "IN PROGRESS"
	StateQueued     = "QUEUED"

	resultsFile = "results.txt"
)

type Github interface {
	TriggerWorkflow(ctx context.Context, owner, repo, workflow, eventType string, params map[string]interface{}) error
	GetLastWorkflowRunNumber(ctx context.Context, owner, repo, workflowName string) (int, error)
	GetWorkflowRunID(ctx context.Context, owner, repo, workflowName string, minRunVersion int) (int64, error)
	GetWorkflowRunStatus(ctx context.Context, owner, repo string, runID int64) (string, string, time.Duration, error)
	CancelWorkflowRun(ctx context.Context, owner, repo string, runID int64) error
}

type WorkflowManager struct {
	gh       Github
	wmConfig *WorkflowManagerConfig

	completedRuns []*workflowRun
	activeRuns    []*workflowRun
	allWorkflows  []*workflowRun
	queue         chan *workflowRun

	wg              sync.WaitGroup
	resultsMutex    sync.Mutex
	activeRunsMutex sync.Mutex
	done            chan struct{}
}

func NewWorkflowManager(commonParams map[string]interface{}, wmc *WorkflowManagerConfig, gh Github) *WorkflowManager {
	wm := &WorkflowManager{
		wmConfig:   wmc,
		gh:         gh,
		activeRuns: make([]*workflowRun, 0),
		done:       make(chan struct{}),
	}
	wm.initQueue(commonParams)

	return wm
}

func (wm *WorkflowManager) initQueue(commonParams map[string]interface{}) {
	wm.queue = make(chan *workflowRun, len(wm.wmConfig.Workflows))
	wm.allWorkflows = make([]*workflowRun, len(wm.wmConfig.Workflows))
	for i, w := range wm.wmConfig.Workflows {
		params := make(map[string]interface{})
		if w.Params != nil {
			params = w.Params
		}
		for k, v := range commonParams {
			params[k] = v
		}
		wr := &workflowRun{
			repo:                                   w.RepoName,
			owner:                                  w.OwnerName,
			workflowName:                           w.WorkflowName,
			initialCheckInterval:                   w.InitialCheckInterval,
			finalCheckInterval:                     w.FinalCheckInterval,
			attemptsThresholdForShortCheckInterval: w.AttemptsThresholdForShortCheckInterval,
			alias:                                  w.Alias,
			params:                                 params,
		}
		wm.queue <- wr
		wm.allWorkflows[i] = wr
	}
	close(wm.queue)
}

func (wm *WorkflowManager) Run(ctx context.Context) error {
	l := logger.FromContext(ctx)

	// handle cancel signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// set timeout for the whole process
	timeout := time.After(wm.wmConfig.MultirunTimeout)

	// periodical printing of results
	ticker := time.NewTicker(wm.wmConfig.PrintResultsInterval)
	defer ticker.Stop()
	go wm.periodicalResultsPrinting(ticker)

	// start workers
	for i := 1; i <= wm.wmConfig.WorkersCount; i++ {
		wm.wg.Add(1)
		go wm.worker(ctx, i)
	}

	// print final results and write them to file
	defer func() {
		wm.printResults()
		wm.resultsToFile(ctx)
	}()

	// wait for all workers to finish
	go func() {
		wm.wg.Wait()
		wm.done <- struct{}{}
	}()

	select {
	case <-wm.done:
		l.Info("all workers are done")
	case <-timeout:
		ticker.Stop()
		wm.printResults()
		l.Error("timeout reached. Exiting...")
		wm.gracefulShutdown(ctx)
	case <-shutdown:
		ticker.Stop()
		l.Info("received signal for cancelling of job")
		wm.gracefulShutdown(ctx)
	}

	return wm.handleResults()
}

func (wm *WorkflowManager) worker(ctx context.Context, workerNumber int) {
	defer wm.wg.Done()
	l := logger.FromContext(ctx)
	for wr := range wm.queue {
		l.Infof("worker #%d started new job for repo %s, workflow %s", workerNumber, wr.repo, wr.workflowName)
		wm.addActiveRun(ctx, wr)
		wm.launchWorkflow(ctx, wr)
		wm.waitForWorkflowCompletion(ctx, wr)
		l.Infof("worker #%d completed job for repo %s, workflow %s", workerNumber, wr.repo, wr.workflowName)
	}
}

func (wm *WorkflowManager) launchWorkflow(ctx context.Context, wr *workflowRun) {
	var (
		l         = logger.FromContext(ctx)
		lastRunID int
		runID     int64
		err       error
	)

	lastRunID, err = wm.gh.GetLastWorkflowRunNumber(ctx, wr.owner, wr.repo, wr.workflowName)
	if err != nil {
		l.WithError(err).WithFields(
			map[string]interface{}{
				"repo":     wr.repo,
				"workflow": wr.workflowName,
			}).Error("failed to get last workflow run number")
		wr.errMsg = fmt.Sprintf("failed to get last workflow run number: %v", err)
		wm.addCompletedRun(ctx, wr)
		return
	}

	if err = wm.gh.TriggerWorkflow(ctx, wr.owner, wr.repo, wr.workflowName, wm.wmConfig.DispatchEventType, wr.params); err != nil {
		l.WithError(err).WithFields(
			map[string]interface{}{
				"repo":     wr.repo,
				"workflow": wr.workflowName,
			}).Error("failed to trigger workflow")
		wr.errMsg = fmt.Sprintf("failed to trigger workflow: %v", err)
		wm.addCompletedRun(ctx, wr)
		return
	}
	l.WithFields(
		map[string]interface{}{
			"repo":     wr.repo,
			"workflow": wr.workflowName,
		}).Info("workflow was successfully triggered, looking for run id")

	time.Sleep(wm.wmConfig.TimeToWaitRunIDs)

	runID, err = wm.gh.GetWorkflowRunID(ctx, wr.owner, wr.repo, wr.workflowName, lastRunID)
	if err != nil {
		l.WithError(err).WithFields(
			map[string]interface{}{
				"repo":     wr.repo,
				"workflow": wr.workflowName,
			}).Error("failed to get workflow run id")
		wr.errMsg = fmt.Sprintf("failed to get workflow run ID: %v", err)
		wm.addCompletedRun(ctx, wr)
		return
	}

	wr.runID = runID
	wr.setLink()

	l.WithFields(
		map[string]interface{}{
			"repo":     wr.repo,
			"workflow": wr.workflowName,
			"link":     wr.link,
		}).Info("workflow was successfully launched")
}

func (wm *WorkflowManager) waitForWorkflowCompletion(ctx context.Context, wr *workflowRun) {
	var (
		l            = logger.FromContext(ctx)
		checkCounter int
	)

	if wr.runID == 0 {
		l.WithFields(
			map[string]interface{}{
				"repo":     wr.repo,
				"workflow": wr.workflowName,
				"errMsg":   wr.errMsg,
			}).Info("run id is empty for this workflow run, skipping waiting for completion")
		return
	}

	wLog := l.WithFields(
		map[string]interface{}{
			"repo":     wr.repo,
			"workflow": wr.workflowName,
			"link":     wr.link,
		})

	wLog.Info("start waiting for workflow completion")

	for {
		checkCounter += 1
		wLog.Debugf("check #%d", checkCounter)

		status, conclusion, duration, err := wm.gh.GetWorkflowRunStatus(ctx, wr.owner, wr.repo, wr.runID)
		if err != nil {
			wLog.WithError(err).Debug("failed to get workflow run status")
			wm.handleError(ctx, wLog, wr, err)
			if wr.completed {
				break
			}
			continue
		}
		wr.duration = duration

		if status != github.StatusCompleted {
			wLog.Debug("workflow is not completed, continue checking")
			wm.waitForNextCheck(wr)
			continue
		}

		switch conclusion {
		case "":
			wLog.Debug("conclusion for completed workflow run is empty")
			wm.handleError(ctx, wLog, wr, err)
			if wr.completed {
				return
			}
		case github.ConclusionSuccess:
			wLog.Info("successful completion of workflow")
			wm.addCompletedRun(ctx, wr)
			return
		default:
			wLog.Info("workflow failed")
			wr.errMsg = fmt.Sprintf("workflow failed with conclusion: %s", conclusion)
			wm.addCompletedRun(ctx, wr)
			return
		}
	}
}

func (wm *WorkflowManager) addCompletedRun(ctx context.Context, wr *workflowRun) {
	l := logger.FromContext(ctx)

	wm.resultsMutex.Lock()
	defer wm.resultsMutex.Unlock()
	wr.completed = true
	wm.completedRuns = append(wm.completedRuns, wr)
	l.WithFields(
		map[string]interface{}{
			"repo":     wr.repo,
			"workflow": wr.workflowName,
			"link":     wr.link,
			"errMsg":   wr.errMsg,
		}).Info("workflow run was added to the list of completed runs")

	wm.activeRunsMutex.Lock()
	defer wm.activeRunsMutex.Unlock()
	for i, r := range wm.activeRuns {
		if r.repo == wr.repo && r.workflowName == wr.workflowName && r.owner == wr.owner {
			wm.activeRuns = append(wm.activeRuns[:i], wm.activeRuns[i+1:]...)
			l.WithFields(
				map[string]interface{}{
					"repo":     wr.repo,
					"workflow": wr.workflowName,
					"link":     wr.link,
				}).Info("run removed from list of active runs")
			break
		}
	}
}

func (wm *WorkflowManager) handleError(ctx context.Context, wLog *log.Entry, wr *workflowRun, err error) {
	if wr.failedCheckAttempts >= wm.wmConfig.MaxFailedCheckAttempts {
		wLog.WithError(err).Debug("failed check attempts limit exceeded")
		wr.errMsg = fmt.Sprintf("failed check attempts limit exceeded, last error: %v", err)
		wm.addCompletedRun(ctx, wr)
		return
	}
	wLog.WithError(err).Debug("verification of workflow status will be rerun after error")
	wr.failedCheckAttempts += 1
}

func (wm *WorkflowManager) waitForNextCheck(wr *workflowRun) {
	defer func() {
		wr.currentCheckAttemptNumber += 1
	}()
	if wr.currentCheckAttemptNumber < wr.attemptsThresholdForShortCheckInterval {
		time.Sleep(wr.initialCheckInterval)
		return
	}
	time.Sleep(wr.finalCheckInterval)
}

func (wm *WorkflowManager) periodicalResultsPrinting(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			wm.printResults()
		default:
			continue
		}
	}
}

func (wm *WorkflowManager) buildResultTable() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetTitle("Results")
	t.Style().Title.Align = text.AlignCenter
	t.Style().Title.Format = text.FormatUpper
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateRows = true
	t.AppendHeader(table.Row{"#", "Repo", "Workflow", "Status", "Link", "Error", "Duration"})
	t.SetRowPainter(func(row table.Row) text.Colors {
		switch row[3] {
		case StateQueued:
			return text.Colors{text.FgYellow}
		case StateFail:
			return text.Colors{text.FgRed}
		case StatePass:
			return text.Colors{text.FgGreen}
		case StateInProgress:
			return text.Colors{text.FgBlue}
		}
		return text.Colors{text.BgHiWhite}
	})

	for i, r := range wm.allWorkflows {
		status := fmt.Sprintf("⏳ %s", StateQueued)
		index := slices.IndexFunc(wm.activeRuns, func(wr *workflowRun) bool {
			return wr == r
		})
		if index != -1 {
			status = fmt.Sprintf("🔄 %s", StateInProgress)
		}

		index = slices.IndexFunc(wm.completedRuns, func(wr *workflowRun) bool {
			return wr == r
		})
		if index != -1 {
			if wm.completedRuns[index].errMsg == "" {
				status = fmt.Sprintf("✅ %s", StatePass)
			} else {
				status = fmt.Sprintf("❌ %s", StateFail)
			}
		}

		name := r.repo
		if r.alias != "" {
			name = r.alias
		}
		t.AppendRow(table.Row{i + 1, name, r.workflowName, status, r.link, r.errMsg, r.duration.String()})
	}

	return t
}

func (wm *WorkflowManager) cancelActiveRuns(ctx context.Context) {
	l := logger.FromContext(ctx)
	l.Info("all active runs will be cancelled")
	for wr := range wm.queue {
		wr.errMsg = "skipped because of cancel"
		wm.addCompletedRun(ctx, wr)
	}

	for _, r := range wm.activeRuns {
		if r.runID == 0 {
			continue
		}
		if err := wm.gh.CancelWorkflowRun(context.Background(), r.owner, r.repo, r.runID); err != nil {
			l.WithError(err).WithFields(
				map[string]interface{}{
					"repo":     r.repo,
					"link":     r.link,
					"workflow": r.workflowName,
				}).Error("failed to cancel workflow run")
		}
	}

	for {
		if len(wm.activeRuns) == 0 {
			break
		}
		time.Sleep(wm.wmConfig.CancelInterval)
		for _, r := range wm.activeRuns {
			status, conclusion, duration, err := wm.gh.GetWorkflowRunStatus(context.Background(), r.owner, r.repo, r.runID)
			if err != nil {
				l.WithError(err).WithFields(
					map[string]interface{}{
						"repo":     r.repo,
						"link":     r.link,
						"workflow": r.workflowName,
					}).Error("failed to get workflow run status")
				r.errMsg = fmt.Sprintf("workflow was cancelled, unable to get it's status: %s", err.Error())
				wm.addCompletedRun(ctx, r)
				continue
			}
			if status == github.StatusCompleted {
				r.errMsg = fmt.Sprintf("workflow was cancelled, conclusion: %s", conclusion)
				r.duration = duration
				wm.addCompletedRun(ctx, r)
				continue
			}
		}
	}
}

func (wm *WorkflowManager) addActiveRun(ctx context.Context, wr *workflowRun) {
	wm.activeRunsMutex.Lock()
	defer wm.activeRunsMutex.Unlock()
	wm.activeRuns = append(wm.activeRuns, wr)
	logger.FromContext(ctx).WithFields(
		map[string]interface{}{
			"repo":     wr.repo,
			"workflow": wr.workflowName,
		}).Info("run added to list of active runs")
}

func (wm *WorkflowManager) printResults() {
	t := wm.buildResultTable()
	fmt.Println(t.Render())
}

func (wm *WorkflowManager) resultsToFile(ctx context.Context) {
	l := logger.FromContext(ctx)
	f, err := os.OpenFile(resultsFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		l.WithError(err).Error("failed to open results file")
	}

	defer func() {
		_ = f.Close()
	}()

	t := wm.buildResultTable()
	if _, err = f.WriteString(t.RenderMarkdown()); err != nil {
		l.WithError(err).Error("failed to write to results file")
	}
}

func (wm *WorkflowManager) handleResults() error {
	failedWorkflows := make([]string, 0)
	for _, r := range wm.completedRuns {
		if r.errMsg != "" {
			failedWorkflows = append(failedWorkflows, fmt.Sprintf("%s - /%s", r.repo, r.workflowName))
		}
	}
	if len(failedWorkflows) > 0 {
		return fmt.Errorf("some workflows failed: %v", failedWorkflows)
	}

	return nil
}

func (wm *WorkflowManager) gracefulShutdown(ctx context.Context) {
	var (
		l      = logger.FromContext(ctx)
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(context.Background(), wm.wmConfig.GracefulShutdownTimeout)
	defer cancel()

	go func() {
		wm.cancelActiveRuns(ctx)
		wm.done <- struct{}{}
	}()

	select {
	case <-wm.done:
		l.Info("all active runs were cancelled successfully")
	case <-ctx.Done():
		wm.printResults()
		l.Error("graceful shutdown timeout reached. Exiting...")
		os.Exit(1)
	}
}

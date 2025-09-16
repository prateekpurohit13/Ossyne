package tui

type viewState int

const (
	viewTasks viewState = iota
	viewSubmit
	viewProjects
	viewCreateTask
	viewProjectTasks
	viewFundBountyForm
)

type taskClaimedMsg struct{ taskID uint }
type contributionSubmittedMsg struct{ taskID uint }
type taskCreatedMsg struct{ taskID uint }
type bountyFundedMsg struct {
	taskID uint
	amount float64
}
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

const (
	textInputHeight = 1
)
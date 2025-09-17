package tui

import "ossyne/internal/models"

type viewState int

const (
	viewLanding viewState = iota
	viewAuth
	viewTasks
	viewSubmit
	viewProjects
	viewCreateProject
	viewManageProjects
	viewCreateProjectForm
	viewCreateTask
	viewProjectTasks
	viewFundBountyForm
	viewMyContributions
	viewReviewContributions
	viewMyWallet
)

type taskClaimedMsg struct{ taskID uint }
type contributionSubmittedMsg struct{ taskID uint }
type taskCreatedMsg struct{ taskID uint }
type projectCreatedMsg struct{ projectID uint }
type projectCreateSubmitMsg struct {
	title         string
	description   string
	repositoryURL string
	isPublic      bool
	tags          []string
}
type bountyFundedMsg struct {
	taskID uint
	amount float64
}
type userFetchedMsg struct{ user *models.User }
type notLoggedInMsg struct{}
type startLoginFlowMsg struct{}
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

const (
	textInputHeight = 1
)
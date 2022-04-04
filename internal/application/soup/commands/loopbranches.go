package commands

import (
	"fmt"
	"time"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/domain/soup"
)

type LoopBranchesRequest struct {
	URL    string
	Period int
}

type LoopBranchesRequestHandler interface {
	Handle(command LoopBranchesRequest) error
}

type loopBranchesRequestHandler struct {
	logger logger.Logger
	repo   soup.Git
}

//NewUpdateCragRequestHandler Constructor
func NewLoopBranchesRequestHandler(git soup.Git, logger logger.Logger) LoopBranchesRequestHandler {

	return loopBranchesRequestHandler{
		repo:   git,
		logger: logger,
	}
}

//Handle Handles the update request
func (h loopBranchesRequestHandler) Handle(command LoopBranchesRequest) error {
	var cloneLocation string
	var branchNames []string
	var err error

	// Clone repo
	cloneLocation = fmt.Sprintf("%s%d", "/tmp/soup/", time.Now().Unix())
	if err = h.repo.PlainClone(cloneLocation, command.URL); err != nil {
		return err
	}
	// Get branch names
	if branchNames, err = h.repo.GetBranchNames(); err != nil {
		return err
	}
	h.logger.WithFields(logger.Fields{
		"branches": branchNames,
	}).Info("Branches parsed")
	return nil
}

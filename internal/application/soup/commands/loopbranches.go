package commands

import (
	"fmt"
	"time"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/domain/soup"
)

type LoopBranchesRequest struct {
	URL      string
	Period   int
	Token    string
	Username string
}

type LoopBranchesRequestHandler interface {
	Handle(command LoopBranchesRequest) error
}

type loopBranchesRequestHandler struct {
	logger logger.Logger
	svc    soup.Git
	config soup.Config
}

//NewUpdateCragRequestHandler Constructor
func NewLoopBranchesRequestHandler(git soup.Git, config soup.Config, logger logger.Logger) LoopBranchesRequestHandler {

	return loopBranchesRequestHandler{
		config: config,
		svc:    git,
		logger: logger,
	}
}

//Handle Handles the update request
func (h loopBranchesRequestHandler) Handle(command LoopBranchesRequest) error {
	var cloneLocation string
	var branchNames []string
	var err error
	var info soup.SoupInfo

	// Clone repo
	cloneLocation = fmt.Sprintf("%s%d", "/tmp/soup/", time.Now().Unix())
	if err = h.svc.PlainClone(cloneLocation, command.URL, command.Username, command.Token); err != nil {
		return err
	}
	// Get branch names
	if branchNames, err = h.svc.GetBranchNames(command.Username, command.Token); err != nil {
		return err
	}
	h.logger.WithFields(logger.Fields{
		"branches": branchNames,
	}).Info("Branches parsed")
	// Fetch branches
	if err = h.svc.Fetch(command.Username, command.Token); err != nil {
		return err
	}

	// Checkout to the branches and do GitOps stuff
	for _, branchName := range branchNames {
		// Checkout
		if err = h.svc.Checkout(branchName); err != nil {
			return err
		}
		// Process branch
		info = h.config.GetSoupInfo(cloneLocation)
		h.logger.WithFields(logger.Fields{
			"info": info,
		}).Info("soup.yaml parsed")
		// err = processBranch(branchName)
		// if err != nil {
		// 	fmt.Println("Error processing branch")
		// 	panic(err)
		// }
	}
	// os.RemoveAll(cloneLocation)

	return nil
}

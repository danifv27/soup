package commands

import (
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/domain/soup"
)

type LoopBranchesRequest struct{}

type LoopBranchesRequestHandler interface {
	Handle(command LoopBranchesRequest) error
}

type loopBranchesRequestHandler struct {
	logger logger.Logger
	svc    soup.Git
	config soup.Config
	deploy soup.Deploy
}

//NewLoopBranchesRequestHandler Constructor
func NewLoopBranchesRequestHandler(git soup.Git, deploy soup.Deploy, config soup.Config, logger logger.Logger) LoopBranchesRequestHandler {

	return loopBranchesRequestHandler{
		config: config,
		svc:    git,
		logger: logger,
		deploy: deploy,
	}
}

//Handle Handles the update request
func (h loopBranchesRequestHandler) Handle(command LoopBranchesRequest) error {
	var cloneLocation string
	var branchNames []string
	var err error

	// Clone repo
	cloneLocation = fmt.Sprintf("%s%d", "/tmp/soup/", time.Now().Unix())
	defer func(path string) {
		os.RemoveAll(path)
	}(cloneLocation)
	if err = h.svc.PlainClone(cloneLocation); err != nil {
		return fmt.Errorf("Handle: %w", err)
	}
	// Get branch names
	if branchNames, err = h.svc.GetBranchNames(); err != nil {
		return fmt.Errorf("Handle: %w", err)
	}
	// Fetch branches
	if err = h.svc.Fetch(); err != nil {
		return fmt.Errorf("Handle: %w", err)
	}

	// Checkout to the branches and do GitOps stuff
	for _, branchName := range branchNames {
		if err = h.checkoutAndProcess(branchName, cloneLocation); err != nil {
			return fmt.Errorf("Handle: %w", err)
		}
	}

	return nil
}

func (h loopBranchesRequestHandler) checkoutAndProcess(branchName string, cloneLocation string) error {
	var err error
	var info soup.SoupInfo
	var yml []byte

	// Checkout
	if err = h.svc.Checkout(branchName); err != nil {
		return err
	}
	// Process branch
	info = h.config.GetSoupInfo(cloneLocation) //If there is no .soup.yaml file info.kustomizations will be empty
	fSys := filesys.MakeFsOnDisk()
	kst := krusty.MakeKustomizer(
		HonorKustomizeFlags(krusty.MakeDefaultOptions()),
	)
	for _, k := range info.Kustomizations {
		if k.Branch != branchName {
			continue
		}
		m, err := kst.Run(fSys, fmt.Sprintf("%s/%s", info.Root, k.Overlay))
		if err != nil {
			return err
		}
		for _, r := range m.Resources() {
			yml, err = r.AsYAML()
			if err != nil {
				return err
			}
			err = h.deploy.Apply(k.Namespace, yml)
			if err != nil {
				return err
			}
			h.logger.WithFields(logger.Fields{
				"branch":    branchName,
				"namespace": k.Namespace,
				"gvk":       r.GetGvk(),
				"name":      r.GetName(),
			}).Info("resource deployed")
		}
	}

	return nil
}

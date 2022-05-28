package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/domain/soup"
)

type ProcessBranchRequest struct {
	// Branch reference, usually obtained from the webhook data.
	// For bitbucket repos is in the form ref/head/<branch name>
	Branch string
}

type ProcessBranchRequestHandler interface {
	Handle(command ProcessBranchRequest) error
}

type processBranchRequestHandler struct {
	logger  logger.Logger
	svc     soup.Git
	config  soup.Config
	deploy  soup.Deploy
	auditer audit.Auditer
}

//NewProcessBranchRequestHandler Constructor
func NewProcessBranchRequestHandler(git soup.Git, deploy soup.Deploy, config soup.Config, logger logger.Logger, audit audit.Auditer) ProcessBranchRequestHandler {

	return processBranchRequestHandler{
		config:  config,
		svc:     git,
		logger:  logger,
		deploy:  deploy,
		auditer: audit,
	}
}

//Handle Handles the update request
func (h processBranchRequestHandler) Handle(command ProcessBranchRequest) error {
	var cloneLocation string
	var err error

	// Clone repo
	cloneLocation = fmt.Sprintf("%s%d", "/tmp/soup/", time.Now().Unix())
	defer func(path string) {
		os.RemoveAll(path)
	}(cloneLocation)
	if err = h.svc.PlainClone(cloneLocation); err != nil {
		return err
	}
	// Fetch branches
	if err = h.svc.Fetch(); err != nil {
		return err
	}

	// Checkout the branch and do GitOps stuff
	if err = h.checkoutAndProcess(command.Branch, cloneLocation); err != nil {
		return err
	}

	return nil
}

func (h processBranchRequestHandler) checkoutAndProcess(branchName string, cloneLocation string) error {
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
		if !strings.Contains(branchName, k.Branch) {
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

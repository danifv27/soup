package commands

import (
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/deployment"
	"github.com/danifv27/soup/internal/domain/soup"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

type LoopBranchesRequest struct {
	URL      string
	Token    string
	Username string
	Path     string
}

type LoopBranchesRequestHandler interface {
	Handle(command LoopBranchesRequest) error
}

type loopBranchesRequestHandler struct {
	logger logger.Logger
	svc    soup.Git
	config soup.Config
}

//NewLoopBranchesRequestHandler Constructor
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
		// h.logger.WithFields(logger.Fields{
		// 	"info": info,
		// }).Info("soup.yaml parsed")
		fSys := filesys.MakeFsOnDisk()
		kst := krusty.MakeKustomizer(
			HonorKustomizeFlags(krusty.MakeDefaultOptions()),
		)
		for _, k := range info.Kustomizations {
			m, err := kst.Run(fSys, fmt.Sprintf("%s/%s", info.Root, k.Overlay))
			if err != nil {
				return err
			}
			for _, r := range m.Resources() {
				yml, err := r.AsYAML()
				if err != nil {
					return err
				}
				// os.Stdout.Write(yml)
				err = deployment.Deploy(command.Path, k.Namespace, yml)
				if err != nil {
					return err
				}
			}
		}

	}
	os.RemoveAll(cloneLocation)

	return nil
}

// HonorKustomizeFlags feeds command line data to the krusty options.
// Flags and such are held in private package variables.
func HonorKustomizeFlags(kOpts *krusty.Options) *krusty.Options {
	kOpts.DoLegacyResourceSort = true
	// Files referenced by a kustomization file must be in
	// or under the directory holding the kustomization
	// file itself.
	kOpts.LoadRestrictions = types.LoadRestrictionsRootOnly
	kOpts.PluginConfig.HelmConfig.Enabled = false
	kOpts.PluginConfig.HelmConfig.Command = ""
	// When true, a label
	//     app.kubernetes.io/managed-by: kustomize-<version>
	// is added to all the resources in the build out.
	kOpts.AddManagedbyLabel = false

	return kOpts
}

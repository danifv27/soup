package commands

import (
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

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

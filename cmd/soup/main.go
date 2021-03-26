package main

import (
	"flag"
	"fmt"
	git "github.com/go-git/go-git/v5"
	config "github.com/go-git/go-git/v5/config"
	plumbing "github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

// Global variables
var programConf ProgramConf

// Structs
type Namespace struct {
	Namespace string
	Branch    string
}

type BuildConf struct {
	Namespaces []Namespace
	Manifests  []string
}

type ProgramConf struct {
	Repo     string
	Interval int
}

// Auxiliary functions
func getBranchNames(r *git.Repository) []string {
	var branchNames []string
	remote, err := r.Remote("origin")
	if err != nil {
		panic(err)
	}
	refList, err := remote.List(&git.ListOptions{})
	if err != nil {
		panic(err)
	}
	refPrefix := "refs/heads/"
	for _, ref := range refList {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]
		branchNames = append(branchNames, branchName)
	}
	return branchNames
}

func getNamespace(branchName string, buildConf BuildConf) string {
	var namespace string = ""
	for _, a := range buildConf.Namespaces {
		matched, err := regexp.MatchString(a.Branch, branchName)
		if err != nil {
			panic(err)
		}
		if matched {
			if a.Namespace == "as-branch" {
				namespace = branchName
			} else {
				namespace = a.Namespace
			}
			return namespace
		}
	}
	return ""
}

func getBuildConf(cloneLocation string) BuildConf {
	var buildConf BuildConf
	yamlFile, err := ioutil.ReadFile(cloneLocation + "/.soup.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &buildConf)
	if err != nil {
		panic(err)
	}
	return buildConf
}

func deploy(namespace string, manifests []string) error {
	//fmt.Println(manifests)
	return nil
}

// Core functions
func init() {
	//var programConf ProgramConf
	flag.StringVar(&programConf.Repo, "repo", "", "url of the repository")
	flag.Parse()
	if programConf.Repo == "" {
		fmt.Println("Exiting, repo flag is not provided")
		os.Exit(1)
	}
	flag.IntVar(&programConf.Interval, "interval", 120, "execution interval")
}

func processBranch(branchName string, cloneLocation string) error {
	// Get configuration from file
	var buildConf BuildConf = getBuildConf(cloneLocation)
	// Process configuration
	var namespace string = getNamespace(branchName, buildConf)
	if namespace == "" {
		fmt.Println("Branch " + branchName + " does not match with any namespace to be deployed")
		return nil
	}
	fmt.Println("Deploying branch " + branchName + " to namespace " + namespace)
	// Deploy
	err := deploy(namespace, buildConf.Manifests)
	if err != nil {
		panic(err)
	}
	return nil
}

func run() error {
	// Clone repo
	cloneLocation := fmt.Sprintf("%s%d", "/tmp/soup/", time.Now().Unix())
	r, err := git.PlainClone(cloneLocation, false, &git.CloneOptions{
		URL: programConf.Repo,
	})
	if err != nil {
		panic(err)
	}
	// Get branch names
	branchNames := getBranchNames(r)
	// Fetch branches
	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	})
	if err != nil {
		panic(err)
	}
	// Checkout to the branches and do GitOps stuff
	w, _ := r.Worktree()
	for _, branchName := range branchNames {
		// Checkout
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + branchName),
			Force:  true,
		})
		if err != nil {
			panic(err)
		}
		// Process branch
		err = processBranch(branchName, cloneLocation)
		if err != nil {
			panic(err)
		}
	}
	os.RemoveAll(cloneLocation)
	fmt.Printf("%s%d%s", "Sleeping ", programConf.Interval, "s until next execution...\n\n")
	time.Sleep(time.Second * time.Duration(programConf.Interval))
	return nil
}

func main() {
	for {
		err := run()
		if err != nil {
			panic(err)
		}
	}
}

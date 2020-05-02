package main

import (
	"errors"
	"fmt"
	"github.com/markbates/pkger"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)


func main() {
	pkger.Include("/templates")

	context, err := getContext()
	if err != nil {
		log.Fatal(err)
	}

	var kustomizeApplications []*ApplicationViewModel
	var helmApplications []*ApplicationViewModel
	var projectViewModels []*ProjectViewModel

	files, err := ioutil.ReadDir(ClustersDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		clusterFile := path.Join(ClustersDir, f.Name(), ClusterFile)
		print("evaluating", clusterFile)

		if !fileExists(clusterFile) {
			print("no cluster file detected, skipping directory")
			continue
		}

		clusterConfig, err := readClusterConfig(clusterFile)
		if err != nil {
			fatal("unable to read cluster configuration:", err)
		}

		for _, app := range clusterConfig.KustomizeApplications {
			appViewModel, err := generateKustomizeApplication(app, clusterConfig, context)
			if err != nil {
				fatal("error while generating kustomize application:", err)
			}
			kustomizeApplications = append(kustomizeApplications, appViewModel)
		}

		for _, app := range clusterConfig.HelmApplications {
			argoApp, err := generateHelmApplication(app, clusterConfig, context)
			if err != nil {
				fatal("error while generating helm application:", err)
			}
			helmApplications = append(helmApplications, argoApp)
		}

		generatorApp, err := generateObjectsGeneratorApplication(clusterConfig, helmApplications)
		if err != nil {
			fatal("error while generating object generator application", err)
		}
		helmApplications = append(helmApplications, generatorApp)

		appProject, err := generateAppProject(clusterConfig)
		if err != nil {
			fatal("error while generating project:", err)
		}
		projectViewModels = append(projectViewModels, appProject)
	}

	for _, app := range kustomizeApplications {
		renderTemplate("/templates/app-kustomize.yaml", app)
	}

	for _, app := range helmApplications {
		app.Values = indent(app.Values, "        ")
		renderTemplate("/templates/app-helm.yaml", app)
	}

	for _, proj := range projectViewModels {
		renderTemplate("/templates/project.yaml", proj)
	}
}

func getContext() (*EnvironmentContext, error) {
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	repoPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	repoUrl, _ := exec.Command("git", "config", "--get", "remote.origin.url").CombinedOutput()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unable to get git remote url: %s", err))
	}

	return &EnvironmentContext{
		BasePath: basePath,
		RepoPath: repoPath,
		RepoUrl:  strings.TrimSpace(string(repoUrl)),
	}, nil
}

package main

import (
	"errors"
	"fmt"
	"github.com/markbates/pkger"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	pkger.Include("/templates")

	envClusters := os.Getenv("CLUSTERS")
	clusters := strings.Split(envClusters, ",")

	context, err := getContext()
	if err != nil {
		fatal(err)
	}

	files, err := ioutil.ReadDir(ClustersDir)
	if err != nil {
		fatal(err)
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		if len(clusters) > 0 && envClusters != "" {
			if !sliceContainsString(clusters, f.Name()) {
				continue
			}
		}

		processCluster(f.Name(), context)
	}
}

func processCluster(clusterName string, context *EnvironmentContext) {
	configFiles := getClusterConfigFiles(clusterName)

	if len(configFiles) == 0 {
		print("no config files for cluster", clusterName)
		return
	}

	var clusterConfig *ClusterConfigFile

	for _, cf := range configFiles {
		clusterConfigPart, err := readClusterConfig(cf)
		if err != nil {
			fatal("unable to read cluster configuration:", err)
		}

		if clusterConfig == nil {
			// first file should be cluster.yaml
			clusterConfig = clusterConfigPart
		} else {
			// TODO merge all fields from clusterConfigPart, including clusterConfig
			clusterConfig.KustomizeApplications = append(clusterConfig.KustomizeApplications, clusterConfigPart.KustomizeApplications...)
			clusterConfig.HelmApplications = append(clusterConfig.HelmApplications, clusterConfigPart.HelmApplications...)
			clusterConfig.PluginApplications = append(clusterConfig.PluginApplications, clusterConfigPart.PluginApplications...)
		}
	}

	var kustomizeApplications []*ApplicationViewModel
	var helmApplications []*ApplicationViewModel
	var pluginApplications []*ApplicationViewModel
	var projectViewModels []*ProjectViewModel

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

	for _, app := range clusterConfig.PluginApplications {
		pluginApp, err := generatePluginApplication(app, clusterConfig, context)
		if err != nil {
			fatal("error while generating plugin application:", err)
		}
		pluginApplications = append(pluginApplications, pluginApp)
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

	for _, app := range kustomizeApplications {
		renderTemplate("/templates/app-kustomize.yaml", app)
	}

	for _, app := range helmApplications {
		app.Values = indent(app.Values, "        ")
		renderTemplate("/templates/app-helm.yaml", app)
	}

	for _, app := range pluginApplications {
		renderTemplate("/templates/app-plugin.yaml", app)
	}

	for _, proj := range projectViewModels {
		renderTemplate("/templates/project.yaml", proj)
	}
}

func getClusterConfigFiles(clusterName string) (configFiles []string) {
	clusterFile := path.Join(ClustersDir, clusterName, ClusterFile)

	if fileExists(clusterFile) {
		configFiles = append(configFiles, clusterFile)
	}

	clusterDirPath := path.Join(ClustersDir, clusterName, ClusterConfigDir)

	if !dirExists(clusterDirPath) {
		return
	}

	files, err := ioutil.ReadDir(clusterDirPath)
	if err != nil {
		fatal(err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		configFiles = append(configFiles, path.Join(clusterDirPath, f.Name()))
	}
	return
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

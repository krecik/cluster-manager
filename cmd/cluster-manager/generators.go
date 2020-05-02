package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"strings"
)

func generateKustomizeApplication(app *KustomizeApplication, clusterConfig *ClusterConfigFile, context *EnvironmentContext) (*ApplicationViewModel, error) {
	cascadeDelete := fallbackBoolWithDefault(false, app.CascadeDelete, clusterConfig.Cluster.CascadeDelete)
	repoUrl := fallbackString(app.RepoUrl, clusterConfig.Cluster.RepoUrl, &context.RepoUrl)
	autoSync := fallbackBoolWithDefault(true, app.AutoSync, clusterConfig.Cluster.AutoSync)
	name := fallbackString(app.Name)
	namespace := fallbackStringWithDefault("default", app.Namespace, app.Name)
	targetRevision := fallbackStringWithDefault("", app.TargetRevision)

	appViewModel := &ApplicationViewModel{
		Name:           name,
		Project:        clusterConfig.Cluster.Name,
		CascadeDelete:  cascadeDelete,
		RepoUrl:        repoUrl,
		Server:         clusterConfig.Cluster.Server,
		Path:           app.Path,
		AutoSync:       autoSync,
		TargetRevision: targetRevision,
		Namespace:      namespace,
	}

	return appViewModel, nil
}

func generateHelmApplication(app *HelmApplication, clusterConfig *ClusterConfigFile, context *EnvironmentContext) (*ApplicationViewModel, error) {
	if app.Include != nil {
		includeFile := path.Join(context.RepoPath, ClustersDir, clusterConfig.Cluster.Name, *app.Include)

		bytes, err := ioutil.ReadFile(includeFile)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(bytes, &app)
		if err != nil {
			return nil, err
		}
	}

	addon := &HelmApplication{}
	if app.Addon != nil {
		clusterAddonFile := path.Join(context.RepoPath, ClustersDir, clusterConfig.Cluster.Name, AddonsDir, fmt.Sprintf("%s.yaml", *app.Addon))
		repoAddonFile := path.Join(context.RepoPath, AddonsDir, fmt.Sprintf("%s.yaml", *app.Addon))

		file := ""
		if fileExists(clusterAddonFile) {
			file = clusterAddonFile
		}
		if fileExists(repoAddonFile) {
			file = repoAddonFile
		}

		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(bytes, &addon)
		if err != nil {
			return nil, err
		}
	}

	// intentionally ignoring addon settings here
	cascadeDelete := fallbackBoolWithDefault(false, app.CascadeDelete, clusterConfig.Cluster.CascadeDelete)
	autoSync := fallbackBoolWithDefault(true, app.AutoSync, clusterConfig.Cluster.AutoSync)

	repoUrl := fallbackString(app.RepoUrl, addon.RepoUrl, clusterConfig.Cluster.RepoUrl, &context.RepoUrl)
	name := fallbackString(app.Name, addon.Name, app.Addon)
	releaseName := fallbackString(app.ReleaseName, addon.ReleaseName, app.Name, app.Addon)
	namespace := fallbackStringWithDefault("default", app.Namespace, addon.Namespace, app.Name, app.Addon)
	targetRevision := fallbackStringWithDefault("", app.TargetRevision, addon.TargetRevision)
	path := fallbackString(&app.Path, &addon.Path)

	// we merge app and addon values into app.Values
	values := mergeStructs(app.Values, addon.Values)

	valueFiles := append(app.ValueFiles, addon.ValueFiles...)
	settings := mergeDicts(addon.Settings, app.Settings)
	parameters := mergeDicts(addon.Parameters, app.Parameters)

	valuesYaml := yamlSerializeToString(values)
	for find, replace := range settings {
		valuesYaml = strings.ReplaceAll(valuesYaml, fmt.Sprintf("%%SETTINGS_%s", find), replace)
	}

	appViewModel := &ApplicationViewModel{
		Name:           name,
		Project:        clusterConfig.Cluster.Name,
		CascadeDelete:  cascadeDelete,
		RepoUrl:        repoUrl,
		Server:         clusterConfig.Cluster.Server,
		Path:           path,
		AutoSync:       autoSync,
		TargetRevision: targetRevision,
		Values:         valuesYaml,
		ValueFiles:     valueFiles,
		ReleaseName:    releaseName,
		Parameters:     parameters,
		Namespace:      namespace,
	}

	return appViewModel, nil
}

func generateObjectsGeneratorApplication(clusterConfig *ClusterConfigFile, applications []*ApplicationViewModel) (*ApplicationViewModel, error) {
	var namespaces []string
	for _, app := range applications {
		if app.Namespace != "default" || app.Namespace != "kube-system" {
			namespaces = append(namespaces, app.Namespace)
		}
	}

	values := &ObjectsGeneratorViewModel{
		Namespaces: namespaces,
	}

	valuesStr := renderTemplateToString("/templates/objects-generator-values.yaml", values)

	app := &ApplicationViewModel{
		Name:          ObjectsGeneratorAppName,
		CascadeDelete: true,
		Project:       clusterConfig.Cluster.Name,
		RepoUrl:       ObjectGeneratorRepoUrl,
		Path:          "chart",
		Values:        valuesStr,
		ReleaseName:   ObjectsGeneratorAppName,
		Server:        clusterConfig.Cluster.Server,
		Namespace:     "kube-system",
		AutoSync:      true,
	}

	return app, nil
}

func generateAppProject(config *ClusterConfigFile) (*ProjectViewModel, error) {
	project := &ProjectViewModel{
		Name:         config.Cluster.Name,
		Server:       config.Cluster.Server,
		ProjectRoles: []ProjectRole{},
	}

	return project, nil
}

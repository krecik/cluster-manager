package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"strings"
)

func generatePluginApplication(app *PluginApplication, clusterConfig *ClusterConfigFile, context *EnvironmentContext) (*ApplicationViewModel, error) {
	if app.Include != nil {
		err := loadInclude(*app.Include, clusterConfig.Cluster.Name, context, app)
		if err != nil {
			return nil, err
		}
	}

	addon := &PluginAddon{}
	if app.Addon != nil {
		err := loadAddon(*app.Addon, clusterConfig.Cluster.Name, context, addon)
		if err != nil {
			return nil, err
		}
	}

	// intentionally ignoring addon settings here
	cascadeDelete := fallbackBoolWithDefault(false, app.CascadeDelete, clusterConfig.Cluster.CascadeDelete)
	autoSync := fallbackBoolWithDefault(true, app.AutoSync, clusterConfig.Cluster.AutoSync)

	repoUrl := fallbackString(app.RepoUrl, addon.RepoUrl, clusterConfig.Cluster.RepoUrl, &context.RepoUrl)
	name := fallbackString(app.Name, addon.Name, app.Addon)
	namespace := fallbackStringWithDefault("default", app.Namespace, addon.Namespace, app.Name, app.Addon)
	targetRevision := fallbackStringWithDefault("", app.TargetRevision, addon.TargetRevision)
	path := fallbackString(&app.Path, &addon.Path)

	pluginName := fallbackString(&app.PluginName, &addon.PluginName)
	pluginEnv := mergeDicts(addon.PluginEnv, app.PluginEnv)

	appViewModel := &ApplicationViewModel{
		Name:           name,
		Project:        clusterConfig.Cluster.Name,
		CascadeDelete:  cascadeDelete,
		RepoUrl:        repoUrl,
		Server:         clusterConfig.Cluster.Server,
		Path:           path,
		AutoSync:       autoSync,
		TargetRevision: targetRevision,
		Namespace:      namespace,
		PluginName:     pluginName,
		PluginEnv:      pluginEnv,
	}

	return appViewModel, nil
}

func generateKustomizeApplication(app *KustomizeApplication, clusterConfig *ClusterConfigFile, context *EnvironmentContext) (*ApplicationViewModel, error) {
	if app.Include != nil {
		err := loadInclude(*app.Include, clusterConfig.Cluster.Name, context, app)
		if err != nil {
			return nil, err
		}
	}

	addon := &KustomizeAddon{}
	if app.Addon != nil {
		err := loadAddon(*app.Addon, clusterConfig.Cluster.Name, context, addon)
		if err != nil {
			return nil, err
		}
	}

	// intentionally ignoring addon settings here
	cascadeDelete := fallbackBoolWithDefault(false, app.CascadeDelete, clusterConfig.Cluster.CascadeDelete)
	autoSync := fallbackBoolWithDefault(true, app.AutoSync, clusterConfig.Cluster.AutoSync)

	repoUrl := fallbackString(app.RepoUrl, addon.RepoUrl, clusterConfig.Cluster.RepoUrl, &context.RepoUrl)
	name := fallbackString(app.Name, addon.Name, app.Addon)
	namespace := fallbackStringWithDefault("default", app.Namespace, addon.Namespace, app.Name, app.Addon)
	targetRevision := fallbackStringWithDefault("", app.TargetRevision, addon.TargetRevision)
	path := fallbackString(&app.Path, &addon.Path)

	appViewModel := &ApplicationViewModel{
		Name:           name,
		Project:        clusterConfig.Cluster.Name,
		CascadeDelete:  cascadeDelete,
		RepoUrl:        repoUrl,
		Server:         clusterConfig.Cluster.Server,
		Path:           path,
		AutoSync:       autoSync,
		TargetRevision: targetRevision,
		Namespace:      namespace,
	}

	return appViewModel, nil
}

func generateHelmApplication(app *HelmApplication, clusterConfig *ClusterConfigFile, context *EnvironmentContext) (*ApplicationViewModel, error) {
	if app.Include != nil {
		err := loadInclude(*app.Include, clusterConfig.Cluster.Name, context, app)
		if err != nil {
			return nil, err
		}
	}

	addon := &HelmAddon{}
	if app.Addon != nil {
		err := loadAddon(*app.Addon, clusterConfig.Cluster.Name, context, addon)
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
	oauth2ProxyIngressHost := fallbackStringWithDefault("", app.Oauth2ProxyIngressHost, addon.Oauth2ProxyIngressHost)
	path := fallbackString(&app.Path, &addon.Path)

	// we merge app and addon values into app.Values
	values := mergeStructs(app.Values, addon.Values)

	if addon.OverlayDefinitions != nil {
		for _, overlay := range app.Overlays {
			overlayDefinition, ok := addon.OverlayDefinitions[overlay]
			if !ok {
				continue
			}
			values = mergeStructs(values, overlayDefinition.Values)

			if overlayDefinition.Oauth2ProxyIngressHost != nil {
				oauth2ProxyIngressHost = *overlayDefinition.Oauth2ProxyIngressHost
			}
		}
	}

	valueFiles := append(app.ValueFiles, addon.ValueFiles...)
	settings := mergeDicts(addon.Settings, clusterConfig.Cluster.Settings, app.Settings)
	parameters := mergeDicts(addon.Parameters, app.Parameters)

	valuesYaml := yamlSerializeToString(values)
	for i := 0; i < len(settings); i++ { // run multiple times so settings can refer to other settings
		for find, replace := range settings {
			findFmt := fmt.Sprintf("%%SETTINGS_%s", find)
			valuesYaml = strings.ReplaceAll(valuesYaml, findFmt, replace)
			// we allow using settings in oauth2ProxyIngressHost for convenience
			oauth2ProxyIngressHost = strings.ReplaceAll(oauth2ProxyIngressHost, findFmt, replace)
		}
	}

	appViewModel := &ApplicationViewModel{
		Name:                   name,
		Project:                clusterConfig.Cluster.Name,
		CascadeDelete:          cascadeDelete,
		RepoUrl:                repoUrl,
		Server:                 clusterConfig.Cluster.Server,
		Path:                   path,
		AutoSync:               autoSync,
		TargetRevision:         targetRevision,
		Values:                 valuesYaml,
		ValueFiles:             valueFiles,
		ReleaseName:            releaseName,
		Parameters:             parameters,
		Namespace:              namespace,
		OAuth2ProxyIngressHost: oauth2ProxyIngressHost,
	}

	return appViewModel, nil
}

func generateObjectsGeneratorApplication(clusterConfig *ClusterConfigFile, applications []*ApplicationViewModel) (*ApplicationViewModel, error) {
	var namespaces []string
	oauth2ProxyIngresses := []Oauth2ProxyIngress{}
	autoSync := fallbackBoolWithDefault(true, clusterConfig.Cluster.AutoSync)

	for _, app := range applications {
		if app.Namespace != "default" && app.Namespace != "kube-system" {
			if !sliceContainsString(namespaces, app.Namespace) {
				namespaces = append(namespaces, app.Namespace)
			}
		}

		if app.OAuth2ProxyIngressHost != "" {
			oauth2ProxyIngresses = append(oauth2ProxyIngresses, Oauth2ProxyIngress{
				Name:      app.Name,
				Namespace: app.Namespace,
				Host:      app.OAuth2ProxyIngressHost,
			})
		}
	}

	values := &ObjectsGeneratorViewModel{
		Namespaces:           namespaces,
		Oauth2ProxyIngresses: oauth2ProxyIngresses,
	}

	valuesStr := renderTemplateToString("/templates/objects-generator-values.yaml", values)
	for i := 0; i < len(clusterConfig.Cluster.Settings); i++ { // run multiple times so settings can refer to other settings
		for find, replace := range clusterConfig.Cluster.Settings {
			findFmt := fmt.Sprintf("%%SETTINGS_%s", find)
			valuesStr = strings.ReplaceAll(valuesStr, findFmt, replace)
		}
	}

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
		AutoSync:      autoSync,
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

func loadAddon(addon string, clusterName string, context *EnvironmentContext, out interface{}) error {
	baseAddonFile := path.Join(context.BasePath, AddonsDir, fmt.Sprintf("%s.yaml", addon))
	clusterAddonFile := path.Join(context.RepoPath, ClustersDir, clusterName, AddonsDir, fmt.Sprintf("%s.yaml", addon))
	repoAddonFile := path.Join(context.RepoPath, AddonsDir, fmt.Sprintf("%s.yaml", addon))

	file := ""
	if fileExists(clusterAddonFile) {
		file = clusterAddonFile
	} else if fileExists(repoAddonFile) {
		file = repoAddonFile
	} else if fileExists(baseAddonFile) {
		file = baseAddonFile
	}

	if file == "" {
		return errors.New(fmt.Sprintf("unable to load Helm addon file: %s", addon))
	}

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bytes, out)
	if err != nil {
		return err
	}

	return nil
}

func loadInclude(filename string, clusterName string, context *EnvironmentContext, out interface{}) error {
	includeFile := path.Join(context.RepoPath, ClustersDir, clusterName, filename)

	bytes, err := ioutil.ReadFile(includeFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bytes, out)
	if err != nil {
		return err
	}

	return nil
}

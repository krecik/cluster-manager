package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func readClusterConfig(path string) (*ClusterConfigFile, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ClusterConfigFile
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func yamlSerializeToString(in interface{}) string {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

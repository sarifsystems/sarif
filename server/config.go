package main

import (
	"io/ioutil"
	"os"
	"encoding/json"
)

var config map[string]interface{}

func readConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}
}

func getConfig(field string) interface{} {
	if config == nil {
		readConfig()
	}
	return config[field]
}

func getConfigMap(field string) map[string]interface{} {
	if config == nil {
		readConfig()
	}
	val, _ := config[field].(map[string]interface{})
	return val
}

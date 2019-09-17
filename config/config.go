package config

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var regexpsForConfig map[string]string = make(map[string]string)

func init() {
	// build regexps map for parse config file
	regexpsForConfig["interval"] = "interval=(\\d+)"
	regexpsForConfig["port"] = "port=([\\w-]+)"
	regexpsForConfig["format"] = "format=([\\w-%.]+)"
	regexpsForConfig["metrics"] = "metrics=([\\w.:/]+)"
	regexpsForConfig["elastic"] = "elastic=([\\w.:/]+)"
}

type SecretServiceConfig struct {
	Interval           int
	Port               string
	MetricbeatTemplate string
	MetricsWebAddress  string
	ElasticWebAddress  string
}

func ReadFile(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("error when reading file:", err)
	}
	return string(bytes)
}

func parseConfigFile(configFileContent string) *SecretServiceConfig {
	secretServiceConfig := SecretServiceConfig{}
	for regexpName, expression := range regexpsForConfig {
		r := regexp.MustCompile(expression)
		submatches := r.FindStringSubmatch(configFileContent)
		if len(submatches) == 0 {
			log.Println("something wrong with parsing config file (check that all fields are specified)")
			os.Exit(1)
		}
		if regexpName == "interval" {
			interval, err := strconv.Atoi(submatches[1])
			if err != nil {
				log.Println("error when converting secretServiceConfig.Interval to int:", err)
			}
			secretServiceConfig.Interval = interval
		} else if regexpName == "port" {
			secretServiceConfig.Port = submatches[1]
		} else if regexpName == "format" {
			secretServiceConfig.MetricbeatTemplate = submatches[1]
		} else if regexpName == "metrics" {
			secretServiceConfig.MetricsWebAddress = submatches[1]
		} else if regexpName == "elastic" {
			secretServiceConfig.ElasticWebAddress = submatches[1]
		}
	}
	return &secretServiceConfig
}

func DeleteCommentsFromConfig(configFileContent string) string {
	lines := strings.Split(configFileContent, "\n")
	for index, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			lines[index] = ""
		}
	}
	return strings.Join(lines, "\n")
}

func ParseConfigFile(configFileContent string) *SecretServiceConfig {
	configFileContentWithCommentsDeleted := DeleteCommentsFromConfig(configFileContent)
	secretServiceConfig := parseConfigFile(configFileContentWithCommentsDeleted)
	return secretServiceConfig
}

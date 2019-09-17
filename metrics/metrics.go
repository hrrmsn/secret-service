package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var hostname string
var err error

type Metrics struct {
	Version    interface{}            `json:"version"`
	Gauges     map[string]interface{} `json:"gauges"`
	Counters   map[string]interface{} `json:"counters"`
	Histograms map[string]interface{} `json:"histograms"`
	Meters     map[string]interface{} `json:"meters"`
	Timers     map[string]interface{} `json:"timers"`
}

func (m *Metrics) MakeSecretServiceData(metricsetName string) map[string]interface{} {
	s := make(map[string]interface{})
	if metricsetName == "version" {
		s["version"] = m.Version
		return s
	} else if metricsetName == "gauges" {
		s["gauges"] = m.Gauges
		return s
	} else if metricsetName == "counters" {
		s["counters"] = m.Counters
		return s
	} else if metricsetName == "histograms" {
		s["histograms"] = m.Histograms
		return s
	} else if metricsetName == "meters" {
		s["meters"] = m.Meters
		return s
	}
	s["timers"] = m.Timers
	return s
}

type ElasticMetricFormat struct {
	Index  string `json:"_index"`
	Type   string `json:"_type"`
	Source Source `json:"_source"`
}

type Source struct {
	Timestamp     string                 `json:"@timestamp"`
	Metricset     Metricset              `json:"metricset"`
	SecretService map[string]interface{} `json:"secretservice"`
	Beat          Beat                   `json:"beat"`
}

type Metricset struct {
	Name   string `json:"name"`
	Module string `json:"module"`
}

type Beat struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

func (m *Metrics) MakeJSONAndElasticMetric(metricsetName, metricbeatTemplate string) (string, *ElasticMetricFormat) {
	elasticMetric := MakeElasticMetric(metricbeatTemplate)
	elasticMetric.Source.Metricset.Name = metricsetName
	elasticMetric.Source.SecretService = m.MakeSecretServiceData(metricsetName)
	json, err := json.Marshal(elasticMetric)
	if err != nil {
		log.Println("error when marshalling json from ElasticMetricFormat:", err)
	}
	return string(json), elasticMetric
}

func (m *Metrics) ConvertToElasticFormatWithJSONs(metricbeatTemplate string) (map[string]string,
	[]*ElasticMetricFormat) {

	elasticMetrics := []*ElasticMetricFormat{}
	var elasticMetric *ElasticMetricFormat
	var jsonString string
	jsons := make(map[string]string)
	for _, metricsetName := range []string{"version", "gauges", "counters", "histograms", "meters", "timers"} {
		jsonString, elasticMetric = m.MakeJSONAndElasticMetric(metricsetName, metricbeatTemplate)
		elasticMetrics = append(elasticMetrics, elasticMetric)
		jsons[metricsetName] = jsonString
	}
	return jsons, elasticMetrics
}

func Get(metricsWebAddress string) *Metrics {
	resp, err := http.Get(metricsWebAddress)
	if err != nil {
		log.Println("error when retreiving metrics by http:", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error when reading body of the http response:", err)
	}

	// read Metrics from body
	metrics := Metrics{}
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		log.Println("error when unmarshalling json from http response:", err)
	}
	return &metrics
}

func MakeIndexAndTimestamp(indexTemplate string) (string, string) {
	currentTime := time.Now().UTC()
	currentTimeFormatted := currentTime.Format("2006.01.02")
	return fmt.Sprintf(indexTemplate, currentTimeFormatted), currentTime.Format("2006-01-02T15:04:05.000Z")
}

func GetHostname() (string, error) {
	if len(hostname) > 0 {
		return hostname, nil
	}
	hostname, err = os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

func MakeElasticMetric(metricbeatTemplate string) *ElasticMetricFormat {
	index, timestamp := MakeIndexAndTimestamp(metricbeatTemplate)
	elasticMetric := ElasticMetricFormat{Index: index, Type: "doc"}

	hostname, err := GetHostname()
	if err != nil {
		log.Println("error when getting hostname:", err)
	}
	source := Source{
		Timestamp: timestamp,
		Metricset: Metricset{Module: "secretservice"},
		Beat:      Beat{Name: hostname, Hostname: hostname, Version: "6.2.4"},
	}

	elasticMetric.Source = source
	return &elasticMetric
}

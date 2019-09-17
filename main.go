package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olivere/elastic"

	"noosphere.foundation/secret-service/config"
	"noosphere.foundation/secret-service/metrics"
)

var logFile *os.File
var elasticClient *elastic.Client
var secretServiceConfig *config.SecretServiceConfig

var logFileName string = "log.txt"
var configFileName string = "config.cfg"
var ctx context.Context = context.Background()

func init() {
	// init file logging
	var err error
	logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)

	configFileContent := config.ReadFile(configFileName)
	secretServiceConfig = config.ParseConfigFile(configFileContent)

	// connect to elastic
	log.Printf("connecting to elastic (%s)...\n", secretServiceConfig.ElasticWebAddress)
	elasticClient, err = elastic.NewClient(
		elastic.SetURL(secretServiceConfig.ElasticWebAddress),
		elastic.SetSniff(false))
	if err != nil {
		log.Fatal("error when connecting to elastic:", err)
	}
	log.Println("ok")
}

func handleBulkResponse(blk *elastic.BulkResponse) {
	if blk.Errors {
		for _, m := range blk.Items {
			for key, value := range m {
				if value != nil && value.Error != nil {
					log.Println(key + "->" + value.Error.Type)
					log.Println(value.Error.Reason)
				}
				log.Println()
				log.Println()
			}
		}
	} else {
		log.Println("bulk request didn't return some errors")
	}
}

func SendMetricsToElastic(metricsWebAddress, metricbeatTemplate string) {
	// prepare data to send
	log.Println("retreiving metrics...")
	mtr := metrics.Get(metricsWebAddress)
	_, elasticMetrics := mtr.ConvertToElasticFormatWithJSONs(metricbeatTemplate)
	log.Println("ok")

	log.Println("preparing for bulk insertion...")
	index, _ := metrics.MakeIndexAndTimestamp(metricbeatTemplate)
	bulk := elasticClient.Bulk().Index(index).Type("doc")

	for _, elasticMetric := range elasticMetrics {
		bulk.Add(elastic.NewBulkIndexRequest().Doc(elasticMetric.Source))
	}
	log.Println("ok")

	// send
	log.Println("sending data to elastic...")
	var err error
	var bulkResponse *elastic.BulkResponse
	if bulkResponse, err = bulk.Do(ctx); err != nil {
		log.Fatal("error when performing bulk insertion:", err)
	}
	log.Println("ok")

	handleBulkResponse(bulkResponse)
}

func cleanupBeforeExit() {
	// cleanup actions
	defer logFile.Close()

	log.Println("handling ctrl-c interruption manually")
	log.Println("program will be terminated now")
	log.Println()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	mtr := metrics.Get(secretServiceConfig.MetricsWebAddress)
	jsons, _ := mtr.ConvertToElasticFormatWithJSONs(secretServiceConfig.MetricbeatTemplate)

	fmt.Fprintln(w, "["+jsons["version"]+",\n")
	fmt.Fprintln(w, jsons["gauges"]+",\n")
	fmt.Fprintln(w, jsons["counters"]+",\n")
	fmt.Fprintln(w, jsons["histograms"]+",\n")
	fmt.Fprintln(w, jsons["meters"]+",\n")
	fmt.Fprintln(w, jsons["timers"]+"]")
}

func main() {
	// close log file
	defer log.Println()
	defer logFile.Close()

	// handle ctrl-c interruption
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanupBeforeExit()
		os.Exit(1)
	}()

	ticker := time.NewTicker(time.Second * time.Duration(secretServiceConfig.Interval))
	go func() {
		for t := range ticker.C {
			SendMetricsToElastic(secretServiceConfig.MetricsWebAddress, secretServiceConfig.MetricbeatTemplate)
			log.Println("tick at", t)
			log.Println()
		}
	}()

	log.Println("listening on port " + secretServiceConfig.Port + "...")
	http.HandleFunc("/", Handler)
	log.Fatal(http.ListenAndServe(":"+secretServiceConfig.Port, nil))
}

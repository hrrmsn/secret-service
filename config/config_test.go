package config

import (
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	configFileContent := ReadFile("config_test.cfg")
	secretServiceConfig := ParseConfigFile(configFileContent)

	if secretServiceConfig.Interval != 123 {
		t.Error("expected interval is 123, but found", secretServiceConfig.Interval)
	}
	if secretServiceConfig.Port != "8081" {
		t.Error("expected port is 8081, but found", secretServiceConfig.Port)
	}
	if secretServiceConfig.MetricbeatTemplate != "metricbeat-6.2.4-%s" {
		t.Errorf(`expected template is metricbeat-6.2.4-%s, but found %s`, "%s", secretServiceConfig.MetricbeatTemplate)
	}
	if secretServiceConfig.MetricsWebAddress != "http://IP_ADDRESS:PORT/metrics" {
		t.Error(`expected metrics web address is http://IP_ADDRESS:PORT/metrics, but found`,
			secretServiceConfig.MetricsWebAddress)
	}
	if secretServiceConfig.ElasticWebAddress != "http://ELASTIC_ADDR:PORT/" {
		t.Error(`expected elastic web address is http://ELASTIC_ADDR:PORT/, but found`,
			secretServiceConfig.ElasticWebAddress)
	}
}

func TestDeleteCommentsFromConfig(t *testing.T) {
	configFileContent := ReadFile("config_with_comments_test.cfg")
	secretServiceConfig := ParseConfigFile(configFileContent)

	if secretServiceConfig.Interval != 256 {
		t.Error("expected interval is 256, but found", secretServiceConfig.Interval)
	}
	if secretServiceConfig.Port != "7867" {
		t.Error("expected port is 7867, but found", secretServiceConfig.Port)
	}
	if secretServiceConfig.MetricbeatTemplate != "metricbeat-1.2.3-%s" {
		t.Errorf(`expected template is metricbeat-1.2.3-%s, but found %s`, "%s", secretServiceConfig.MetricbeatTemplate)
	}
	if secretServiceConfig.MetricsWebAddress != "http://localhost:8080/metrics_test" {
		t.Error(`expected metrics web address is http://localhost:8080/metrics_test, but found`,
			secretServiceConfig.MetricsWebAddress)
	}
	if secretServiceConfig.ElasticWebAddress != "http://my.company.com:9200/" {
		t.Error(`expected elastic web address is elastic=http://my.company.com:9200/, but found`,
			secretServiceConfig.ElasticWebAddress)
	}
}

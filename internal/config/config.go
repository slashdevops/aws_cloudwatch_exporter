/*
Copyright © 2020 Christian González Di Antonio christian@slashdevops.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

type All struct {
	ServerConf            `mapstructure:",squash"`
	ApplicationConf       `mapstructure:",squash"`
	MetricDataQueriesConf `mapstructure:",squash"`
}

func (c *All) ToJSON() string {
	out, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (c *All) ToYAML() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// This is a convenient structure to allow config files nested (server.[keys])
// server.conf file
// server:
//
//	address:
//	port:
type ServerConf struct {
	Server `mapstructure:"server" json:"server" yaml:"server"`
}

type Server struct {
	Address           string        `mapstructure:"address" json:"address" yaml:"address"`
	Port              uint16        `mapstructure:"port" json:"port" yaml:"port"`
	ReadTimeout       time.Duration `mapstructure:"readTimeout" json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout      time.Duration `mapstructure:"writeTimeout" json:"writeTimeout" yaml:"writeTimeout"`
	IdleTimeout       time.Duration `mapstructure:"idleTimeout" json:"idleTimeout" yaml:"idleTimeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"readHeaderTimeout" json:"readHeaderTimeout" yaml:"readHeaderTimeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdownTimeout" json:"shutdownTimeout" yaml:"shutdownTimeout"`
	KeepAlivesEnabled bool          `mapstructure:"keepAlivesEnabled" json:"keepAlivesEnabled" yaml:"keepAlivesEnabled"`
	LogFormat         string        `mapstructure:"logFormat" json:"logFormat" yaml:"logFormat"`
	Debug             bool          `mapstructure:"debug" json:"debug" yaml:"debug"`
}

// This is a convenient structure to allow config files nested (application.[keys])
// No File conf
// application:
//
//	name:
//	description:
type ApplicationConf struct {
	Application `mapstructure:"application" json:"application" yaml:"application"`
}

type Application struct {
	Name             string   `json:"name" yaml:"name"`
	Description      string   `json:"description" yaml:"description"`
	GitRepository    string   `json:"gitRepository" yaml:"gitRepository"`
	Version          string   `json:"version" yaml:"version"`
	Revision         string   `json:"revision" yaml:"revision"`
	Branch           string   `json:"branch" yaml:"branch"`
	BuildUser        string   `json:"buildUser" yaml:"buildUser"`
	BuildDate        string   `json:"buildDate" yaml:"buildDate"`
	GoVersion        string   `json:"goVersion" yaml:"goVersion"`
	VersionInfo      string   `json:"versionInfo" yaml:"versionInfo"`
	BuildInfo        string   `json:"buildInfo" yaml:"buildInfo"`
	ServerFile       string   `mapstructure:"serverFile" json:"serverFile" yaml:"serverFile"`
	HealthPath       string   `json:"healthPath" yaml:"healthPath"`
	MetricsPath      string   `json:"metricsPath" yaml:"metricsPath"`
	MetricsFiles     []string `mapstructure:"metricsFiles" json:"metricsFiles" yaml:"metricsFiles"`
	MetricStatPeriod string   `mapstructure:"metricStatPeriod" json:"metricStatPeriod" yaml:"metricStatPeriod"`
	MetricTimeWindow string   `mapstructure:"metricTimeWindow" json:"metricTimeWindow" yaml:"metricTimeWindow"`
}

// This is a convenient structure to allow config files nested (MetricDataQueries.[keys])
// File conf metrics.yaml
// Will be filled with que Metrics Queries
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricStat.html
// NOTE: This structure is nested because I don't use its internal structure in anywhere
type MetricDataQueriesConf struct {
	MetricDataQueries []MetricDataQuery `mapstructure:"MetricDataQueries" json:"MetricDataQueries" yaml:"MetricDataQueries"`
}

type MetricDataQuery struct {
	ID         string `mapstructure:"Id" json:"Id" yaml:"Id"`
	MetricStat struct {
		Metric struct {
			Namespace  string `mapstructure:"Namespace" json:"Namespace" yaml:"Namespace"`
			MetricName string `mapstructure:"MetricName" json:"MetricName" yaml:"MetricName"`
			Dimensions []struct {
				Name  string `mapstructure:"Name" json:"Name" yaml:"Name"`
				Value string `mapstructure:"Value" json:"Value" yaml:"Value"`
			} `mapstructure:"Dimensions" json:"Dimensions" yaml:"Dimensions"`
		} `mapstructure:"Metric" json:"Metric" yaml:"Metric"`
		Period int64  `mapstructure:"Period" json:"Period" yaml:"Period"`
		Stat   string `mapstructure:"Stat" json:"Stat" yaml:"Stat"`
		Unit   string `mapstructure:"Unit" json:"Unit" yaml:"Unit"`
	} `mapstructure:"MetricStat" json:"MetricStat" yaml:"MetricStat"`
}

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
package camelcase

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantEntries [][]string
	}{
		{
			name: "Original Test Cases",
			args: []string{
				"",
				"lowercase",
				"Class",
				"MyClass",
				"MyC",
				"HTML",
				"PDFLoader",
				"AString",
				"SimpleXMLParser",
				"vimRPCPlugin",
				"GL11Version",
				"99Bottles",
				"May5",
				"BFG9000",
				"BöseÜberraschung",
				"Two  spaces",
				"BadUTF8\xe2\xe2\xa1",
			},
			wantEntries: [][]string{
				{""},
				{"lowercase"},
				{"Class"},
				{"My", "Class"},
				{"My", "C"},
				{"HTML"},
				{"PDF", "Loader"},
				{"A", "String"},
				{"Simple", "XML", "Parser"},
				{"vim", "RPC", "Plugin"},
				{"GL", "11", "Version"},
				{"99", "Bottles"},
				{"May", "5"},
				{"BFG", "9000"},
				{"Böse", "Überraschung"},
				{"Two", "  ", "spaces"},
				{"BadUTF8\xe2\xe2\xa1"},
			},
		},
		{
			name: "AWS Metrics Names Test Cases",
			args: []string{
				"NetworkOut",
				"CPUUtilization",
				"AutoScalingGroupName",
				"AWS/ApiGateway",
				"AWS/ElasticBeanstalk",
			},
			wantEntries: [][]string{
				{"Network", "Out"},
				{"CPU", "Utilization"},
				{"Auto", "Scaling", "Group", "Name"},
				{"AWS", "/", "Api", "Gateway"},
				{"AWS", "/", "Elastic", "Beanstalk"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, arg := range tt.args {
				// reflect.DeepEqual []string{} different []string(nil)
				if len(arg) != 0 && len(tt.wantEntries[i]) != 0 {
					if gotEntries := Split(arg); !reflect.DeepEqual(gotEntries, tt.wantEntries[i]) {
						t.Errorf("Split() = '%v', want '%v'", gotEntries, tt.wantEntries[i])
					}
				}
			}
		})
	}
}

func TestSplitToLower(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantEntries [][]string
	}{
		{
			name: "AWS Metrics Names Test Cases",
			args: []string{
				"NetworkOut",
				"CPUUtilization",
				"AutoScalingGroupName",
				"AWS/ApiGateway",
				"AWS/ElasticBeanstalk",
			},
			wantEntries: [][]string{
				{"network", "out"},
				{"cpu", "utilization"},
				{"auto", "scaling", "group", "name"},
				{"aws", "/", "api", "gateway"},
				{"aws", "/", "elastic", "beanstalk"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, arg := range tt.args {
				// reflect.DeepEqual []string{} different []string(nil)
				if len(arg) != 0 && len(tt.wantEntries[i]) != 0 {
					if gotEntries := SplitToLower(arg); !reflect.DeepEqual(gotEntries, tt.wantEntries[i]) {
						t.Errorf("SplitToLower() = '%v', want '%v'", gotEntries, tt.wantEntries[i])
					}
				}
			}
		})
	}
}

func TestToSnake(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantEntries []string
	}{
		{
			name: "Original Test Cases",
			args: []string{
				"",
				"lowercase",
				"Class",
				"MyClass",
				"MyC",
				"HTML",
				"PDFLoader",
				"AString",
				"SimpleXMLParser",
				"vimRPCPlugin",
				"GL11Version",
				"99Bottles",
				"May5",
				"BFG9000",
				"BöseÜberraschung",
				"Two  spaces",
				"BadUTF8\xe2\xe2\xa1",
			},
			wantEntries: []string{
				"",
				"lowercase",
				"class",
				"my_class",
				"my_c",
				"html",
				"pdf_loader",
				"a_string",
				"simple_xml_parser",
				"vim_rpc_plugin",
				"gl_11_version",
				"99_bottles",
				"may_5",
				"bfg_9000",
				"",
				"two_spaces",
				"",
			},
		},
		{
			name: "AWS Metrics Names Test Cases",
			args: []string{
				"NetworkOut",
				"CPUUtilization",
				"AutoScalingGroupName",
				"AWS/ApiGateway",
				"AWS/ElasticBeanstalk",
			},
			wantEntries: []string{
				"network_out",
				"cpu_utilization",
				"auto_scaling_group_name",
				"aws_api_gateway",
				"aws_elastic_beanstalk",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, arg := range tt.args {
				if gotEntries := ToSnake(arg); !reflect.DeepEqual(gotEntries, tt.wantEntries[i]) {
					t.Errorf("ToSnake() = '%v', want '%v'", gotEntries, tt.wantEntries[i])
				}
			}
		})
	}
}

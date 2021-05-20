// Copyright Contributors to the Open Cluster Management project

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"sigs.k8s.io/yaml"
)

type Plan struct {
	Id   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

type Creator struct {
	Id   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

type Subscription struct {
	Id                  string   `json:"id"`
	Kind                string   `json:"kind"`
	Href                string   `json:"href"`
	Plan                Plan     `json:"plan"`
	Cluster_id          string   `json:"cluster_id"`
	External_cluster_id string   `json:"external_cluster_id"`
	Organization_id     string   `json:"organization_id"`
	Last_telemetry_date string   `json:"last_telemetry_date"`
	Created_at          string   `json:"created_at"`
	Updated_at          string   `json:"updated_at"`
	Support_level       string   `json:"support_level"`
	Display_name        string   `json:"display_name"`
	Creator             Creator  `json:"creator"`
	Managed             bool     `json:"managed"`
	Status              string   `json:"status"`
	Provenance          string   `json:"provenance"`
	Last_reconcile_date string   `json:"last_reconcile_date"`
	Console_url         string   `json:"console_url"`
	Last_released_at    string   `json:"last_released_at"`
	Metrics             []Metric `json:"metrics"`
	Cloud_provider_id   string   `json:"cloud_provider_id"`
	Region_id           string   `json:"region_id"`
	Trial_end_date      string   `json:"trial_end_date"`
}

type NodeDet struct {
	Updated_timestamp string  `json:"updated_timestamp"`
	Used              Couplet `json:"used"`
	Total             Couplet `json:"total"`
}

type Couplet struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

type Node struct {
	Total   int `json:"total"`
	Master  int `json:"master"`
	Compute int `json:"compute"`
}

type Upgrade struct {
	Updated_timestamp string `json:"updated_timestamp"`
	Available         bool   `json:"available"`
}

type Metric struct {
	Health_state                   string  `json:"health_state"`
	Memory                         NodeDet `json:"memory"`
	Cpu                            NodeDet `json:"cpu"`
	Sockets                        NodeDet `json:"sockets"`
	Compute_nodes_memory           NodeDet `json:"compute_nodes_memory"`
	Compute_nodes_cpu              NodeDet `json:"compute_nodes_cpu"`
	Compute_nodes_sockets          NodeDet `json:"compute_nodes_sockets"`
	Storage                        NodeDet `json:"storage"`
	Nodes                          Node    `json:"nodes"`
	Operating_system               string  `json:"operating_system"`
	Upgrade                        Upgrade `json:"upgrade"`
	State                          string  `json:"state"`
	State_description              string  `json:"state_description"`
	Openshift_version              string  `json:"openshift_version"`
	Cloud_provider                 string  `json:"cloud_provider"`
	Region                         string  `json:"region"`
	Console_url                    string  `json:"console_url"`
	Critical_alerts_firing         int     `json:"critical_alerts_firing"`
	Operators_condition_failing    int     `json:"operators_condition_failing"`
	Subscription_cpu_total         int     `json:"subscription_cpu_total"`
	Subscription_socket_total      int     `json:"subscription_socket_total"`
	Subscription_obligation_exists int     `json:"subscription_obligation_exists"`
	Cluster_type                   string  `json:"cluster_type"`
}
type SubscriptionList struct {
	Kind  string         `json:"kind"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Total int            `json:"total"`
	Items []Subscription `json:"items"`
}

type ManagedCluster struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name   string `yaml:"name"`
	Labels Label  `yaml:"labels"`
}
type Label struct {
	LocalCluster string `yaml:"local-cluster"`
	Cloud        string `yaml:"cloud"`
	Vendor       string `yaml:"vendor"`
	ClusterID    string `yaml:"clusterID"`
}
type Spec struct {
	HubAcceptsClient     bool `yaml:"hubAcceptsClient"`
	LeaseDurationSeconds int  `yaml:"leaseDurationSeconds"`
}

func main() {

	input := flag.String("input", "./testserver/data/scenarios/onek_clusters/subscription_response.json", "File location to save location of input discovered clusters file")
	output := flag.String("output", "./testserver/data/sample_managed_clusters.yaml", "File location to save output of managed clusters")
	var numFlag = flag.Int("tot", 1, "The total number of managedclusters will be created. Default is 50")
	flag.Parse()
	e := os.Remove(*output)
	if e != nil {
		fmt.Println("No file was deleted")
	}
	fo, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	file, _ := ioutil.ReadFile(*input)

	data := SubscriptionList{}

	_ = json.Unmarshal([]byte(file), &data)
	if *numFlag > data.Total {
		panic("Managed clusters must be fewer than the number of discovered clusters")
	}
	s := Spec{true, 60}
	buffer := []byte("---\n")
	if _, err := fo.Write(buffer); err != nil {
		panic(err)
	}
	for i := 0; i < *numFlag; i++ {
		l := Label{"false", "auto-detect", "auto-detect", data.Items[i].External_cluster_id}
		m := Metadata{"testmc" + strconv.Itoa(i), l}
		mc := ManagedCluster{"cluster.open-cluster-management.io/v1", "ManagedCluster", m, s}
		b, _ := yaml.Marshal(mc)
		if _, err := fo.Write(b); err != nil {
			panic(err)
		}
		if _, err := fo.Write(buffer); err != nil {
			panic(err)
		}

	}
}

// Copyright Contributors to the Open Cluster Management project

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"
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

var loweralphanum = []rune("abcdefghijklmnopqrstuvwxyz123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = loweralphanum[rand.Intn(len(loweralphanum))] // #nosec G404 (cryptographic strength irrelevant)
	}
	return string(b)
}

func collectRandomId(n int, l int) []string {
	var m = make(map[string]bool)
	var a = []string{}
	for len(a) < n {
		id := randSeq(l)
		if m[id] {
			continue // Already in the map
		}
		a = append(a, id)
		m[id] = true
	}
	return a

}

func collectRandomExtId(n int) []string {
	var m = make(map[string]bool)
	var a = []string{}
	for len(a) < n {
		id := randSeq(8) + "-" + randSeq(8) + "-" + randSeq(4) + "-" + randSeq(4) + "-" + randSeq(4) + "-" + randSeq(10)
		if m[id] {
			continue // Already in the map
		}
		a = append(a, id)
		m[id] = true
	}
	return a

}

// randomly generates a list of n Subscriptions
func generateSubscriptions(n int) []Subscription {
	ids := collectRandomId(n, 27)
	clusterIds := collectRandomId(n, 32)
	extIds := collectRandomExtId(n)
	var test []Subscription
	var mets []Metric
	c := Creator{"1Yu9TMhpDebs1S6wjLPIgYLlOn4", "Account", "/api/accounts_mgmt/v1/accounts/1Yu9TMhpDebs1S6wjLPIgYLlOn4"}
	p := Plan{"OCP", "Plan", "/api/accounts_mgmt/v1/plans/OCP"}
	upgrade_a := Upgrade{"2021-02-23T21:50:27.431Z", true}
	couplet_a := Couplet{0, ""}
	node_a := Node{6, 3, 3}
	node_det_a := NodeDet{"0001-01-01T00:00:00Z", couplet_a, couplet_a}
	metric := Metric{"healthy", node_det_a, node_det_a, node_det_a, node_det_a, node_det_a, node_det_a, node_det_a, node_a, "", upgrade_a, "ready", "", "4.6.9", "aws", "us-east-1", "https://console-openshift-console.apps.jdgray-kfmxg.dev01.red-chesterfield.com", 0, 0, 6, 3, 2, ""}
	mets = append(mets, metric)
	for i := 0; i < n; i++ {
		now := time.Now()
		floor := now.AddDate(0, 0, -10) // Date from last ten days
		maxDelta := now.Sub(floor)
		randomCreatedDelta := rand.Int63n(int64(maxDelta.Nanoseconds())) // #nosec G404 (cryptographic strength irrelevant)
		createdDate := floor.Add(time.Nanosecond * time.Duration(randomCreatedDelta))
		updateDelta := now.Sub(createdDate)
		// randomUpdatedDelta := rand.Int63n(int64(updateDelta.Nanoseconds())) // #nosec G404 (cryptographic strength irrelevant)
		// updatedDate := createdDate.Add(time.Nanosecond * time.Duration(randomUpdatedDelta))
		randomTelemDelta := rand.Int63n(int64(updateDelta.Nanoseconds())) // #nosec G404 (cryptographic strength irrelevant)
		telemDate := createdDate.Add(time.Nanosecond * time.Duration(randomTelemDelta))
		// createdDateStr := createdDate.UTC().Format("2006-01-02T15:04:05.000000Z0700")
		// updatedDateStr := updatedDate.UTC().Format("2006-01-02T15:04:05.000000Z0700")
		telemDateStr := telemDate.UTC().Format("2006-01-02T15:04:05.000000Z0700")
		s := Subscription{ids[i], "Subscription", "/api/accounts_mgmt/v1/subscriptions/1YuEObNEl4Z8b79mbbHD7a9hkl6", p, clusterIds[i], extIds[i], "1Yu9TWVAfvJu9Cj5hbMT6iYkdk8", telemDateStr, "--TODAY--", "--TODAY--", "None", extIds[i], c, false, "Active", "Telemetry", "2020-03-10T20:43:46.428922Z", "https://console-openshift-console.apps.jdgray-c6mvq.dev01.red-chesterfield.com", "0001-01-01T00:00:00Z", mets, "aws", "us-east-1", "0001-01-01T00:00:00Z"}
		test = append(test, s)
	}
	return test
}

// paginateSubscriptions splits out the subscriptions into SubscriptionLists of size n
func paginateSubscriptions(subs []Subscription, n int) []SubscriptionList {
	pages := len(subs)/n + 1
	sublists := make([]SubscriptionList, pages)

	for i := 0; i < pages; i++ {
		low := n * i
		high := low + n
		if high > len(subs) {
			high = len(subs)
		}
		items := subs[low:high]
		sl := SubscriptionList{
			Kind:  "SubscriptionList",
			Page:  i + 1,
			Size:  len(items),
			Total: len(subs),
			Items: items,
		}
		sublists[i] = sl
	}
	return sublists
}

// saveToFolder splits the list into files for each page inside the dir
func saveToFolder(sl []SubscriptionList, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		panic(err)
	}

	for i, page := range sl {
		filename := "subscription_response.json"
		if i > 0 {
			// Additional pages end in _<page_num>
			filename = fmt.Sprintf("subscription_response_%d.json", i+1)
		}
		b, _ := json.MarshalIndent(page, "", "    ")
		err := os.WriteFile(path.Join(dir, filename), b, 0600)
		if err != nil {
			panic(err)
		}
		fmt.Println("Wrote to file ", path.Join(dir, filename))
	}
}

func main() {
	outputDir := flag.String("output", "", "File location to save output. Example: './testserver/data/scenarios/onek_clusters'.")
	total := flag.Int("total", 1, "The total number of clusters will be created. Default is 1.")
	flag.Parse()

	subscriptions := generateSubscriptions(*total)
	pages := paginateSubscriptions(subscriptions, 1000)

	if *outputDir == "" {
		b, _ := json.MarshalIndent(pages, "", "  ")
		fmt.Printf("%s", b)
	} else {
		saveToFolder(pages, *outputDir)
	}
}

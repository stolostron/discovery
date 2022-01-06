// Copyright Contributors to the Open Cluster Management project

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	discovery "github.com/stolostron/discovery/api/v1alpha1"
)

const (
	DiscoveryConfigName = "discovery"
	SecretName          = "test-connection-secret"
)

var (
	scheme             = runtime.NewScheme()
	k8sClient          client.Client
	discoveryNamespace = "open-cluster-management"
	// Number of Discoveryconfigs to create
	configTotal = 10
	// Time to wait between applying Discoveryconfigs
	waitPeriod = 1 * time.Minute
	// Time to watch after applying Discoveryconfigs
	postApplyPeriod = 30 * time.Minute
	// Time to watch after deleting Discoveryconfigs
	postDeletePeriod = 10 * time.Minute
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(discovery.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

func main() {
	c, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	k8sClient, err = client.New(c, client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	// Cleanup namespaces at the end
	defer cleanup()

	done := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(1)

	// Continually log metrics throughout test
	go func() {
		startMetricsLogger(done)
		wg.Done()
	}()

	runScaleTest()

	// Stop metrics logger, wait for it to flush
	fmt.Println("Stopping logger")
	done <- true
	wg.Wait()
}

func runScaleTest() {
	// Create new discoveryconfigs at an interval
	for i := 0; i < configTotal; i++ {
		iter := fmt.Sprintf("scale%d", i)
		fmt.Printf("Creating config #%d\n", i)
		if err := createDiscoveryResources(iter); err != nil {
			log.Println("error creating discovery resources:", err)
		}
		time.Sleep(waitPeriod)
	}

	// Give time for re-reconciliation
	log.Println("Wait to reach steady state")
	time.Sleep(postApplyPeriod)

	deleteAllConfigs()
	time.Sleep(postDeletePeriod)
}

func createDiscoveryResources(ns string) error {
	if err := createNamespace(ns); err != nil {
		return err
	}
	if err := createSecret(ns); err != nil {
		return err
	}
	if err := createConfig(ns); err != nil {
		return err
	}
	return nil
}

func createNamespace(ns string) error {
	return k8sClient.Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	})
}

func createSecret(ns string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: ns,
		},
		StringData: map[string]string{
			"ocmAPIToken": "dummytoken",
		},
	}

	return k8sClient.Create(context.TODO(), secret)

}

func createConfig(ns string) error {
	config := &discovery.DiscoveryConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DiscoveryConfigName,
			Namespace: ns,
		},
		Spec: discovery.DiscoveryConfigSpec{
			Credential: SecretName,
			Filters:    discovery.Filter{LastActive: 7},
		},
	}

	baseURL := fmt.Sprintf("http://mock-ocm-server.%s.svc.cluster.local:3000", discoveryNamespace)
	scenario := "1000"
	config.SetAnnotations(map[string]string{"ocmBaseURL": baseURL + "/" + scenario, "authBaseURL": baseURL + "/" + scenario})
	return k8sClient.Create(context.TODO(), config)
}

func countConfigs() (int, error) {
	configs := &discovery.DiscoveryConfigList{}
	err := k8sClient.List(context.TODO(), configs)
	return len(configs.Items), err
}

func deleteAllConfigs() {
	for i := 0; i < configTotal; i++ {
		ns := fmt.Sprintf("scale%d", i)
		fmt.Println("Deleting config in namespace", ns)
		err := k8sClient.Delete(context.TODO(), &discovery.DiscoveryConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DiscoveryConfigName,
				Namespace: ns,
			},
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func cleanup() {
	for i := 0; i < configTotal; i++ {
		ns := fmt.Sprintf("scale%d", i)
		fmt.Println("Deleting namespace ", ns)
		err := k8sClient.Delete(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func discoveryPodName() (string, error) {
	pod := &corev1.PodList{}
	matchingLabels := client.MatchingLabels(map[string]string{"app": "discovery-operator"})
	if err := k8sClient.List(context.Background(), pod, matchingLabels); err != nil {
		return "", err
	}
	return pod.Items[0].Name, nil
}

func startMetricsLogger(done <-chan bool) {
	config := config.GetConfigOrDie()
	ns := "open-cluster-management"

	podName, err := discoveryPodName()
	if err != nil {
		log.Fatal(err)
	}

	mc, err := metrics.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("./test/scale/results/result.csv")
	if err != nil {
		log.Fatalln("error creating results file:", err)
	}

	writer := csv.NewWriter(file)
	err = writer.Write([]string{"Time elapsed", "Configs", "CPU usage", "Memory usage"})
	if err != nil {
		log.Println("error on write:", err)
	}

	now := time.Now()
	secondsPassed := func() int {
		return int(time.Since(now).Seconds())
	}

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-done:
			ticker.Stop()
			if err := file.Close(); err != nil {
				log.Println("error closing file:", err)
			}
			return
		case <-ticker.C:
			configs, cpu, mem := getMetricsFromMetricsAPI(mc, podName, ns)
			err := writer.Write([]string{strconv.Itoa(secondsPassed()), strconv.Itoa(configs), strconv.Itoa(cpu), strconv.Itoa(mem)})
			writer.Flush()
			if err != nil {
				log.Println("error writing record to csv:", err)
			}
		}
	}
}

func getMetricsFromMetricsAPI(clientset *metrics.Clientset, podName string, namespace string) (int, int, int) {
	podMetrics, err := clientset.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return 0, 0, 0
	}

	configs, err := countConfigs()
	if err != nil {
		log.Println(err)
		return 0, 0, 0
	}

	container := podMetrics.Containers[0]
	cpuQuantity := container.Usage.Cpu()
	memQuantity := container.Usage.Memory()

	msg := fmt.Sprintf("Configs: %d \t CPU usage: %vm \t Memory usage: %vMi", configs, cpuQuantity.MilliValue(), memQuantity.Value()/(1024*1024))
	fmt.Println(msg)

	return configs, int(cpuQuantity.MilliValue()), int(memQuantity.Value() / (1024 * 1024))
}

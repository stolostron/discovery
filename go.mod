module github.com/open-cluster-management/discovery

go 1.17

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.20.0
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/metrics v0.21.1
	sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/yaml v1.2.0
)

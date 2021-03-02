module github.com/open-cluster-management/discovery/testserver

go 1.15

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/open-cluster-management/discovery v0.0.0
	github.com/rs/zerolog v1.20.0
)

replace github.com/open-cluster-management/discovery => ../

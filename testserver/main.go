// Copyright Contributors to the Open Cluster Management project

package main

import (
	"flag"
	"os"

	api "github.com/open-cluster-management/discovery/testserver/api"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// func init() {
// 	flag.StringVar(&api.Scenario, "scenario", "tenClusters", "The address the metric endpoint binds to.")
// 	flag.Parse()
// }

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)

	r := gin.Default()
	flag.Parse()

	api.SetupEndpoints(r, log.Logger)

	err := r.Run(":3000")
	if err != nil {
		panic(err)
	}
}

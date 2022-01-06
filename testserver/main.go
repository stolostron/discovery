// Copyright Contributors to the Open Cluster Management project

package main

import (
	"os"

	api "github.com/stolostron/discovery/testserver/api"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

	api.SetupEndpoints(r, log.Logger)

	err := r.Run(":3000")
	if err != nil {
		panic(err)
	}
}

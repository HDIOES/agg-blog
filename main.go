package main

import (
	"log"
	"os"

	"github.com/gorilla/mux"

	"net/http"

	_ "github.com/lib/pq"

	"strconv"

	"github.com/HDIOES/agg-blog/di"
	"github.com/HDIOES/agg-blog/rest/util"
)

//CreateDI function to build new DI container
func main() {
	log.Println("Application has been runned")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configuration.json"
	}
	di := di.CreateDI(configPath, "migrations", false)
	di.Invoke(func(router *mux.Router, configuration *util.Configuration) {
		listenandserveErr := http.ListenAndServe(":"+strconv.Itoa(configuration.Port), router)
		if listenandserveErr != nil {
			panic(listenandserveErr)
		}
	})
}

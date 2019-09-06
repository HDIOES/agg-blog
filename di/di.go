package di

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/HDIOES/agg-blog/models"
	"github.com/HDIOES/agg-blog/rest"
	"github.com/HDIOES/agg-blog/rest/util"
	"github.com/gorilla/mux"
	"github.com/ory/dockertest"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/tkanos/gonfig"
	"go.uber.org/dig"
)

const dbType = "postgres"
const dbVersion = "11.4"
const dbPortMapping = "5432/tcp"

const dbUser = "test_forna_user"
const dbUserVar = "POSTGRES_USER=" + dbUser

const dbPassword = "12345"
const dbPasswordVar = "POSTGRES_PASSWORD=" + dbPassword

const dbName = "test_forna"
const dbNameVar = "POSTGRES_DB=" + dbName

const dbURLTemplate = dbType + "://" + dbUser + ":" + dbPassword + "@localhost:%s/" + dbName + "?sslmode=disable"

//CreateDI function to build di-container
func CreateDI(configPath, migrationPath string, test bool) *dig.Container {
	container := dig.New()
	container.Provide(func() *util.Configuration {
		log.Println("Loading configuration...")
		configuration := &util.Configuration{}
		gonfigErr := gonfig.GetConf(configPath, configuration)
		if gonfigErr != nil {
			panic(gonfigErr)
		}
		return configuration
	})
	container.Provide(func(configuration *util.Configuration) (sqlDb *sql.DB, dockerResource *dockertest.Resource, err error) {
		if test {
			pool, err := dockertest.NewPool("")
			if err != nil {
				return nil, nil, errors.Wrap(err, "")
			}
			resource, rErr := pool.Run(dbType, dbVersion, []string{
				dbUserVar,
				dbPasswordVar,
				dbNameVar})
			log.Print("Starting test container...")
			time.Sleep(10 * time.Second)
			if rErr != nil {
				defer resource.Close()
				return nil, nil, errors.Wrap(rErr, "")
			}
			configuration.DatabaseURL = fmt.Sprintf(dbURLTemplate, resource.GetPort(dbPortMapping))
			log.Println("PORT = " + resource.GetPort(dbPortMapping))
			dockerResource = resource
		}
		log.Println("Prepating db...")
		db, err := sql.Open("postgres", configuration.DatabaseURL)
		if err != nil {
			panic(err)
		}
		db.SetMaxIdleConns(configuration.MaxIdleConnections)
		db.SetMaxOpenConns(configuration.MaxOpenConnections)
		timeout := strconv.Itoa(configuration.ConnectionTimeout) + "s"
		timeoutDuration, durationErr := time.ParseDuration(timeout)
		if durationErr != nil {
			log.Println("Error parsing of timeout parameter")
			panic(durationErr)
		} else {
			db.SetConnMaxLifetime(timeoutDuration)
		}
		migrations := &migrate.FileMigrationSource{
			Dir: migrationPath,
		}
		if n, err := migrate.Exec(db, "postgres", migrations, migrate.Up); err == nil {
			log.Printf("Applied %d migrations!\n", n)
		} else {
			return nil, nil, errors.Wrap(err, "")
		}
		sqlDb = db
		return
	})
	container.Provide(func(db *sql.DB) *models.NewDAO {
		return &models.NewDAO{Db: db}
	})
	container.Provide(func(
		newDao *models.NewDAO,
		configuration *util.Configuration) (
		*rest.CreateNewHandler,
		*rest.FindNewHandler) {
		createNewHandler := &rest.CreateNewHandler{Dao: newDao}
		findNewHandler := &rest.FindNewHandler{Dao: newDao}
		return createNewHandler, findNewHandler
	})
	container.Provide(func(createNewHandler *rest.CreateNewHandler,
		findNewHandler *rest.FindNewHandler) *mux.Router {
		router := mux.NewRouter()
		router.Handle("/api/new", createNewHandler).
			Methods("POST")
		router.Handle("/api/new", findNewHandler).
			Methods("GET")
		router.Handle("/api/new", nil).
			Methods("DELETE")
		http.Handle("/", router)
		return router
	})
	return container
}

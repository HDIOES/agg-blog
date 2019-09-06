package test

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/HDIOES/agg-blog/models"
	"github.com/HDIOES/agg-blog/rest/util"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/gorilla/mux"

	"go.uber.org/dig"

	"github.com/HDIOES/agg-blog/di"
	"github.com/ory/dockertest"
	migrate "github.com/rubenv/sql-migrate"
)

var diContainer *dig.Container

func init() {
	diContainer = di.CreateDI("../configuration.json", "../migrations", true)
}

func TestMain(m *testing.M) {
	//prepare test database, test configuration and test router
	os.Exit(wrapperTestMain(m))
}

func wrapperTestMain(m *testing.M) int {
	defer diContainer.Invoke(func(db *sql.DB, testContainer *dockertest.Resource) {
		db.Close()
		testContainer.Close()
	})
	defer log.Print("Stopping test container")
	return m.Run()
}

func markAsFailAndAbortNow(t *testing.T, err error) {
	util.HandleError(err)
	t.FailNow()
}

func abortIfFail(t *testing.T, result bool) {
	if !result {
		t.FailNow()
	}
}

func EqualInt64Values(t *testing.T, expected, actual *int64, msgAndArgs ...interface{}) bool {
	if actual != nil && expected != nil {
		return assert.Equal(t, *expected, *actual, msgAndArgs)
	}
	return assert.True(t, expected == nil, msgAndArgs) && assert.True(t, actual == nil, msgAndArgs)
}

func EqualStringValues(t *testing.T, expected, actual *string, msgAndArgs ...interface{}) bool {
	if actual != nil && expected != nil {
		return assert.Equal(t, *expected, *actual, msgAndArgs)
	}
	return assert.True(t, expected == nil, msgAndArgs) && assert.True(t, actual == nil, msgAndArgs)
}

func EqualBoolValues(t *testing.T, expected, actual *bool, msgAndArgs ...interface{}) bool {
	if actual != nil && expected != nil {
		return assert.Equal(t, *expected, *actual, msgAndArgs)
	}
	return assert.True(t, expected == nil, msgAndArgs) && assert.True(t, actual == nil, msgAndArgs)
}

func EqualFloat64Values(t *testing.T, expected, actual *float64, msgAndArgs ...interface{}) bool {
	if actual != nil && expected != nil {
		return assert.Equal(t, *expected, *actual, msgAndArgs)
	}
	return assert.True(t, expected == nil, msgAndArgs) && assert.True(t, actual == nil, msgAndArgs)
}

func executeRequest(req *http.Request, router *mux.Router) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func insertNewToDatabase(newDao *models.NewDAO, name *string, body *string) (int64, error) {
	id, err := newDao.Create(models.NewDTO{
		Name: name,
		Body: body,
	})
	if err != nil {
		return 0, errors.Wrap(err, "")
	}
	return id, nil
}

func clearDb(newDao *models.NewDAO) error {
	if err1 := newDao.DeleteAll(); err1 != nil {
		return errors.Wrap(err1, "")
	}
	return nil
}

func applyMigrations(db *sql.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "../migrations",
	}
	if n, err := migrate.Exec(db, "postgres", migrations, migrate.Up); err == nil {
		log.Printf("Applied %d migrations!\n", n)
	} else {
		return errors.Wrap(err, "")
	}
	return nil
}

package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"bitbucket.org/danstutzman/nof1-backend/internal/model"
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Fixtures struct {
	dbConn *sql.DB
	router *mux.Router
	server *httptest.Server
}

func setupFixtures() *Fixtures {
	dbConn := db.PrepareFakeDb()
	model := model.NewModel(model.Config{
		AwsAccessKeyId:     "fakeaccesskeyid",
		AwsRegion:          "fakeregion",
		AwsS3Bucket:        "fakebucket",
		AwsSecretAccessKey: "fakesecret",
		DbConn:             dbConn,
		UploadDir:          "UPLOAD_DIR",
	})
	webapp := NewWebApp(model, dbConn, "STATIC_DIR", "fakepassword")
	router := NewRouter(webapp)
	server := httptest.NewServer(router)

	return &Fixtures{
		dbConn: dbConn,
		router: router,
		server: server,
	}
}

func teardownFixtures(fixtures *Fixtures) {
	fixtures.dbConn.Close()
	fixtures.server.Close()
}

func httpGet(t *testing.T, url string, expectedStatus int) string {
	resp, err := http.Get(url)
	assert.Nil(t, err)
	assert.Equal(t, expectedStatus, resp.StatusCode)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

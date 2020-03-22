package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/model"
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type Fixtures struct {
	dbConn *sql.DB
	router *mux.Router
	server *httptest.Server
}

func setupFixtures() *Fixtures {
	dbConn := db.PrepareFakeDb()
	model := model.NewModel(dbConn, "UPLOAD_DIR")
	webapp := NewWebApp(model, dbConn, "STATIC_DIR")
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

func TestRouterGet404(t *testing.T) {
	fixtures := setupFixtures()
	defer teardownFixtures(fixtures)

	httpGet(t, fixtures.server.URL+"/unknown", http.StatusNotFound)

	var nullString null.String
	requests := db.FromRequests(fixtures.dbConn, "")

	assert.Equal(t, 1, len(requests))
	requests[0].RemoteAddr = "NO_ASSERTION"
	requests[0].ReceivedAt = time.Time{}
	assert.Equal(t, requests, []db.RequestsRow{{
		Id:          1,
		ReceivedAt:  time.Time{},
		RemoteAddr:  "NO_ASSERTION",
		BrowserId:   null.Int{},
		HttpVersion: "HTTP/1.1",
		TlsProtocol: nullString,
		TlsCipher:   nullString,
		Method:      "GET",
		Path:        "/unknown",
		DurationMs:  0,
		StatusCode:  404,
		Size:        9,
		ErrorStack:  nullString,
	}})
}

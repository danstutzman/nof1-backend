package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"testing"
	"time"
)

func TestRouterGet404(t *testing.T) {
	fixtures := setupFixtures()
	defer teardownFixtures(fixtures)

	text := httpGet(t, fixtures.server.URL+"/unknown", http.StatusNotFound)
	assert.Equal(t, "Not Found\n", text)

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
		TlsProtocol: null.String{},
		TlsCipher:   null.String{},
		Method:      "GET",
		Path:        "/unknown",
		DurationMs:  0,
		StatusCode:  http.StatusNotFound,
		Size:        len([]byte(text)) - 1, // -1 for final newline
		ErrorStack:  null.String{},
	}})
}

package main

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"log"
	"os"

	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"

	"github.com/gin-gonic/gin"
	lp "github.com/septianw/jas-location/package"
	"github.com/stretchr/testify/assert"
)

type header map[string]string
type headers []header
type payload struct {
	Method string
	Url    string
	Body   io.Reader
}
type expectation struct {
	Code int
	Body string
}
type quest struct {
	pload  payload
	heads  headers
	expect expectation
}
type quests []quest

var LastPostID int64
var locin lp.LocationIn

func getArm() (*gin.Engine, *httptest.ResponseRecorder) {
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	Router(router)

	recorder := httptest.NewRecorder()
	return router, recorder
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func doTheTest(load payload, heads headers) *httptest.ResponseRecorder {
	var router, recorder = getArm()

	req, err := http.NewRequest(load.Method, load.Url, load.Body)
	handleErr(err)

	if len(heads) != 0 {
		for _, head := range heads {
			for key, value := range head {
				req.Header.Set(key, value)
			}
		}
	}
	router.ServeHTTP(recorder, req)

	return recorder
}

func SetupRouter() *gin.Engine {
	return gin.New()
}

func SetEnvironment() {
	var rt types.Runtime
	var Dbconf types.Dbconf

	Dbconf.Database = "ipoint"
	Dbconf.Host = "localhost"
	Dbconf.Pass = "dummypass"
	Dbconf.Port = 3306
	Dbconf.Type = "mysql"
	Dbconf.User = "asep"

	rt.Dbconf = Dbconf
	rt.Libloc = "/home/asep/gocode/src/github.com/septianw/jas/libs"

	common.WriteRuntime(rt)
}

func UnsetEnvironment() {
	os.Remove("/tmp/shinyRuntimeFile")
}

func TestLocationPostPositive(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()
	var locin lp.LocationIn

	locin.Name = "kedunggalar"
	locin.Latitude = -12.85993810
	locin.Longitude = 100.34589023

	locinJson, err := json.Marshal(locin)
	t.Logf("locinJson: %+v", string(locinJson))
	t.Logf("locinJson: %+v", locin)
	if err != nil {
		// t.Logf("Locations found: %+v", Locations)
		t.Logf("err FindLocation: %+v", err)
		t.Fail()
	}

	q := quest{
		pload:  payload{"POST", "/api/v1/location", strings.NewReader(string(locinJson))},
		heads:  headers{},
		expect: expectation{201, "contact post"},
	}

	rec := doTheTest(q.pload, q.heads)
	log.Println(rec.Body.String())

	// Locations, err := lp.FindLocation(locin)
	// if err != nil {
	// 	t.Logf("Locations found: %+v", Locations)
	// 	t.Logf("err FindLocation: %+v", err)
	// 	t.Fail()
	// }

	// locationsJson, err := json.Marshal(Locations[0])
	// if err != nil {
	// 	t.Logf("LocationJson: %+v", string(locationsJson))
	// 	t.Logf("err: %+v", err)
	// 	t.Fail()
	// }

	assert.Equal(t, q.expect.Code, rec.Code)
	// assert.Equal(t, string(locationsJson), strings.TrimSpace(rec.Body.String()))
	assert.Equal(t, "test", strings.TrimSpace(rec.Body.String()))
}

// FIXME
func TestLocationGetPositive(t *testing.T) {

}

// FIXME
func TestLocationPutPositive(t *testing.T) {

}

// FIXME
func TestLocationDeletePositive(t *testing.T) {

}

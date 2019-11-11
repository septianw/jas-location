package location

import (
	"log"
	"testing"

	"encoding/json"
	"os"

	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"
)

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

func TestInsertLocation(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()
	var locin LocationIn

	locin.Name = "karanggedang"
	locin.Latitude = -90.4455595
	locin.Longitude = 110.99288848
	Location, err := InsertLocation(locin)

	log.Println(Location, err)
}

func TestGetLocation(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	Locations, err := GetLocation(-1, 10, 0)
	t.Log(Locations, err)

	Locations, err = GetLocation(LastId, 10, 0)
	t.Log(Locations, err)
	LocationsJson, err := json.Marshal(Locations)
	t.Log(string(LocationsJson), err)
}

func TestUpdateLocation(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()
	var locin LocationIn

	Locations, err := GetLocation(LastId, 0, 0)
	t.Log(Locations, err)
	if err != nil {
		t.Fail()
	}
	if len(Locations) == 0 {
		t.Fail()
	}

	locin.Name = "karangturi"
	locin.Latitude = -90.44554454
	locin.Longitude = 110.99288456

	Location, err := UpdateLocation(LastId, locin)

	if (Location.Name == Locations[0].Name) &&
		(Location.Latitude == Locations[0].Latitude) &&
		(Location.Longitude == Locations[0].Longitude) {
		t.Log(Location)
		t.Log(err)
		t.Fail()
	}

	t.Log(locin)
	t.Log(Location, err)
}

func TestDeleteLocation(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	SetEnvironment()
	defer UnsetEnvironment()

	err := DeleteLocation(LastId)
	log.Println(err)
	if err != nil {
		t.Fail()
	}
}

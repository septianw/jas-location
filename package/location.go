package location

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/septianw/jas/common"
)

func getdbobj() (db *sql.DB, err error) {
	rt := common.ReadRuntime()
	dbs := common.LoadDatabase(filepath.Join(rt.Libloc, "database.so"), rt.Dbconf)
	db, err = dbs.OpenDb(rt.Dbconf)
	return
}

func Query(q string) (*sql.Rows, error) {
	db, err := getdbobj()
	common.ErrHandler(err)
	defer db.Close()

	return db.Query(q)
}

func Exec(q string) (sql.Result, error) {
	db, err := getdbobj()
	common.ErrHandler(err)
	defer db.Close()

	return db.Exec(q)
}

type LocationFull struct {
	Locid     int64
	Name      sql.NullString
	Latitude  sql.NullFloat64
	Longitude sql.NullFloat64
	Deleted   int8
}

type LocationOut struct {
	Locid     int64   `json:"locid" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type LocationIn struct {
	Name      string  `json:"name" binding:"required"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

/*
ERROR CODE LEGEND:
error containt 4 digits,
first digit represent error location either module or main app
1 for main app
2 for module

second digit represent error at level app or database
1 for app
2 for database

third digit represent error with input variable or variable manipulation
0 for skipping this error
1 for input validation error
2 for variable manipulation error

fourth digit represent error with logic, this type of error have
increasing error number based on which part of code that error.
0 for skipping this error
1 for unknown logical error
2 for whole operation fail, operation end unexpectedly
*/

const DATABASE_EXEC_FAIL = 2200
const MODULE_OPERATION_FAIL = 2102
const INPUT_VALIDATION_FAIL = 2110

var NOT_ACCEPTABLE = gin.H{"code": "NOT_ACCEPTABLE", "message": "You are trying to request something not acceptible here."}
var NOT_FOUND = gin.H{"code": "NOT_FOUND", "message": "You are find something we can't found it here."}
var LastId int64

/*
CREATE TABLE `location` (
  `locid` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `latitude` decimal(10,8) DEFAULT NULL,
  `longitude` decimal(11,8) DEFAULT NULL,
  `deleted` TINYINT(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`locid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
*/

func InsertLocation(locin LocationIn) (Location LocationOut, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var sbLoc strings.Builder
	var Locations []LocationOut

	sbLoc.WriteString(fmt.Sprintf(`Insert into location (name, latitude, longitude, deleted)
	values ('%s', %2.8f, %3.8f, 0)`, locin.Name, locin.Latitude, locin.Longitude))
	log.Println(sbLoc.String())

	result, err := Exec(sbLoc.String())
	log.Println(result, err)
	if err != nil {
		return
	}

	LastId, err = result.LastInsertId()
	if err != nil {
		return
	}

	Locations, err = GetLocation(LastId, 0, 0)
	if err != nil {
		return
	}
	if len(Locations) == 0 {
		return LocationOut{}, errors.New("Location not found.")
	}

	Location = Locations[0]

	return
}

func FindLocation(Location LocationIn) (Locations []LocationOut, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var sbLoc strings.Builder
	var lf LocationFull
	var locout LocationOut

	sbLoc.WriteString("SELECT locid, name, latitude, longitude FROM location WHERE deleted = 0")

	sbLoc.WriteString(fmt.Sprintf(" and name = '%s' and latitude = %2.8f and longitude = %3.8f",
		Location.Name, Location.Latitude, Location.Longitude))

	log.Println(sbLoc.String())
	rows, err := Query(sbLoc.String())
	if err != nil {
		return Locations, err
	}

	for rows.Next() {
		rows.Scan(&lf.Locid, &lf.Name, &lf.Latitude, &lf.Longitude)
		if lf.Name.Valid {
			locout.Name = lf.Name.String
		}
		if lf.Latitude.Valid {
			locout.Latitude = lf.Latitude.Float64
		}
		if lf.Longitude.Valid {
			locout.Longitude = lf.Longitude.Float64
		}
		locout.Locid = lf.Locid
		Locations = append(Locations, locout)
	}

	if len(Locations) == 0 {
		return Locations, errors.New("Location not found.")
	}

	return Locations, err
}

func GetLocation(id, limit, offset int64) (Locations []LocationOut, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var sbLoc strings.Builder
	var lf LocationFull
	var Location LocationOut

	sbLoc.WriteString(fmt.Sprintf("SELECT locid, name, latitude, longitude FROM location WHERE deleted = 0"))

	// ambil all
	if id == -1 {
		if limit == 0 {
			sbLoc.WriteString(fmt.Sprintf(" limit %d offset %d", 10, 0))
		} else {
			sbLoc.WriteString(fmt.Sprintf(" limit %d offset %d", limit, offset))
		}
	} else { // ambil id
		sbLoc.WriteString(fmt.Sprintf(" AND locid = %d", id))
	}
	log.Println(sbLoc.String())

	rows, err := Query(sbLoc.String())
	if err != nil {
		return Locations, err
	}

	for rows.Next() {
		rows.Scan(&lf.Locid, &lf.Name, &lf.Latitude, &lf.Longitude)
		Location.Locid = lf.Locid
		if lf.Name.Valid {
			Location.Name = lf.Name.String
		}
		if lf.Latitude.Valid {
			Location.Latitude = lf.Latitude.Float64
		}
		if lf.Longitude.Valid {
			Location.Longitude = lf.Longitude.Float64
		}
		Locations = append(Locations, Location)
	}

	if len(Locations) == 0 {
		return Locations, errors.New("Location not found.")
	}

	return Locations, err
}

func UpdateLocation(id int64, locin LocationIn) (Location LocationOut, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var sbLoc strings.Builder
	// var lf LocationFull
	var Locations []LocationOut

	Locations, err = GetLocation(id, 0, 0)
	if err != nil {
		return
	}
	if len(Locations) == 0 {
		return LocationOut{}, errors.New("Location not found.")
	}

	sbLoc.WriteString(fmt.Sprintf(`UPDATE location SET
	name = '%s', latitude = %2.8f, longitude = %3.8f WHERE locid = %d`,
		locin.Name, locin.Latitude, locin.Longitude, id))

	log.Println(sbLoc.String())
	result, err := Exec(sbLoc.String())
	if err != nil {
		return
	}
	log.Println(result.RowsAffected())

	Locations, err = GetLocation(id, 0, 0)
	if err != nil {
		return
	}

	Location = Locations[0]

	return
}

func DeleteLocation(id int64) (err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var sbLoc strings.Builder

	sbLoc.WriteString(fmt.Sprintf(`UPDATE location SET deleted = 1 WHERE locid = %d`, id))

	// Locations, err := GetLocation(id, 0, 0)
	// log.Println(Locations)
	// if err != nil {
	// 	return
	// }
	// if len(Locations) == 0 {
	// 	return errors.New("Location not found.")
	// }

	result, err := Exec(sbLoc.String())
	if err != nil {
		return
	}
	raff, err := result.RowsAffected()
	log.Println(raff)
	if err != nil {
		return
	}
	if raff == 0 {
		return errors.New("Location not found.")
	}

	lcs, err := GetLocation(id, 0, 0)
	if err.Error() == "Location not found." {
		return nil
	}
	if len(lcs) != 0 {
		err = errors.New("Record deletion fail.")
	}

	return
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	// "encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	loc "github.com/septianw/jas-location"
)

/*
  `uid` INT NOT NULL AUTO_INCREMENT,
  `uname` VARCHAR(225) NOT NULL,
  `upass` TEXT NOT NULL,
  `contact_contactid` INT NOT NULL,
*/

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

const VERSION = loc.Version

var NOT_ACCEPTABLE = gin.H{"code": "NOT_ACCEPTABLE", "message": "You are trying to request something not acceptible here."}
var NOT_FOUND = gin.H{"code": "NOT_FOUND", "message": "You are find something we can't found it here."}

var segments []string

func Bootstrap() {
	fmt.Println("Module location bootstrap.")
}

/*
POST   /user
GET    /user/(:uid)
GET    /user/all/(:offset)/(:limit)
-----
ini masuk ke terminal
GET    /user/login
	basic auth
	return token, refresh token
-----
PUT    /user/(:uid)
DELETE /user/(:uid)
*/

func Router(r *gin.Engine) {
	// db := common.LoadDatabase()
	r.Any("/api/v1/location/*path1", deflt)
	// r.GET("/user/list", func(c *gin.Context) {
	// 	c.String(http.StatusOK, "wow")
	// })
}

func deflt(c *gin.Context) {
	segments := strings.Split(c.Param("path1"), "/")
	// log.Printf("\n%+v\n", c.Request.Method)
	// log.Printf("\n%+v\n", c.Param("path1"))
	// log.Printf("\n%+v\n", segments)
	// log.Printf("\n%+v\n", len(segments))
	switch c.Request.Method {
	case "POST":
		if strings.Compare(segments[1], "") == 0 {
			PostLocationHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		}
		break
	case "GET":
		if strings.Compare(segments[1], "all") == 0 {
			GetLocationAllHandler(c)
		} else if i, e := strconv.Atoi(segments[1]); (e == nil) && (i > 0) {
			GetLocationHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, loc.NOT_ACCEPTABLE)
		}
		break
	case "PUT":
		if i, e := strconv.Atoi(segments[1]); (e == nil) && (i > 0) {
			PutLocationHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		}
		break
	case "DELETE":
		if i, e := strconv.Atoi(segments[1]); (e == nil) && (i > 0) {
			DeleteLocationHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		}
		break
	default:
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, loc.NOT_ACCEPTABLE)
		break
	}
	// c.String(http.StatusOK, "hai")
}

func dummyResponse(c *gin.Context) {
	c.String(http.StatusOK, "wow")
}

func PostLocationHandler(c *gin.Context) {
	var input loc.LocationIn

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": loc.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	Location, err := loc.InsertLocation(input)
	if err != nil {
		if strings.Compare("Contact not found.", err.Error()) == 0 {
			c.JSON(http.StatusNotFound, loc.NOT_FOUND)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": loc.DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	}

	c.JSON(http.StatusCreated, Location)
}

func GetLocationAllHandler(c *gin.Context) {
	var segments = strings.Split(c.Param("path1"), "/")
	var l, o int64
	var limit, offset int
	var err error

	if len(segments) == 3 {
		limit = 10
		offset, err = strconv.Atoi(segments[2])
	} else if len(segments) == 4 {
		limit, err = strconv.Atoi(segments[3])
		offset, err = strconv.Atoi(segments[2])
	} else {
		limit = 10
		offset = 0
	}

	if err == nil { // tidak ada error dari konversi
		l = int64(limit)
		o = int64(offset)
	}

	Locations, err := loc.GetLocation(-1, l, o)
	if err != nil {
		if strings.Compare("Contact not found.", err.Error()) == 0 {
			c.JSON(http.StatusNotFound, loc.NOT_FOUND)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": loc.DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	}

	c.JSON(http.StatusOK, Locations)
}

func GetLocationHandler(c *gin.Context) {
	var segments = strings.Split(c.Param("path1"), "/")
	var id int64 = 0

	i, e := strconv.Atoi(segments[1])

	if e == nil { // konversi berhasil
		id = int64(i)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": loc.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", e.Error())})
		return
	}

	Locations, err := loc.GetLocation(id, 0, 0)
	if err != nil {
		if strings.Compare("Contact not found.", err.Error()) == 0 {
			c.JSON(http.StatusNotFound, loc.NOT_FOUND)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": loc.DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	}

	c.JSON(http.StatusOK, Locations[0])
	return
}

func PutLocationHandler(c *gin.Context) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var input loc.LocationIn

	var segments = strings.Split(c.Param("path1"), "/")
	var id int64

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": loc.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	i, e := strconv.Atoi(segments[1])
	if e == nil { // konversi berhasil
		id = int64(i)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": loc.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", e.Error())})
		return
	}

	Location, err := loc.UpdateLocation(id, input)
	if err != nil {
		if strings.Compare("Contact not found.", err.Error()) == 0 {
			c.JSON(http.StatusNotFound, loc.NOT_FOUND)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": loc.DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	}
	c.JSON(http.StatusOK, Location)
	return
}

func DeleteLocationHandler(c *gin.Context) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var input loc.LocationIn

	var segments = strings.Split(c.Param("path1"), "/")
	var id int64

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": loc.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	i, e := strconv.Atoi(segments[1])
	if e == nil { // konversi berhasil
		id = int64(i)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": loc.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", e.Error())})
		return
	}

	err := loc.DeleteLocation(id)
	if err != nil {
		if strings.Compare("Contact not found.", err.Error()) == 0 {
			c.JSON(http.StatusNotFound, loc.NOT_FOUND)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": loc.DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	}

	// FIXME: Perbaiki ini, ini harusnya bukan no content tapi ok dan ada status di body.
	c.JSON(http.StatusNoContent, nil)
	return
}

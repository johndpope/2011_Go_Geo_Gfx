package dbutil

import (
	"fmt"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"reflect"
	"strings"
	"tshared/coreutil"
)

var (
	geoNamesDbName = "" 
)

var (
	DbSafe = &mgo.Safe {}
	IsHost = false
	Panic = false
)

func BsonMapToObject (bmap interface{}, ptr interface{}) {
	var objPtr = reflect.ValueOf(ptr)
	var objType = objPtr.Elem()
	var objField reflect.Value
	if objType.Kind() == reflect.Struct {
		if bsonM, isBsonM := bmap.(bson.M); isBsonM {
			for mapKey, mapVal := range bsonM {
				if objField = objType.FieldByName(coreutil.Ifstr(mapKey == "_id", "Id", strings.Title(mapKey))); objField.IsValid() && objField.CanSet() {
					objField.Set(reflect.ValueOf(mapVal))
				}
			}
		}
	}
}

func ConnectTo (url string) (*mgo.Session, error) {
	var conn, err = mgo.Dial(url)
	if conn != nil {
		if err == nil {
			conn.SetSafe(DbSafe)
		} else {
			conn.Close()
			conn = nil
			if Panic {
				panic(err)
			}
		}
	}
	return conn, err
}

func ConnectToGlobal () (*mgo.Session, error) {
	return ConnectTo(ConnectUrl("localhost", 9106))
}

func ConnectToLocal () (*mgo.Session, error) {
	return ConnectTo(ConnectUrl("localhost", coreutil.Ifi(IsHost, 4057, 5317)))
}

func ConnectUrl (host string, port int) string {
	return fmt.Sprintf("%s:%d?connect=direct", host, port)
}

func DropDatabase (dbConn *mgo.Session, dbName string) error {
	var err = dbConn.DB(dbName).DropDatabase()
	if Panic && (err != nil) {
		panic(err)
	}
	return err
}

func EnsureIndex (dbConn *mgo.Session, dbName string, collName string, index *mgo.Index) error {
	var err = dbConn.DB(dbName).C(collName).EnsureIndex(*index)
	if Panic && (err != nil) {
		panic(err)
	}
	return err
}

func Find (dbConn *mgo.Session, dbName string, collName string, query interface {}) *mgo.Query {
	return dbConn.DB(dbName).C(collName).Find(query)
}

func FindOne (dbConn *mgo.Session, dbName string, collName string, query interface{}, result interface{}) error {
	var err = Find(dbConn, dbName, collName, query).Limit(1).All(result)
	if Panic && (err != nil) {
		panic(err)
	}
	return err
}

func FindAll (dbConn *mgo.Session, dbName string, collName string, query interface {}, result interface{}) error {
	var err = Find(dbConn, dbName, collName, query).All(result)
	if Panic && (err != nil) {
		panic(err)
	}
	return err
}

func Insert (dbConn *mgo.Session, dbName string, collName string, recs ... interface{}) error {
	var err = dbConn.DB(dbName).C(collName).Insert(recs ...)
	if Panic && (err != nil) {
		panic(err)
	}
	return err
}

func GeoNamesDbName (dbConn *mgo.Session, force bool) string {
	if force || (len(geoNamesDbName) == 0) {
		var last = ""
		dbNames, err := dbConn.DatabaseNames()
		if err != nil {
			panic(err)
		}
		for _, dbn := range dbNames {
			if strings.HasPrefix(dbn, "gn_") {
				last = dbn
			}
		}
		geoNamesDbName = last
	}
	return geoNamesDbName
}

package main

// N27E010

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"math"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
	"tshared/dbutil"
	"tshared/coreutil"
	"tshared/fileutil"
	"tshared/geoutil"
	"tshared/gobutil"
	"tshared/netutil"
	"tshared/numutil"
	"tshared/stringutil"
)

type adminDiv struct {
	id int
	n string
	ac1 string
	ac2 string
	ac3 string
	ac4 string
}

type recMaker func (index int, rec []string) bson.M

var (
	flagFetch = flag.String("fetch", "", "_tmp/geonames/ Specify a temporary location to download and extract geonames.org .txt dumps to")
	flagMakeDb = flag.String("makedb", "", "_tmp/geonames/ Specify the location of geonames txt files to create and populate geonames DB collections from")
	flagMakeElev = flag.String("makeelev", "", "/ssd1/m/e/ Specify the target location to create elevation PNG tiles in")
	flagMakeGobs = flag.String("makegobs", "", "/ssd1/m/ Specify the location of geonames gob files to create and populate from geonames DB collections")
	flagTest = flag.String("test", "", "anyfoo Some testing")
	adminNil = &adminDiv { -1, "", "", "", "", "" }
	channel = make(chan string)
	mrs = make([]interface{}, 9000000)
)

func createDb (sourceDir string) {
	var err error
	var dbConn *mgo.Session
	var dbIndex, lastIndex = 0, 0
	var dbName, tmp string
	var dbNames []string
	var admins []*adminDiv
	var features = []string {}
	var timezones = []string {}
	var countries = []string {}
	fmt.Println("Connecting...")
	dbutil.Panic = true
	dbConn, err = dbutil.ConnectToGlobal()
	dbNames, err = dbConn.DatabaseNames()
	if err != nil {
		panic(err)
	}
	for _, dbName = range dbNames {
		if strings.HasPrefix(dbName, "gn_") {
			dbIndex++
		}
	}
	for {
		dbName = fmt.Sprintf("gn_%d", dbIndex)
		if stringutil.InSliceAt(dbNames, dbName) < 0 {
			break
		} else {
			dbIndex++
		}
	}
	createDbCollection(dbConn, dbName, path.Join(sourceDir, "timeZones.txt"), "t", true, false, func (index int, rec []string) bson.M {
		var n = strings.Replace(rec[0], "_", " ", -1)
		timezones = append(timezones, n)
		return bson.M { "_id": index - 1, "n": n, "g": stringutil.ToFloat32(rec[1]), "d": stringutil.ToFloat32(rec[2]), "r": stringutil.ToFloat32(rec[3]) }
	})
	createDbCollection(dbConn, dbName, path.Join(sourceDir, "featureCodes_en.txt"), "f", false, false, func (index int, rec []string) bson.M {
		features = append(features, rec[0])
		return bson.M { "_id": index, "n": rec[0], "t": rec[1], "d": rec[2] }
	})
	createDbCollection(dbConn, dbName, path.Join(sourceDir, "countryInfo.txt"), "c", false, false, func (index int, rec []string) bson.M {
		/*
#ISO	ISO3	ISO-Numeric	fips	Country					Capital				Area(in sq km)	Population	Continent	tld	CurrencyCode	CurrencyName	Phone	Postal Code Format	Postal Code Regex	Languages			geonameid	neighbours	EquivalentFipsCode
AD		AND		020			AN		Andorra					Andorra la Vella	468				84000		EU			.ad	EUR				Euro			376		AD###				^(?:AD)*(\d{3})$	ca					3041565		ES,FR	
AE		ARE		784			AE		United Arab Emirates	Abu Dhabi			82880			4975593		AS			.ae	AED				Dirham			971												ar-AE,fa,en,hi,ur	290557		SA,OM	
		*/
		countries = append(countries, rec[0])
		return bson.M { "_id": index, "i": rec[0], "i3": rec[1], "f": rec[2], "t": rec[4], "ca": rec[5], "co": rec[8], "d": rec[9], "cc": rec[10], "cn": rec[11], "p": rec[12], "l": stringutil.Split(rec[15], ","), "g": stringutil.ToInt(rec[16]), "n": stringutil.Split(rec[17], ",") }
	})
	createDbCollection(dbConn, dbName, path.Join(sourceDir, "zip_allCountries.txt"), "z", false, true, func (index int, rec []string) bson.M {
		var an = []string { stringutil.Title(rec[3]), stringutil.Title(rec[5]), stringutil.Title(rec[7]) }
		var ac = []string { rec[4], rec[6], rec[8] }
		var ll, err = numutil.NewDvec2(rec[10], rec[9])
		var n = rec[2]
		var r = bson.M { "_id": index, "c": stringutil.InSliceAt(countries, rec[0]), "z": rec[1], "n": n }
		var words []string
		if (err == nil) && (ll.X >= geoutil.LonMin) && (ll.X <= geoutil.LonMax) && (ll.Y >= geoutil.LatMin) && (ll.Y <= geoutil.LatMax) {
			r["l"] = ll
		} else {
			return nil
		}
		r["a"], admins = findAdminIndex(admins, rec[0], ac[0], ac[1], ac[2], an)
		if len(n) > 0 {
			if n == strings.ToUpper(n) {
				n = stringutil.Title(n)
			}
			if words = stringutil.Split(n, " "); len(words) > 1 {
				n = ""
				for i, w := range words {
					if stringutil.InSliceAt(words, w) == i {
						n = stringutil.Concat(n, w, " ")
					}
				}
				r["n"] = n[0:len(n) - 1]
			} else {
				r["n"] = n
			}
		}
		return r
	})
	for i, ad := range admins {
		dbutil.Insert(dbConn, dbName, "a", fixupAdminDiv(bson.M { "_id": i, "a": stringutil.NonEmpties(true, ad.ac1, ad.ac2, ad.ac3, ad.ac4), "n": ad.n }, countries))
	}
	stringutil.ForEach(func(i int, s string) {
		createDbCollection(dbConn, dbName, path.Join(sourceDir, s), "a", false, false, func (index int, rec []string) bson.M {
			var a = stringutil.Split(rec[0], ".")
			var n, na = rec[1], rec[2]
			var r bson.M
			for _, ad := range admins {
				if (ad.ac1 == a[0]) && (ad.ac2 == a[1]) && ((len(a) == 2) || (ad.ac3 == a[2])) {
					return nil
				}
			}
			index += len(admins)
			if i == 0 {
				lastIndex = index
			} else {
				index = index + 1 + lastIndex
			}
			r = bson.M { "_id": index, "a": a, "g": stringutil.ToInt(rec[3]), "n": n, "na": na }
			return fixupAdminDiv(r, countries)
		})
	}, "admin1CodesASCII.txt", "admin2Codes.txt")
	lastIndex = 0
	stringutil.ForEach(func (i int, s string) {
		createDbCollection(dbConn, dbName, path.Join(sourceDir, s), "n", false, true, func(index int, rec [] string) bson.M {
			/*
				0	geonameid         : integer id of record in geonames database
				1	name              : name of geographical point (utf8) varchar(200)
				2	asciiname         : name of geographical point in plain ascii characters, varchar(200)
				3	alternatenames    : alternatenames, comma separated varchar(5000)
				4	latitude          : latitude in decimal degrees (wgs84)
				5	longitude         : longitude in decimal degrees (wgs84)
				6	feature class     : see http://www.geonames.org/export/codes.html, char(1)
				7	feature code      : see http://www.geonames.org/export/codes.html, varchar(10)
				8	country code      : ISO-3166 2-letter country code, 2 characters
				9	cc2               : alternate country codes, comma separated, ISO-3166 2-letter country code, 60 characters
				10	admin1 code       : fipscode (subject to change to iso code), see exceptions below, see file admin1Codes.txt for display names of this code; varchar(20)
				11	admin2 code       : code for the second administrative division, a county in the US, see file admin2Codes.txt; varchar(80) 
				12	admin3 code       : code for third level administrative division, varchar(20)
				13	admin4 code       : code for fourth level administrative division, varchar(20)
				17	timezone          : the timezone id (see file timeZone.txt)
			*/
			var pos, c = -1, -1
			var ll, err = numutil.NewDvec2(rec[5], rec[4])
			var n, na = rec[1], rec[2]
			var an = stringutil.Without(stringutil.Split(rec[3], ","), false, n, na, "")
			var tz = strings.Replace(rec[17], "_", " ", -1)
			var r = bson.M { "g": stringutil.ToInt(rec[0]) }
			if (err == nil) && (ll.X >= geoutil.LonMin) && (ll.X <= geoutil.LonMax) && (ll.Y >= geoutil.LatMin) && (ll.Y <= geoutil.LatMax) {
				r["l"] = ll
			} else {
				return nil
			}
			if i == 0 {
				lastIndex = index
			} else {
				index = index + 1 + lastIndex
			}
			r["_id"] = index
			if len(n) == 0 {
				n = na
			}
			if (len(n) == 0) && (len(an) > 0) {
				n = an[0]
			}
			if (len(na) > 0) && (strings.ToLower(na) == strings.ToLower(n)) {
				na = ""
			}
			an = stringutil.Without(an, false, n, na)
			if len(n) > 0 {
				r["n"] = n
			}
			if len(na) > 0 {
				r["na"] = na
			}
			if len(an) > 0 {
				r["an"] = an
			}
			if pos = stringutil.InSliceAt(timezones, tz); pos >= 0 {
				r["t"] = pos
			}
			if pos = stringutil.InSliceAt(countries, rec[8]); pos >= 0 {
				c = pos
				r["c"] = pos
			} else if (len(rec[9]) > 0) {
				for _, cn := range stringutil.Split(rec[9], ",") {
					if pos = stringutil.InSliceAt(countries, cn); pos >= 0 {
						c = pos
						r["c"] = pos
						break
					}
				}
			}
			if (c >= 0) {
				pos, _ = findAdminIndex(admins, countries[c], rec[10], rec[11], rec[12], nil)
				if pos >= 0 {
					r["a"] = pos
				}
			}
			if pos = stringutil.InSliceAt(features, stringutil.Concat(rec[6], ".", rec[7])); pos >= 0 {
				r["f"] = pos
			}
			return r
		})
	}, "allCountries.txt", "null.txt")
	dbNames, err = dbConn.DatabaseNames()
	if err == nil {
		dbutil.Panic = false
		for _, tmp = range dbNames {
			if (tmp != dbName) && strings.HasPrefix(tmp, "gn_") {
				dbutil.DropDatabase(dbConn, tmp)
			}
		}
	}
}

func createDbCollection (dbConn *mgo.Session, dbName string, sourceFilePath, collName string, skipFirst bool, hasGeoIndex bool, makeRec recMaker) {
	var fr *os.File
	var br *bufio.Reader
	var err error
	var rec []string
	var line []byte
	var isPrefix, isDvec2 bool
	var geoIndexDone = false
	var mr bson.M
	var ll numutil.Dvec2
	var i, ri, rc = 0, 0, 0
	fr, err = os.Open(sourceFilePath)
	if fr != nil {
		defer fr.Close()
	}
	if err != nil {
		panic(err)
	}
	if fr != nil {
		fmt.Println("Reading", sourceFilePath)
		br = bufio.NewReaderSize(fr, coreutil.Ifi(strings.HasSuffix(sourceFilePath, "allCountries.txt"), 1024 * 1024 * 1024, 1024 * 1024 * 72))
		for rec = nil; err == nil; line, isPrefix, err = br.ReadLine() {
			if isPrefix || err != nil {
				fmt.Println(err)
				panic("readline")
			} else {
				rec = stringutil.Split(string(line), "\t")
			}
			if (rec != nil) && (len(rec) > 0) && !strings.HasPrefix(rec[0], "#") {
				if ((i != 0) || !skipFirst) {
					if mr = makeRec(i, rec); mr != nil {
						mrs[ri] = mr
						ri++
						rc++
						if hasGeoIndex {
							if ll, isDvec2 = mr["l"].(numutil.Dvec2); isDvec2 {
								mr["l"] = []float64 { ll.X, ll.Y }
								if !geoIndexDone {
									dbutil.EnsureIndex(dbConn, dbName, collName, &mgo.Index { Key: []string { "@l" }, Bits: 32, Min: -180, Max: 181 })
									geoIndexDone = true
								}
							}
						}
						if (i % 250000) == 0 {
							fmt.Println("Read", i)
						}
					}
				}
				i++
			}
		}
		for i = 0; i < rc; i = i + 4096 {
			if ri = i + 4096; ri > rc {
				ri = rc
			}
			dbutil.Insert(dbConn, dbName, collName, mrs[i:ri] ...)
			if (i % (4096 * 50)) == 0 {
				fmt.Println("Insert", i)
			}
		}
		fmt.Println("Recs total: ", collName, rc)
	}
}

func createElevs (targetDir string) {
	type sourceRec struct {
		dirPath string
		filePath string
		fileTime time.Time
		byteOrder binary.ByteOrder
		fileLen int64
	}
	var srcBasePath = "/media/hdx/elev_raw/"
	var force = []string { /*"N64W024"*/ }
	var initSourceRec = func (dirRelPath string) *sourceRec {
		return &sourceRec { path.Join(srcBasePath, dirRelPath), "", time.Time {}, binary.LittleEndian, 0 }
	}
	var fexts = []string { ".hgt", ".hgt.le", ".HGT", ".HGT.LE" }
	var srcRecs = map[string] *sourceRec { "": initSourceRec(""), "bathy": initSourceRec("bathy/hgt"), "srtm": initSourceRec("srtm/hgt"), "rmw": initSourceRec("rmw3"), "vfp": initSourceRec("vfp") }
	var lola numutil.Dvec2
	var fp, fbName, skey, nkey string
	var hasBathy, hasElev bool
	var srec *sourceRec
	var newest, srtmTime time.Time
	var pngMakers = 0
	var pngChan = make(chan bool)
	var makePng = func (outFilePath string, elevFile *sourceRec, bathyFilePath string) {
		var ps, ps1, psx = 1200, 1201, 1200 * 2400
		var bathyLandVal int16 = 0
		var bathyMapping = []int16 { -4500, -2500, -9500, -6500, -10500, -3500, bathyLandVal, -5500, -7500, -200, -8500, -1500, -750 }
		var pngSize = image.Rect(0, 0, ps, ps)
		var px, py int
		var bv, ev, fv int16
		var srcFile, bathFile, pngFile *os.File
		var err error
		var pngImage *image.NRGBA
		fmt.Println(outFilePath)
		if pngFile, err = os.Create(outFilePath); err != nil {
			fmt.Println("PANIC", err)
			panic(err)
		}
		if bathFile, err = os.Open(bathyFilePath); err != nil {
			bathFile = nil
		}
		if srcFile, err = os.Open(elevFile.filePath); err != nil {
			srcFile = nil
		}
		if (bathFile != nil) || (srcFile != nil) {
			pngImage = image.NewNRGBA(pngSize)
			for py = 0; py < pngSize.Max.Y; py++ {
				for px = 0; px < pngSize.Max.X; px++ {
					if (bathFile == nil) || !fileutil.ReadFromBinary(bathFile, int64((math.Floor(float64(py) / 40) * 30) + math.Floor(float64(px) / 40)) * 2, binary.LittleEndian, &bv) {
						bv = bathyLandVal
					} else {
						bv = bathyMapping[bv]
					}
					if (srcFile == nil) || !fileutil.ReadFromBinary(srcFile, (int64((py * coreutil.Ifi(int(elevFile.fileLen) == psx, ps, ps1)) + px) * 2) /*+ coreutil.Ifl(int(elevFile.fileLen) == psx, 0, 2402)*/, elevFile.byteOrder, &ev) {
						ev = -32768
					}
					fv = coreutil.Ifs(bv == bathyLandVal, ev, coreutil.Ifs((ev != 0) && (ev > bv), ev, bv))
					pngImage.Set(px, py, color.NRGBA { 0, byte(fv), byte(fv >> 8), 255 })
				}
			}
			if bathFile != nil {
				bathFile.Close()
			}
			if srcFile != nil {
				srcFile.Close()
			}
			err = png.Encode(pngFile, pngImage)
			if err != nil {
				fmt.Println("PANIC", err)
				panic(err)
			}
		}
		pngFile.Close()
		pngChan <- true
	}
	var goMakePng = func (outFilePath string, elevFile *sourceRec, bathyFilePath string) {
		for pngMakers >= 8 {
			<- pngChan
			pngMakers--
		}
		pngMakers++
		go makePng(outFilePath, elevFile, bathyFilePath)
	}
	srtmTime = time.Date(2008, 10, 15, 12, 30, 30, 500, time.Local)
	for lola.Y = geoutil.LatMin; lola.Y < geoutil.LatMax; lola.Y++ {
		for lola.X = geoutil.LonMin; lola.X < geoutil.LonMax; lola.X++ {
			fbName = geoutil.LoLaFileName(lola.X, lola.Y)
			if (len(force) == 0) || (stringutil.InSliceAt(force, fbName) >= 0) {
				fp = path.Join(targetDir, fbName)
				for skey, srec = range srcRecs {
					if len(skey) > 0 {
						srec.filePath, srec.fileTime, srec.fileLen = fileutil.FileExistsPath(srec.dirPath, fbName, fexts, true, false)
						if skey == "srtm" {
							srec.fileTime = srtmTime
						}
					}
				}
				newest = time.Time {}
				nkey = ""
				for skey, srec = range srcRecs {
					if (len(skey) > 0) && (skey != "bathy") && (len(srec.filePath) > 0) && (srec.fileTime.After(newest)) {
						newest = srec.fileTime
						nkey = skey
						if strings.HasSuffix(strings.ToLower(srec.filePath), ".hgt.le") {
							srec.byteOrder = binary.LittleEndian
						} else {
							srec.byteOrder = binary.BigEndian
						}
					}
				}
				if hasBathy, hasElev = len(srcRecs["bathy"].filePath) > 0, len(nkey) > 0; hasBathy || hasElev {
					goMakePng(fp, srcRecs[nkey], srcRecs["bathy"].filePath)
				} else {
					fmt.Println("\tNO BATHY, NO ELEV!\n")
				}
				if (len(srcRecs["rmw"].filePath) > 0) && (len(srcRecs["srtm"].filePath) > 0) && (len(srcRecs["vfp"].filePath) > 0) {
					goMakePng(path.Join(targetDir + "_rsv", fbName + "." + nkey + ".r.png"), srcRecs["rmw"], srcRecs["bathy"].filePath)
					goMakePng(path.Join(targetDir + "_rsv", fbName + "." + nkey + ".s.png"), srcRecs["srtm"], srcRecs["bathy"].filePath)
					goMakePng(path.Join(targetDir + "_rsv", fbName + "." + nkey + ".v.png"), srcRecs["vfp"], srcRecs["bathy"].filePath)
				} else if (len(srcRecs["rmw"].filePath) > 0) && (len(srcRecs["vfp"].filePath) > 0) {
					goMakePng(path.Join(targetDir + "_rv", fbName + "." + nkey + ".r.png"), srcRecs["rmw"], srcRecs["bathy"].filePath)
					goMakePng(path.Join(targetDir + "_rv", fbName + "." + nkey + ".v.png"), srcRecs["vfp"], srcRecs["bathy"].filePath)
				} else if (len(srcRecs["rmw"].filePath) > 0) && (len(srcRecs["srtm"].filePath) > 0) {
					goMakePng(path.Join(targetDir + "_rs", fbName + "." + nkey + ".r.png"), srcRecs["rmw"], srcRecs["bathy"].filePath)
					goMakePng(path.Join(targetDir + "_rs", fbName + "." + nkey + ".s.png"), srcRecs["srtm"], srcRecs["bathy"].filePath)
				} else if (len(srcRecs["vfp"].filePath) > 0) && (len(srcRecs["srtm"].filePath) > 0) {
					goMakePng(path.Join(targetDir + "_sv", fbName + "." + nkey + ".v.png"), srcRecs["vfp"], srcRecs["bathy"].filePath)
					goMakePng(path.Join(targetDir + "_sv", fbName + "." + nkey + ".s.png"), srcRecs["srtm"], srcRecs["bathy"].filePath)
				}
			}
		}
	}
}

func createGobs (targetDir string) {
	var makeGob = func (bmap interface{}, ptr interface{}) interface{} {
		dbutil.BsonMapToObject(bmap, ptr)
		return ptr
	}
	var makeAdminGob = func (bmap interface{}) interface{} {
		return makeGob(bmap, &geoutil.GeoNamesAdminRecord{})
	}
	var makeNameGob = func (bmap interface{}) interface{} {
		return makeGob(bmap, &geoutil.GeoNamesNameRecord{})
	}
	var makeZipGob = func (bmap interface{}) interface{} {
		return makeGob(bmap, &geoutil.GeoNamesZipRecord{})
	}
	var lola numutil.Dvec2
	var box [2][2]float64
	var dbConn *mgo.Session
	var dbName, fbName, fp string
	var geoRecs []interface{}
	dbutil.Panic = true
	dbConn, _ = dbutil.ConnectToGlobal()
	defer dbConn.Close()
	dbName = dbutil.GeoNamesDbName(dbConn, true)
	dbutil.FindAll(dbConn, dbName, "a", nil, &geoRecs)
	gobutil.CreateGobsFile(path.Join(targetDir, "ga"), &geoRecs, makeAdminGob, true)
	for lola.Y = geoutil.LatMin; lola.Y < geoutil.LatMax; lola.Y++ {
		for lola.X = geoutil.LonMin; lola.X < geoutil.LonMax; lola.X++ {
			fbName = geoutil.LoLaFileName(lola.X, lola.Y)
			fmt.Println(fbName)
			geoRecs = nil
			box[0][0] = lola.X
			box[0][1] = lola.Y
			box[1][0] = lola.X + 1
			box[1][1] = lola.Y + 1
			dbutil.FindAll(dbConn, dbName, "n", bson.M { "l": bson.M { "$within":  bson.M { "$box": box } } }, &geoRecs)
			if len(geoRecs) == 0 {
				geoRecs = nil
				dbutil.FindOne(dbConn, dbName, "n", bson.M { "l": bson.M { "$near": []float64 { lola.X + 0.5, lola.Y + 0.5 } } }, &geoRecs)
			}
			fp = path.Join(path.Join(targetDir, "gn"), fbName)
			fmt.Println(fp)
			gobutil.CreateGobsFile(fp, &geoRecs, makeNameGob, true)
			geoRecs = nil
			dbutil.FindAll(dbConn, dbName, "z", bson.M { "l": bson.M { "$within":  bson.M { "$box": box } } }, &geoRecs)
			if len(geoRecs) > 0 {
				fp = path.Join(path.Join(targetDir, "gz"), fbName)
				fmt.Println(fp)
				gobutil.CreateGobsFile(fp, &geoRecs, makeZipGob, true)
			}
		}
	}
}

func fetch (targetDir string) {
	var (
		fileCount = 0
		fileDone string
		geoFiles = map[string] string {
			"featureCodes_en.txt": "dump/featureCodes_en.txt",
			"null.txt": "dump/no-country.zip",
			"timeZones.txt": "dump/timeZones.txt",
			"countryInfo.txt": "dump/countryInfo.txt",
			"admin1CodesASCII.txt": "dump/admin1CodesASCII.txt",
			"admin2Codes.txt": "dump/admin2Codes.txt",
			// "hierarchy.txt": "dump/hierarchy.zip",
			"zip_allCountries.txt": "zip/allCountries.zip",
			"allCountries.txt": "dump/allCountries.zip",
		}
	)
	for fileName, relUrl := range geoFiles {
		go fetchFile(targetDir, fileName, relUrl)
	}
	for fileCount < len(geoFiles) {
		fileDone = <- channel
		fileCount = fileCount + 1
		fmt.Printf("DONE: %s --- %d remaining...\n", fileDone, len(geoFiles) - fileCount)
	}
}

func fetchFile (targetDir, fileName, relUrl string) {
	var fullUrl = "http://download.geonames.org/export/" + relUrl
	var filePath = path.Join(targetDir, strings.Replace(relUrl, "/", "_", -1))
	var err = netutil.DownloadFile(fullUrl, filePath)
	if err == nil {
		if strings.HasSuffix(filePath, ".zip") {
			fmt.Printf("UNZIP: %s from %s\n", fileName, filePath)
			err = fileutil.ExtractZipFile(filePath, targetDir, true, "zip_", fileName)
		} else {
			err = os.Rename(filePath, path.Join(targetDir, fileName))
		}
	}
	if err != nil {
		panic(err)
	}
	channel <- fileName
}

func findAdminIndex (admins []*adminDiv, ac1, ac2, ac3, ac4 string, an []string) (int, []*adminDiv) {
	var adminIndex = -1
	var aname string
	for i, ad := range admins {
		if (ad.ac1 == ac1) && (ad.ac2 == ac2) && (ad.ac3 == ac3) && (ad.ac4 == ac4) {
			adminIndex = i
			break
		}
	}
	if (adminIndex == -1) && (len(an) > 0) {
		if aname = stringutil.FirstNonEmpty(-1, an...); len(aname) > 0 {
			adminIndex = len(admins)
			admins = append(admins, &adminDiv { len(admins), aname, ac1, ac2, ac3, ac4 })
		}
	}
	return adminIndex, admins
}

func fixupAdminDiv (r bson.M, countries []string) bson.M {
	var n, na = stringutil.ToString(r["n"], ""), stringutil.ToString(r["na"], "")
	var a = stringutil.ToStrings(r["a"])
	var pos int
	delete(r, "n")
	delete(r, "na")
	if (a != nil) && (len(a) > 0) {
		if pos = stringutil.InSliceAt(countries, a[0]); pos >= 0 {
			r["c"] = pos
		}
		a = a[1:]
	}
	if (a != nil) && (len(a) > 0) {
		r["a"] = a
	} else {
		delete(r, "a")
	}
	if (len(n) == 0) && (len(na) > 0) {
		n = na
		na = ""
	}
	if n == na {
		na = ""
	}
	if len(na) == 0 {
		pos = strings.Index(n, " / ")
		if (pos >= 0) && (pos == strings.LastIndex(n, " / ")) && stringutil.IsAscii(n[pos + 3:]) && !stringutil.IsAscii(n[:pos]) {
			na = n[pos + 3:]
			n = n[:pos]
		}
	}
	if len(n) > 0 {
		r["n"] = n
	}
	if len(na) > 0 {
		r["na"] = na
	}
	return r
}

func main () {
	runtime.GOMAXPROCS(8)
	flag.Parse()
	if *flagFetch != "" {
		fetch(*flagFetch)
	}
	if *flagMakeDb != "" {
		createDb(*flagMakeDb)
	}
	if *flagMakeElev != "" {
		createElevs(*flagMakeElev)
	}
	if *flagMakeGobs != "" {
		createGobs(*flagMakeGobs)
	}
	if *flagTest != "" {
		type tmpType struct {
			name string
			lo float64
			la float64
			hasOld bool
			hasNew bool
			hasBathy bool
			hasSrtm bool
			hasRmw bool
			hasVfp bool
		}
		var nold, nnew, nbathy, nsrtm, nrmw, nvfp, tc = 0, 0, 0, 0, 0, 0, 0
		var tmps []tmpType
		var tmp tmpType
		var missings []string
		for lat := geoutil.LatMin; lat < geoutil.LatMax; lat++ {
			for lon := geoutil.LonMin; lon < geoutil.LonMax; lon++ {
				tmp.lo = lon
				tmp.la = lat
				tmp.name = geoutil.LoLaFileName(lon, lat)
				if tmp.hasOld = fileutil.FileExists("/ssd1/m/oe/" + tmp.name + ".png"); tmp.hasOld {
					nold++
				}
				if tmp.hasNew = fileutil.FileExists("/ssd1/m/e/" + tmp.name); tmp.hasNew {
					nnew++
				}
				if tmp.hasOld && !tmp.hasNew {
					missings = append(missings, tmp.name)
				}
				if tmp.hasBathy = fileutil.FileExists("/media/hdx/elev_raw/bathy/hgt/" + tmp.name + ".hgt.le"); tmp.hasBathy {
					nbathy++
				}
				if tmp.hasSrtm = fileutil.FileExists("/media/hdx/elev_raw/srtm/hgt/" + tmp.name + ".hgt.le"); tmp.hasSrtm {
					nsrtm++
				}
				if tmp.hasRmw = fileutil.FileExists("/media/hdx/elev_raw/rmw3/" + tmp.name + ".HGT"); tmp.hasRmw {
					nrmw++
				}
				if tmp.hasVfp = fileutil.FileExists("/media/hdx/elev_raw/vfp/" + tmp.name + ".hgt"); tmp.hasVfp {
					nvfp++
				}
				tmps = append(tmps, tmp)
			}
		}
		fmt.Printf("Old: %d\tNew: %d\tBathy: %d\tSrtm: %d\tRmw: %d\tVfp: %d\t\n", nold, nnew, nbathy, nsrtm, nrmw, nvfp)
		fmt.Println(missings)
		type chk func (t tmpType) bool
		var checks = map[string]chk {
			"srtm only": func (t tmpType) bool { return t.hasSrtm && !(t.hasRmw || t.hasVfp) },
			"rmw only": func (t tmpType) bool { return t.hasRmw && !(t.hasSrtm || t.hasVfp) },
			"vfp only": func (t tmpType) bool { return t.hasVfp && !(t.hasSrtm || t.hasRmw) },
			"s+r only": func (t tmpType) bool { return t.hasRmw && t.hasSrtm && !t.hasVfp },
			"s+v only": func (t tmpType) bool { return t.hasVfp && t.hasSrtm && !t.hasRmw },
			"r+v only": func (t tmpType) bool { return t.hasVfp && t.hasRmw && !t.hasSrtm },
			"srv": func (t tmpType) bool { return t.hasVfp && t.hasRmw && t.hasSrtm },
		}
		for cn, cf := range checks {
			tc = 0
			for _, tmp = range tmps {
				if cf(tmp) {
					tc++
				}
			}
			fmt.Printf("%s: %d\n", cn, tc)
		}
	}
}

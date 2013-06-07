package geoutil

import (
	"fmt"
	"math"
	"tshared/coreutil"
)

type GeoNamesZipRecord struct {
	Id int
	C int
	Z string
	N string
	L []interface{}
	A int
}

type GeoNamesAdminRecord struct {
	Id int
	A []interface{}
	G int
	N string
	Na string
}

type GeoNamesNameRecord struct {
	Id int
	G int
	L []interface{}
	N string
	Na string
	An []interface{}
	T int
	C int
	A int
	F int
}

var (
	LatMin float64 = -60
	LatMax float64 = 84
	LonMin float64 = -180
	LonMax float64 = 179.999999999999997
)

func LoLaFileName (lo, la float64) string {
	var err error
	if lo, err = LonTileDegree(lo, 0); err != nil {
		panic(err)
	}
	if la, err = LatTileDegree(la, 0); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s%02.0f%s%03.0f", coreutil.Ifstr(la < 0, "S", "N"), math.Abs(la), coreutil.Ifstr(lo < 0, "W", "E"), math.Abs(lo))
}

func LatTileDegree (lat, plus float64) (float64, error) {
	return TileDegree(lat, LatMin, LatMax, plus)
}

func LonTileDegree (lon, plus float64) (float64, error) {
	return TileDegree(lon, LonMin, LonMax, plus)
}

func TileDegree (val, vmin, vmax, plus float64) (float64, error) {
	var err error = nil
	if (val >= vmin) && (val <= vmax) {
		val = plus + math.Floor(float64(val))
	} else {
		err = fmt.Errorf("%v out of bounds (min: %v max: %v)", val, vmin, vmax)
	}
	return val, err
}

func ToKph (mps float64) float64 {
	return mps * 3.6
}

func ToMps (kph float64) float64 {
	return kph / 3.6
}

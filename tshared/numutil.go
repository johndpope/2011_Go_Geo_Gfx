package numutil

import (
	"math"
	"strconv"
)

type Mat4 []float64

func NewMat4Frustum (left, right, bottom, top, near, far float64) Mat4 {
	var rl, tb, fn, n2 = right - left, top - bottom, far - near, near * 2
	return Mat4 {
		n2 / rl, 0, 0, 0,
		0, n2 / tb, 0, 0,
		(right + left) / rl, (top + bottom) / tb, -(far + near) / fn, -1,
		0, 0, -(far * n2) / fn, 0,
	}
}

func NewMat4Identity () Mat4 {
	return Mat4 {
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewMat4LookAt (t, d, k Vec3) Mat4 {
	var z = d.NormalizedScaled(-1)
	var x = d.CrossNormalized(&k)
	var y = z.Cross(&x)
	return Mat4 {
		x.X, y.X, z.X, -t.X,
		x.Y, y.Y, z.Y, -t.X,
		x.Z, y.Z, z.Z, -t.X,
		0, 0, 0, 1,
	}
}

func NewMat4LookAt2 (pos, xaxis, yaxis, zaxis Vec3, tmp []float64) Mat4 {
	var mat = Mat4 {
		xaxis.X, xaxis.Y, xaxis.Z, 0,
		yaxis.X, yaxis.Y, yaxis.Z, 0,
		zaxis.X, zaxis.Y, zaxis.Z, 0,
		0, 0, 0, 0,
	}
	mat.Invert(tmp)
	mat[3], mat[7], mat[11] = pos.X, pos.Y, pos.Z
	return mat
}

func NewMat4Perspective (yFov, aspect, near, far float64) Mat4 {
	var f = math.Tan((math.Pi / 2) - yFov)
	return Mat4 {
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) / (near - far), (2 * far * near) / (near - far),
		0, 0, -1, 0,
	}
}

func NewMat4RotationX (amount float64) Mat4 {
	var c, s = math.Cos(amount), math.Sin(amount)
	return Mat4 {
		1, 0, 0, 0,
		0, c, -s, 0,
		0, s, c, 0,
		0, 0, 0, 1,
	}
}

func NewMat4RotationY (amount float64) Mat4 {
	var c, s = math.Cos(amount), math.Sin(amount)
	return Mat4 {
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

func NewMat4RotationZ (amount float64) Mat4 {
	var c, s = math.Cos(amount), math.Sin(amount)
	return Mat4 {
		c, -s, 0, 0,
		s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewMat4Scaled (x, y, z float64) Mat4 {
	return Mat4 {
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func NewMat4Translation (x, y, z float64) Mat4 {
	return Mat4 {
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}
}

func (me Mat4) Invert (tmp []float64) {
	tmp[17], tmp[18], tmp[19] = -me[3], -me[7], -me[11]
	for h := 0; h < 3; h++ { for v := 0; v < 3; v++ { tmp[h + v * 4] = me[v + h * 4] } }
	for i := 0; i < 11; i++ { me[i] = tmp[i] }
	me[3] = tmp[17] * me[0] + tmp[18] * me[1] + tmp[19] * me[2]
	me[7] = tmp[17] * me[4] + tmp[18] * me[5] + tmp[19] * me[6]
	me[11] = tmp[17] * me[8] + tmp[18] * me[9] + tmp[19] * me[10]
}

func (me Mat4) Mult (mat Mat4) Mat4 {
	return Mat4 {
		(me[0] * mat[0]) + (me[1] * mat[4]) + (me[2] * mat[8]) + (me[3] * mat[12]),
		(me[0] * mat[1]) + (me[1] * mat[5]) + (me[2] * mat[9]) + (me[3] * mat[13]),
		(me[0] * mat[2]) + (me[1] * mat[6]) + (me[2] * mat[10]) + (me[3] * mat[14]),
		(me[0] * mat[3]) + (me[1] * mat[7]) + (me[2] * mat[11]) + (me[3] * mat[15]),

		(me[4] * mat[0]) + (me[5] * mat[4]) + (me[6] * mat[8]) + (me[7] * mat[12]),
		(me[4] * mat[1]) + (me[5] * mat[5]) + (me[6] * mat[9]) + (me[7] * mat[13]),
		(me[4] * mat[2]) + (me[5] * mat[6]) + (me[6] * mat[10]) + (me[7] * mat[14]),
		(me[4] * mat[3]) + (me[5] * mat[7]) + (me[6] * mat[11]) + (me[7] * mat[15]),

		(me[8] * mat[0]) + (me[9] * mat[4]) + (me[10] * mat[8])+ (me[11] * mat[12]),
		(me[8] * mat[1]) + (me[9] * mat[5]) + (me[10] * mat[9])+ (me[11] * mat[13]),
		(me[8] * mat[2]) + (me[9] * mat[6]) + (me[10] * mat[10]) + (me[11] * mat[14]),
		(me[8] * mat[3]) + (me[9] * mat[7]) + (me[10] * mat[11]) + (me[11] * mat[15]),

		(me[12] * mat[0]) + (me[13] * mat[4]) + (me[14] * mat[8])+ (me[15] * mat[12]),
		(me[12] * mat[1]) + (me[13] * mat[5]) + (me[14] * mat[9])+ (me[15] * mat[13]),
		(me[12] * mat[2]) + (me[13] * mat[6]) + (me[14] * mat[10]) + (me[15] * mat[14]),
		(me[12] * mat[3]) + (me[13] * mat[7]) + (me[14] * mat[11]) + (me[15] * mat[15]),
	}
}

func (me Mat4) Transposed () Mat4 {
	return Mat4 {
		me[0], me[4], me[8], me[12],
		me[1], me[5], me[9], me[13],
		me[2], me[6], me[10], me[14],
		me[3], me[7], me[11], me[15],
	}
}

type Vec2 struct {
	X float64
	Y float64
}

func NewVec2 (vals ... string) (Vec2, error) {
	var err error
	var f Vec2
	if f.X, err = strconv.ParseFloat(vals[0], 64); err == nil {
		f.Y, err = strconv.ParseFloat(vals[1], 64)
	}
	return f, err
}

func (me Vec2) Div (vec Vec2) Vec2 {
	return Vec2 { me.X / vec.X, me.Y / vec.Y }
}

func (me Vec2) Dot (vec Vec2) float64 {
	return (me.X * vec.X) + (me.Y * vec.Y)
}

func (me Vec2) Length () float64 {
	return (me.X * me.X) + (me.Y * me.Y)
}

func (me Vec2) Magnitude () float64 {
	return math.Sqrt(me.Length())
}

func (me Vec2) Mult (vec Vec2) Vec2 {
	return Vec2 { me.X * vec.X, me.Y * vec.Y }
}

func (me *Vec2) Normalize () {
	var l = 1 / me.Magnitude()
	me.X *= l
	me.Y *= l
}

func (me Vec2) Normalized () Vec2 {
	var l = 1 / me.Magnitude()
	return Vec2 { me.X * l, me.Y * l }
}

func (me Vec2) NormalizedScaled (by float64) Vec2 {
	var l = 1 / me.Magnitude()
	return Vec2 { me.X * l * by, me.Y * l * by }
}

func (me Vec2) Scaled (by float64) Vec2 {
	return Vec2 { me.X * by, me.Y * by }
}

func (me Vec2) Sub (vec Vec2) Vec2 {
	return Vec2 { me.X - vec.X, me.Y - vec.Y }
}

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func (me *Vec3) Add (vec *Vec3) {
	me.X, me.Y, me.Z = me.X + vec.X, me.Y + vec.Y, me.Z + vec.Z
}

func (me *Vec3) Add1 (add float64)  {
	me.X, me.Y, me.Z = me.X + add, me.Y + add, me.Z + add
}

func (me *Vec3) AllEqual (val float64) bool {
	return (me.X == val) && (me.Y == val) && (me.Z == val)
}

func (me *Vec3) AllGreaterOrEqual (test *Vec3) bool {
	return (me.X >= test.X) && (me.Y >= test.Y) && (me.Z >= test.Z)
}

func (me *Vec3) AllInRange (min, max float64) bool {
	return (me.X >= min) && (me.X < max) && (me.Y >= min) && (me.Y < max) && (me.Z >= min) && (me.Z < max)
}

func (me *Vec3) AllInside (min, max *Vec3) bool {
	return (me.X > min.X) && (me.X < max.X) && (me.Y > min.Y) && (me.Y < max.Y) && (me.Z > min.Z) && (me.Z < max.Z)
}

func (me *Vec3) AllLessOrEqual (test *Vec3) bool {
	return (me.X <= test.X) && (me.Y <= test.Y) && (me.Z <= test.Z)
}

func (me *Vec3) Cross (vec *Vec3) Vec3 {
	return Vec3 { (me.Y * vec.Z) - (me.Z * vec.Y), (me.Z * vec.X) - (me.X * vec.Z), (me.X * vec.Y) - (me.Y * vec.X) }
}

func (me *Vec3) CrossNormalized (vec *Vec3) Vec3 {
	var r = Vec3 { (me.Y * vec.Z) - (me.Z * vec.Y), (me.Z * vec.X) - (me.X * vec.Z), (me.X * vec.Y) - (me.Y * vec.X) }
	r.Normalize()
	return r
}

func (me *Vec3) Div (vec *Vec3) Vec3 {
	return Vec3 { me.X / vec.X, me.Y / vec.Y, me.Z / vec.Z }
}

func (me *Vec3) Div1 (val float64) Vec3 {
	return Vec3 { me.X / val, me.Y / val, me.Z / val }
}

func (me *Vec3) Dot (vec *Vec3) float64 {
	return (me.X * vec.X) + (me.Y * vec.Y) + (me.Z * vec.Z)
}

func (me *Vec3) DotSub (vec1, vec2 *Vec3) float64 {
	return (me.X * (vec1.X - vec2.X)) + (me.Y * (vec1.Y - vec2.Y)) + (me.Z * (vec1.Z - vec2.Z))
}

func (me *Vec3) Equals (vec *Vec3) bool {
	return (me.X == vec.X) && (me.Y == vec.Y) && (me.Z == vec.Z)
}

func (me *Vec3) Inv () Vec3 {
	return Vec3 { 1 / me.X, 1 / me.Y, 1 / me.Z }
}

func (me *Vec3) Length () float64 {
	return (me.X * me.X) + (me.Y * me.Y) + (me.Z * me.Z)
}

func (me *Vec3) Magnitude () float64 {
	return math.Sqrt(me.Length())
}

func (me *Vec3) MakeFinite (v *Vec3) {
	if math.IsInf(me.X, 0) { me.X = v.X }
	if math.IsInf(me.Y, 0) { me.Y = v.Y }
	if math.IsInf(me.Z, 0) { me.Z = v.Z }
}

func (me *Vec3) Max () float64 {
	return math.Max(me.X, math.Max(me.Y, me.Z))
}

func (me *Vec3) Min () float64 {
	return math.Min(me.X, math.Min(me.Y, me.Z))
}

func (me *Vec3) Mult (vec *Vec3) Vec3 {
	return Vec3 { me.X * vec.X, me.Y * vec.Y, me.Z * vec.Z }
}

func (me *Vec3) MultMat (mat Mat4) {
	me.X = (mat[0] * me.X) + (mat[1] * me.Y) + (mat[2] * me.Z) // + mat[3]
	me.Y = (mat[4] * me.X) + (mat[5] * me.Y) + (mat[6] * me.Z) // + mat[7]
	me.Z = (mat[8] * me.X) + (mat[9] * me.Y) + (mat[10] * me.Z) // + mat[11]
	// me.X = (mat[0] * me.X) + (mat[4] * me.Y) + (mat[8] * me.Z) // + mat[12]
	// me.Y = (mat[1] * me.X) + (mat[5] * me.Y) + (mat[9] * me.Z) // + mat[13]
	// me.Z = (mat[2] * me.X) + (mat[6] * me.Y) + (mat[10] * me.Z) // + mat[14]
}

func (me *Vec3) MultMat2 (mat Mat4) {
	me.X = mat[0] * me.X + mat[1] * me.Y + mat[2] * me.Z + mat[3]
	me.Y = mat[4] * me.X + mat[5] * me.Y + mat[6] * me.Z + mat[7]
	me.Z = mat[8] * me.X + mat[9] * me.Y + mat[10] * me.Z + mat[11]
}

func (me *Vec3) Normalize () {
	var le = me.Magnitude()
	if (le == 0) || (le == 1) {
		me.X, me.Y, me.Z = le, le, le
	} else {
		le = 1 / le
		me.X, me.Y, me.Z = me.X * le, me.Y * le, me.Z * le
	}
}

func (me *Vec3) Normalized () Vec3 {
	var le = me.Magnitude()
	if (le == 0) || (le == 1) {
		return Vec3 { le, le, le }
	}
	le = 1 / le
	return Vec3 { me.X * le, me.Y * le, me.Z * le }
}

func (me *Vec3) NormalizedScaled (mul float64) Vec3 {
	var vec = me.Normalized()
	vec.Scale(mul)
	return vec
}

func (me *Vec3) Scale (mul float64) {
	me.X, me.Y, me.Z = me.X * mul, me.Y * mul, me.Z * mul
}

func (me *Vec3) ScaleAdd (mul, add *Vec3) {
	me.X, me.Y, me.Z = (me.X * mul.X) + add.X, (me.Y * mul.Y) + add.Y, (me.Z * mul.Z) + add.Z
}

func (me *Vec3) Scaled (by float64) Vec3 {
	return Vec3 { me.X * by, me.Y * by, me.Z * by }
}

func (me *Vec3) ScaledAdded (mul float64, add *Vec3) Vec3 {
	return Vec3 { (me.X * mul) + add.X, (me.Y * mul) + add.Y, (me.Z * mul) + add.Z }
}

func (me *Vec3) SetFrom (vec *Vec3) {
	me.X, me.Y, me.Z = vec.X, vec.Y, vec.Z
}

func (me *Vec3) SetFromAdd (vec1, vec2 *Vec3) {
	me.X, me.Y, me.Z = vec1.X + vec2.X, vec1.Y + vec2.Y, vec1.Z + vec2.Z
}

func (me *Vec3) SetFromAddMult (add, mul1, mul2 *Vec3) {
	me.X, me.Y, me.Z = add.X + (mul1.X * mul2.X), add.Y + (mul1.Y * mul2.Y), add.Z + (mul1.Z * mul2.Z)
}

func (me *Vec3) SetFromAddMult1 (vec1, vec2 *Vec3, mul float64) {
	me.X, me.Y, me.Z = vec1.X + (vec2.X * mul), vec1.Y + (vec2.Y * mul), vec1.Z + (vec2.Z * mul)
}

func (me *Vec3) SetFromCos (vec *Vec3) {
	me.X, me.Y, me.Z = math.Cos(vec.X), math.Cos(vec.Y), math.Cos(vec.Z)
}

func (me *Vec3) SetFromDegToRad (deg *Vec3) {
	me.X, me.Y, me.Z = DegToRad(deg.X), DegToRad(deg.Y), DegToRad(deg.Z)
}

func (me *Vec3) SetFromEpsilon () {
	if math.Abs(me.X) < Epsilon { me.X = Epsilon }
	if math.Abs(me.Y) < Epsilon { me.Y = Epsilon }
	if math.Abs(me.Z) < Epsilon { me.Z = Epsilon }
}

func (me *Vec3) SetFromInv (vec *Vec3) {
	me.X, me.Y, me.Z = 1 / vec.X, 1 / vec.Y, 1 / vec.Z
}

func (me *Vec3) SetFromMult (v1, v2 *Vec3) {
	me.X, me.Y, me.Z = v1.X * v2.X, v1.Y * v2.Y, v1.Z * v2.Z
}

func (me *Vec3) SetFromMult1 (vec *Vec3, mul float64) {
	me.X, me.Y, me.Z = vec.X * mul, vec.Y * mul, vec.Z * mul
}

func (me *Vec3) SetFromMult1Sub (vec1, vec2 *Vec3, mul float64) {
	me.X, me.Y, me.Z = (vec1.X - vec2.X) * mul, (vec1.Y - vec2.Y) * mul, (vec1.Z - vec2.Z) * mul
}

func (me *Vec3) SetFromRotation (pos, rotCos, rotSin Vec3) {
	var tmp = ((pos.Y * rotSin.X) + (pos.Z * rotCos.X))
	me.X = (pos.X * rotCos.Y) + (tmp * rotSin.Y)
	me.Y = (pos.Y * rotCos.X) - (pos.Z * rotSin.X)
	me.Z = (-pos.X * rotSin.Y) + (tmp * rotCos.Y)
}

func (me *Vec3) SetFromSin (vec *Vec3) {
	me.X, me.Y, me.Z = math.Sin(vec.X), math.Sin(vec.Y), math.Sin(vec.Z)
}

func (me *Vec3) SetFromStep1 (edge float64, vec, zero, one *Vec3) {
	if vec.X < edge { me.X = zero.X } else { me.X = one.X }
	if vec.Y < edge { me.Y = zero.Y } else { me.Y = one.Y }
	if vec.Z < edge { me.Z = zero.Z } else { me.Z = one.Z }
}

func (me *Vec3) SetFromSub (vec1, vec2 *Vec3) {
	me.X, me.Y, me.Z = vec1.X - vec2.X, vec1.Y - vec2.Y, vec1.Z - vec2.Z
}

func (me *Vec3) SetFromSubMult (sub1, sub2, mul *Vec3) {
	me.X, me.Y, me.Z = (sub1.X - sub2.X) * mul.X, (sub1.Y - sub2.Y) * mul.Y, (sub1.Z - sub2.Z) * mul.Z
}

func (me *Vec3) SetFromSubMult1 (vec1, vec2 *Vec3, mul float64) {
	me.X, me.Y, me.Z = vec1.X - (vec2.X * mul), vec1.Y - (vec2.Y * mul), vec1.Z - (vec2.Z * mul)
}

func (me *Vec3) Sign () Vec3 {
	var r = Vec3 { 0, 0, 0 }
	if me.X < 0 { r.X = -1 } else if me.X > 0 { r.X = 1 }
	if me.Y < 0 { r.Y = -1 } else if me.Y > 0 { r.Y = 1 }
	if me.Z < 0 { r.Z = -1 } else if me.Z > 0 { r.Z = 1 }
	return r
}

func (me *Vec3) Sub (vec *Vec3) Vec3 {
	return Vec3 { me.X - vec.X, me.Y - vec.Y, me.Z - vec.Z }
}

func (me *Vec3) SubDivMult (sub, div, mul *Vec3) Vec3 {
	return Vec3 { ((me.X - sub.X) / div.X) * mul.X, ((me.Y - sub.Y) / div.Y) * mul.Y, ((me.Z - sub.Z) / div.Z) * mul.Z }
}

func (me *Vec3) SubDot (vec *Vec3) float64 {
	return ((me.X - vec.X) * (me.X - vec.X)) + ((me.Y - vec.Y) * (me.Y - vec.Y)) + ((me.Z - vec.Z) * (me.Z - vec.Z))
}

func (me *Vec3) SubFloorDivMult (floorDiv, mul float64) Vec3 {
	return me.Sub(&Vec3 { math.Floor(me.X / floorDiv) * mul, math.Floor(me.Y / floorDiv) * mul, math.Floor(me.Z / floorDiv) * mul })
}

func (me *Vec3) SubFrom (val float64) Vec3 {
	return Vec3 { val - me.X, val - me.Y, val - me.Z }
}

func (me *Vec3) SubMult (vec *Vec3, val float64) Vec3 {
	return Vec3 { (me.X - vec.X) * val, (me.Y - vec.Y) * val, (me.Z - vec.Z) * val }
}

func (me *Vec3) SubVec (vec *Vec3) {
	me.X, me.Y, me.Z = me.X - vec.X, me.Y - vec.Y, me.Z - vec.Z
}

func (me *Vec3) SwapSigns () {
	me.X, me.Y, me.Z = -me.X, -me.Y, -me.Z
}

func (me *Vec3) ToInts () []int {
	return []int { int(me.X), int(me.Y), int(me.Z) }
}

const (
	Pi180 = math.Pi / 180
	PiInv = 0.5 / math.Pi
)

var (
	Epsilon float64 = 0
	EpsilonMax float64 = 0
	Infinity float64
	NegInfinity float64
)

func AllEqual (test float64, vals ... float64) bool {
	for i := 0; i < len(vals); i++ { if vals[i] != test { return false } }
	return true
}

func DegToRad (deg float64) float64 {
	return Pi180 * deg
}

func Din1 (val, max float64) float64 {
	return 1 / (max / val)
}

func Fin1 (val, max float32) float32 {
	return 1 / (max / val)
}

func Iin1 (val, max int) int {
	return 1 / (max / val)
}

func IsEveni (val int) bool {
	return (math.Mod(float64(val), 2) == 0)
}

func IsInt (val float64) bool {
	_, f := math.Modf(val)
	return f == 0
}

func IsMod0 (v, m int) bool {
	return math.Mod(float64(v), float64(m)) == 0
}

func IsVec2 (any interface{}) bool {
	if any != nil {
		if _, isT := any.(Vec2); isT {
			return true
		}
	}
	return false
}

func Lin1 (val, max int64) int64 {
	return 1 / (max / val)
}

func Absi (v int32) int32 {
	return v - (v ^ (v >> 31))
}

func Absl (v int64) int64 {
	return v - (v ^ (v >> 63))
}

func Max (x, y float64) float64 {
	return 0.5 * (x + y + math.Abs(x - y))
}

func Min (x, y float64) float64 {
	return 0.5 * (x + y - math.Abs(x - y))
}

func Mini (v1, v2 int) int {
	if v1 < v2 { return v1 }
	return v2
}

func Mix (x, y, a float64) float64 {
	return (x * y) + ((1 - y) * a)
}

func RadToDeg (rad float64) float64 {
	return rad * Pi180
}

func Round (v float64) float64 {
	var frac float64
	if _, frac = math.Modf(v); frac >= 0.5 { return math.Ceil(v) }
	return math.Floor(v)
}

func Step (edge, x float64) int {
	if x < edge { return 0 }
	return 1
}

func init () {
	var eps, i float64
	Infinity, NegInfinity = math.Inf(1), math.Inf(-1)
	for i = 0; i <= 8192; i++ {
		if eps = math.Pow(2, -i); eps == 0 { break }
		if i == 23 {
			Epsilon = eps
		} else {
			EpsilonMax = eps
		}
	}
}

package gltypes

import (
	"math"

	gl "github.com/chsc/gogl/gl42"
)

type GlUint2 struct {
	GlPtr1, GlPtr2 gl.Uint
}

type Gmat4 []gl.Float

func NewGmat4Identity () Gmat4 {
	return Gmat4 {
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewGmat4LookAt (t, d, k Gvec3) Gmat4 {
	var z = d.NormalizedScaled(-1)
	var x = d.CrossNormalized(k)
	var y = z.Cross(x)
	return Gmat4 {
		x.X, y.X, z.X, -t.X,
		x.Y, y.Y, z.Y, -t.X,
		x.Z, y.Z, z.Z, -t.X,
		0, 0, 0, 1,
	}
}

func NewGmat4Perspective (yFov, aspect, zNear, zFar gl.Float) Gmat4 {
	var f = gl.Float(math.Tan((math.Pi / 2) - float64(yFov)))
	return Gmat4 {
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (zFar + zNear) / (zNear - zFar), (2 * zFar * zNear) / (zNear - zFar),
		0, 0, -1, 0,
	}
}

func NewGmat4Rotation (angle gl.Float, vec Gvec3) Gmat4 {
	var c, s = gl.Float(math.Cos(float64(angle))), gl.Float(math.Sin(float64(angle)))
	return Gmat4 {
		c, -s, 0, 0,
		s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewGmat4RotationX (amount gl.Float) Gmat4 {
	var c, s = gl.Float(math.Cos(float64(amount))), gl.Float(math.Sin(float64(amount)))
	return Gmat4 {
		1, 0, 0, 0,
		0, c, -s, 0,
		0, s, c, 0,
		0, 0, 0, 1,
	}
}

func NewGmat4RotationY (amount gl.Float) Gmat4 {
	var c, s = gl.Float(math.Cos(float64(amount))), gl.Float(math.Sin(float64(amount)))
	return Gmat4 {
		c, 0, s, 0,
		0, 1, 0, 0,
		-s, 0, c, 0,
		0, 0, 0, 1,
	}
}

func NewGmat4RotationZ (amount gl.Float) Gmat4 {
	var c, s = gl.Float(math.Cos(float64(amount))), gl.Float(math.Sin(float64(amount)))
	return Gmat4 {
		c, -s, 0, 0,
		s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewGmat4Scaled (x, y, z gl.Float) Gmat4 {
	return Gmat4 {
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func NewGmat4Translation (x, y, z gl.Float) Gmat4 {
	return Gmat4 {
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}
}

func (me Gmat4) Mult (mat Gmat4) Gmat4 {
	return Gmat4 {
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

func (me Gmat4) Transposed () Gmat4 {
	return Gmat4 {
		me[0], me[4], me[8], me[12],
		me[1], me[5], me[9], me[13],
		me[2], me[6], me[10], me[14],
		me[3], me[7], me[11], me[15],
	}
}

type Gvec3 struct {
	X gl.Float
	Y gl.Float
	Z gl.Float
}

func (me Gvec3) Cross (vec Gvec3) Gvec3 {
	return Gvec3 { (me.Y * vec.Z) - (me.Z * vec.Y), (me.Z * vec.X) - (me.X * vec.Z), me.X - vec.Y - (me.Y * vec.X) }
}

func (me Gvec3) CrossNormalized (vec Gvec3) Gvec3 {
	var r = Gvec3 { (me.Y * vec.Z) - (me.Z * vec.Y), (me.Z * vec.X) - (me.X * vec.Z), me.X - vec.Y - (me.Y * vec.X) }
	var l = 1 / r.Magnitude()
	r.X, r.Y, r.Z = r.X * l, r.Y * l, r.Z * l
	return r
}

func (me Gvec3) Dot (vec Gvec3) gl.Float {
	return (me.X * vec.X) + (me.Y * vec.Y) + (me.Z * vec.Z)
}

func (me Gvec3) LenSqrt () gl.Float {
	return (me.X * me.X) + (me.Y * me.Y) + (me.Z * me.Z)
}

func (me Gvec3) Magnitude () gl.Float {
	return gl.Float(math.Sqrt(float64(me.LenSqrt())))
}

func (me Gvec3) Mult (vec Gvec3) Gvec3 {
	return Gvec3 { me.X * vec.X, me.Y * vec.Y, me.Z * vec.Z }
}

func (me *Gvec3) Normalize () {
	var l = 1 / me.Magnitude()
	me.X *= l
	me.Y *= l
	me.Z *= l
}

func (me Gvec3) Normalized () Gvec3 {
	var l = 1 / me.Magnitude()
	return Gvec3 { me.X * l, me.Y * l, me.Z * l }
}

func (me Gvec3) NormalizedScaled (by gl.Float) Gvec3 {
	var l = 1 / me.Magnitude()
	return Gvec3 { me.X * l * by, me.Y * l * by, me.Z * l * by }
}

func (me Gvec3) Scaled (by gl.Float) Gvec3 {
	return Gvec3 { me.X * by, me.Y * by, me.Z * by }
}

func (me Gvec3) Sub (vec Gvec3) Gvec3 {
	return Gvec3 { me.X - vec.X, me.Y - vec.Y, me.Z - vec.Z }
}

type ShaderManager struct {
	Canvas, Cast, PostBlur, PostBright, PostFx, PostLum2, PostLum3, Texture *ShaderProgram
}

func NewShaderManager () *ShaderManager {
	return &ShaderManager { nil, nil, nil, nil, nil, nil, nil, nil }
}

func (me *ShaderManager) CleanUp () {
	var cleanUp = func (sprog **ShaderProgram) {
		var sp *ShaderProgram = *sprog
		if sp != nil { sp.CleanUp(); *sprog = nil }
	}
	cleanUp(&me.Cast)
	cleanUp(&me.Canvas)
	cleanUp(&me.PostBlur)
	cleanUp(&me.PostBright)
	cleanUp(&me.PostLum2)
	cleanUp(&me.PostLum3)
	cleanUp(&me.PostFx)
	cleanUp(&me.Texture)
}

type ShaderProgram struct {
	Name string
	Program, FShader, VShader gl.Uint
	UnifCamLook, UnifCamPos, UnifScreen, UnifTex0, UnifTex1, UnifTex2, UnifTime gl.Int
}

func NewShaderProgram (name string, glProg, glFShader, glVShader gl.Uint) *ShaderProgram {
	return &ShaderProgram { name, glProg, glFShader, glVShader, 0, 0, 0, 0, 0, 0, 0 }
}

func (me *ShaderProgram) CleanUp () {
	gl.DetachShader(me.Program, me.FShader)
	gl.DetachShader(me.Program, me.VShader)
	gl.DeleteShader(me.FShader)
	gl.DeleteShader(me.VShader)
	gl.DeleteProgram(me.Program)
}

var (
	MatrixIdentity = NewGmat4Identity()
	SizeOfGlUint gl.Sizeiptr = 4
)

func Fin1 (val, max gl.Float) gl.Float {
	return 1 / (max / val)
}

func Ife (cond bool, ifTrue, ifFalse gl.Enum) gl.Enum {
	if cond { return ifTrue }
	return ifFalse
}

func Ifui (cond bool, ifTrue, ifFalse gl.Uint) gl.Uint {
	if cond { return ifTrue }
	return ifFalse
}

func InSliceAt (vals []gl.Enum, val gl.Enum) int {
	for i, v := range vals {
		if v == val {
			return i
		}
	}
	return -1
}

func OffsetIntPtr (ptr gl.Pointer, offset gl.Sizei) gl.Intptr {
	return gl.Intptr(uintptr(ptr) + uintptr(offset))
}

func OffsetPointer (ptr gl.Pointer, offset uint) gl.Pointer {
	return gl.Pointer(uintptr(ptr) + uintptr(offset))
}

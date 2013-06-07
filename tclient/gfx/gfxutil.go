package gfxutil

import (
	"math"

	num "tshared/numutil"
)

type TColor struct {
	R, G, B, A float64
}

func (me *TColor) Blend (fore, back *TColor) {
	me.A = fore.A + (back.A * (1 - fore.A))
	if me.A == 0 {
		me.Reset()
	} else {
		me.R = ((fore.R * fore.A) + (back.R * back.A * (1 - fore.A))) / me.A
		me.G = ((fore.G * fore.A) + (back.G * back.A * (1 - fore.A))) / me.A
		me.B = ((fore.B * fore.A) + (back.B * back.A * (1 - fore.A))) / me.A
	}
}

func (me *TColor) Equals (col *TColor) bool {
	return ((me.A == 0) && (col.A == 0)) || ((me.R == col.R) && (me.G == col.G) && (me.B == col.B) && (me.A == col.A))
}

func (me *TColor) Mix (fore, back *TColor) {
	me.R = (fore.R * fore.A) + (back.R * back.A * (1 - fore.A))
	me.G = (fore.G * fore.A) + (back.G * back.A * (1 - fore.A))
	me.B = (fore.B * fore.A) + (back.B * back.A * (1 - fore.A))
	me.A = fore.A + (back.A * (1 - fore.A))
}

func (me *TColor) Mult (col *TColor) {
	me.R, me.G, me.B, me.A = me.R * col.R, me.G * col.G, me.B * col.B, me.A * col.A
}

func (me *TColor) Mult1 (val float64) {
	me.R, me.G, me.B = me.R * val, me.G * val, me.B * val
}

func (me *TColor) PreMult (fore, back *TColor) {
	me.A = fore.A + (back.A * (1 - fore.A))
	me.R = fore.R + (back.R * (1 - fore.A))
	me.G = fore.G + (back.G * (1 - fore.A))
	me.B = fore.B + (back.B * (1 - fore.A))
}

func (me *TColor) Reset () {
	me.R, me.G, me.B, me.A = 0, 0, 0, 0
}

func (me *TColor) SetFrom (col *TColor) {
	me.R, me.G, me.B, me.A = col.R, col.G, col.B, col.A
}

func (me *TColor) SetFromDiv (col *TColor, div float64) {
	me.R, me.G, me.B, me.A = col.R / div, col.G / div, col.B / div, col.A / div
}

type TLight struct {
	Pos, Size, Rot, RotPos num.Vec3
	Col TColor
	Intensity float64

	tmp float64
	rotRad, rotCos, rotSin num.Vec3
}

func NewLight (pos, size num.Vec3, col TColor, intensity float64) *TLight {
	var me = &TLight {}
	me.Pos, me.Size, me.Col, me.Intensity = pos, size, col, intensity
	me.rotCos.X, me.rotCos.Y, me.rotCos.Z = 1, 1, 1
	return me
}

func (me *TLight) SetRotation (x, y float64) {
	me.Rot.X, me.Rot.Y, me.Rot.Z = x, y, 0
	me.rotRad.SetFromDegToRad(&me.Rot)
	me.rotCos.SetFromCos(&me.rotRad)
	me.rotSin.SetFromSin(&me.rotSin)
	me.RotPos.SetFromRotation(me.Pos, me.rotCos, me.rotSin)
}

type TObject struct {
	Col TColor
	Pos num.Vec3
	Refl float64
	Extent float64

	Min, Max, Size num.Vec3
	radius, radInv, radPow2 float64
}

func NewBox (col TColor, pos, size num.Vec3, refl float64) *TObject {
	var me = &TObject {}
	me.Col, me.Pos, me.Refl = col, pos, refl
	me.SetSize(size)
	return me
}

func NewSphere (col TColor, pos num.Vec3, radius, refl float64) *TObject {
	var me = &TObject {}
	me.Col, me.Pos, me.Refl = col, pos, refl
	me.SetRadius(radius)
	return me
}

// func (me *TObject) FindBoxExit (exit *num.Vec3) {
// 	picker.tmpVec1.SetFromStep1(0, &picker.Ray.InvDir, &me.Min, &me.Max)
// 	picker.tmpVec2.SetFromSubMult(&picker.tmpVec1, &picker.Exit, &picker.Ray.InvDir)
// 	picker.Exit.SetFromAddMult1(&picker.Exit, &picker.Ray.Dir, picker.tmpVec2.Min())
// }

func (me *TObject) Pick (picker *TPicker) {
	picker.IsPick = false
	if me.radius != 0 {
		// SPHERE
		picker.tmpU = picker.Ray.dirMagnitude.DotSub(&me.Pos, &picker.Ray.Pos)
		if picker.tmpU > 0 {
			picker.tmpVec1.SetFromAddMult1(&picker.Ray.Pos, &picker.Ray.dirMagnitude, picker.tmpU)
			picker.tmpU = picker.tmpVec1.SubDot(&me.Pos)
			if picker.tmpU < me.radPow2 {
				picker.tmpU = math.Sqrt(me.radPow2 - picker.tmpU)
				picker.Entry.SetFromSubMult1(&picker.tmpVec1, &picker.Ray.dirMagnitude, picker.tmpU)
				picker.Dist = math.Sqrt(picker.Entry.SubDot(&picker.Ray.Pos))
				if (!picker.IgnoreDist) && (picker.Dist > picker.Ray.magnitude) { return }
				picker.Dir.SetFromMult1Sub(&picker.Entry, &me.Pos, me.radInv)
				picker.IsPick = true
				return
			}
		}
	} else {
		// BOX
		if picker.FullyInside = picker.Ray.Pos.AllInside(&me.Min, &me.Max); picker.FullyInside {
			picker.Dir, picker.Dist, picker.Entry, picker.NumIndices = picker.Ray.Dir, -1, picker.Ray.Pos, 8
			picker.NodeIndices[0], picker.NodeIndices[1], picker.NodeIndices[2], picker.NodeIndices[3], picker.NodeIndices[4], picker.NodeIndices[5], picker.NodeIndices[6], picker.NodeIndices[7] = 0, 1, 2, 3, 4, 5, 6, 7
			picker.IsPick = true
			return
		}
		picker.Dir.X, picker.Dir.Y, picker.Dir.Z = 0, 0, 0
		if picker.Ray.DirXGeq0 && (me.Min.X >= picker.Ray.Pos.X) {
			picker.tmpU = (me.Min.X - picker.Ray.Pos.X) * picker.Ray.InvDir.X
			picker.Entry.X, picker.Entry.Y, picker.Entry.Z = me.Min.X, picker.Ray.Pos.Y + (picker.tmpU * picker.Ray.Dir.Y), picker.Ray.Pos.Z + (picker.tmpU * picker.Ray.Dir.Z)
			if (picker.Entry.Y >= me.Min.Y) && (picker.Entry.Y <= me.Max.Y) && (picker.Entry.Z >= me.Min.Z) && (picker.Entry.Z <= me.Max.Z) {
				picker.Dist = math.Sqrt(picker.Entry.SubDot(&picker.Ray.Pos))
				if (picker.IgnoreDist || (picker.Dist < picker.Ray.magnitude)) {
					picker.NumIndices, picker.NodeIndices[0], picker.NodeIndices[1], picker.NodeIndices[2], picker.NodeIndices[3] = 4, 3, 2, 1, 0
					picker.Dir.X = -1
					picker.IsPick = true
					return
				}
			}
		}
		if picker.Ray.DirXLeq0 && (me.Max.X <= picker.Ray.Pos.X) {
			picker.tmpU = (me.Max.X - picker.Ray.Pos.X) * picker.Ray.InvDir.X
			picker.Entry.X, picker.Entry.Y, picker.Entry.Z = me.Max.X, picker.Ray.Pos.Y + (picker.tmpU * picker.Ray.Dir.Y), picker.Ray.Pos.Z + (picker.tmpU * picker.Ray.Dir.Z)
			if (picker.Entry.Y >= me.Min.Y) && (picker.Entry.Y <= me.Max.Y) && (picker.Entry.Z >= me.Min.Z) && (picker.Entry.Z <= me.Max.Z) {
				picker.Dist = math.Sqrt(picker.Entry.SubDot(&picker.Ray.Pos))
				if (picker.IgnoreDist || (picker.Dist < picker.Ray.magnitude)) {
					picker.NumIndices, picker.NodeIndices[0], picker.NodeIndices[1], picker.NodeIndices[2], picker.NodeIndices[3] = 4, 7, 6, 5, 4
					picker.Dir.X = 1
					picker.IsPick = true
					return
				}
			}
		}
		if picker.Ray.DirYGeq0 && (me.Min.Y >= picker.Ray.Pos.Y) {
			picker.tmpU = (me.Min.Y - picker.Ray.Pos.Y) * picker.Ray.InvDir.Y
			picker.Entry.X, picker.Entry.Y, picker.Entry.Z = picker.Ray.Pos.X + (picker.tmpU * picker.Ray.Dir.X), me.Min.Y, picker.Ray.Pos.Z + (picker.tmpU * picker.Ray.Dir.Z)
			if (picker.Entry.X >= me.Min.X) && (picker.Entry.X <= me.Max.X) && (picker.Entry.Z >= me.Min.Z) && (picker.Entry.Z <= me.Max.Z) {
				picker.Dist = math.Sqrt(picker.Entry.SubDot(&picker.Ray.Pos))
				if (picker.IgnoreDist || (picker.Dist < picker.Ray.magnitude)) {
					picker.NumIndices, picker.NodeIndices[0], picker.NodeIndices[1], picker.NodeIndices[2], picker.NodeIndices[3] = 4, 0, 1, 4, 5
					picker.Dir.Y = -1
					picker.IsPick = true
					return
				}
			}
		}
		if picker.Ray.DirYLeq0 && (me.Max.Y <= picker.Ray.Pos.Y) {
			picker.tmpU = (me.Max.Y - picker.Ray.Pos.Y) * picker.Ray.InvDir.Y
			picker.Entry.X, picker.Entry.Y, picker.Entry.Z = picker.Ray.Pos.X + (picker.tmpU * picker.Ray.Dir.X), me.Max.Y, picker.Ray.Pos.Z + (picker.tmpU * picker.Ray.Dir.Z)
			if (picker.Entry.X >= me.Min.X) && (picker.Entry.X <= me.Max.X) && (picker.Entry.Z >= me.Min.Z) && (picker.Entry.Z <= me.Max.Z) {
				picker.Dist = math.Sqrt(picker.Entry.SubDot(&picker.Ray.Pos))
				if (picker.IgnoreDist || (picker.Dist < picker.Ray.magnitude)) {
					picker.NumIndices, picker.NodeIndices[0], picker.NodeIndices[1], picker.NodeIndices[2], picker.NodeIndices[3] = 4, 2, 3, 6, 7
					picker.Dir.Y = 1
					picker.IsPick = true
					return
				}
			}
		}
		if picker.Ray.DirZGeq0 && (me.Min.Z >= picker.Ray.Pos.Z) {
			picker.tmpU = (me.Min.Z - picker.Ray.Pos.Z) * picker.Ray.InvDir.Z
			picker.Entry.X, picker.Entry.Y, picker.Entry.Z = picker.Ray.Pos.X + (picker.tmpU * picker.Ray.Dir.X), picker.Ray.Pos.Y + (picker.tmpU * picker.Ray.Dir.Y), me.Min.Z
			if (picker.Entry.X >= me.Min.X) && (picker.Entry.X <= me.Max.X) && (picker.Entry.Y >= me.Min.Y) && (picker.Entry.Y <= me.Max.Y) {
				picker.Dist = math.Sqrt(picker.Entry.SubDot(&picker.Ray.Pos))
				if (picker.IgnoreDist || (picker.Dist < picker.Ray.magnitude)) {
					picker.NumIndices, picker.NodeIndices[0], picker.NodeIndices[1], picker.NodeIndices[2], picker.NodeIndices[3] = 4, 0, 2, 4, 6
					picker.Dir.Z = -1
					picker.IsPick = true
					return
				}
			}
		}
		if picker.Ray.DirZLeq0 && (me.Max.Z <= picker.Ray.Pos.Z) {
			picker.tmpU = (me.Max.Z - picker.Ray.Pos.Z) * picker.Ray.InvDir.Z
			picker.Entry.X, picker.Entry.Y, picker.Entry.Z = picker.Ray.Pos.X + (picker.tmpU * picker.Ray.Dir.X), picker.Ray.Pos.Y + (picker.tmpU * picker.Ray.Dir.Y), me.Max.Z
			if (picker.Entry.X >= me.Min.X) && (picker.Entry.X <= me.Max.X) && (picker.Entry.Y >= me.Min.Y) && (picker.Entry.Y <= me.Max.Y) {
				picker.Dist = math.Sqrt(picker.Entry.SubDot(&picker.Ray.Pos))
				if (picker.IgnoreDist || (picker.Dist < picker.Ray.magnitude)) {
					picker.NumIndices, picker.NodeIndices[0], picker.NodeIndices[1], picker.NodeIndices[2], picker.NodeIndices[3] = 4, 1, 3, 5, 7
					picker.Dir.Z = 1
					picker.IsPick = true
					return
				}
			}
		}
	}
}

func (me *TObject) SetRadius (radius float64) {
	me.radius = radius
	me.Min.Y = me.Pos.Y - radius
	me.Max.Y = me.Pos.Y + radius
	me.radInv = 1 / radius
	me.radPow2 = radius * radius
	me.Extent = radius * 2
}

func (me *TObject) SetSize (size num.Vec3) {
	me.Size = size
	me.Extent = (size.X + size.Y + size.Z) / 3
	me.Min.X = me.Pos.X // - (size.X / 2)
	me.Min.Y = me.Pos.Y // - (size.Y / 2)
	me.Min.Z = me.Pos.Z // - (size.Z / 2)
	me.Max.X = me.Pos.X + size.X // 2)
	me.Max.Y = me.Pos.Y + size.Y // 2)
	me.Max.Z = me.Pos.Z + size.Z // 2)
}

type TPicker struct {
	SceneObjects []*TObject
	SceneRange float64
	Picker *TPicker
	Obj *TObject
	Col TColor
	Dist, Refl float64
	FullyInside, IgnoreDist, IsPick, EntryXY, EntryXZ, EntryYZ bool
	NodeIndices []int
	NumIndices int
	Dir, Entry num.Vec3
	Ray *TRay

	cachedPick *TObject
	tmpVec1, tmpVec2, tmpVec3, tmin, tmax num.Vec3
	tmpDist, tmpU, tnear, tfar float64
	tsign int8
}

func NewPicker (sceneObjects []*TObject, sceneRange float64) *TPicker {
	var numIndices = 8
	var me = &TPicker {}
	me.SceneRange, me.SceneObjects, me.IgnoreDist, me.NumIndices, me.NodeIndices = sceneRange, sceneObjects, true, numIndices, make([]int, numIndices)
	me.Picker = &TPicker {}
	me.Picker.SceneRange, me.Picker.SceneObjects, me.Picker.IgnoreDist, me.Picker.NumIndices, me.Picker.NodeIndices = sceneRange, sceneObjects, true, numIndices, make([]int, numIndices)
	for i := 0; i < numIndices; i++ { me.NodeIndices[i] = i; me.Picker.NodeIndices[i] = i }
	return me
}

func (me *TPicker) Pick (ray *TRay) {
	me.Picker.Ray, me.tmpDist, me.Picker.IsPick = ray, me.SceneRange, false
	for _, obj := range me.SceneObjects {
		me.Picker.IgnoreDist = me.IgnoreDist
		if obj.Pick(me.Picker); me.Picker.IsPick && (me.Picker.Dist < me.tmpDist) {
			me.IsPick, me.tmpDist, me.Dist, me.Obj, me.Refl, me.Col = true, me.Picker.Dist, me.Picker.Dist, obj, obj.Refl, obj.Col
			me.Entry.X, me.Entry.Y, me.Entry.Z, me.Dir.X, me.Dir.Y, me.Dir.Z = me.Picker.Entry.X, me.Picker.Entry.Y, me.Picker.Entry.Z, me.Picker.Dir.X, me.Picker.Dir.Y, me.Picker.Dir.Z
		}
	}
	me.IsPick = me.IsPick && (me.tmpDist < me.SceneRange)
}

func (me *TPicker) PickCached (ray *TRay) {
	me.Ray, me.IsPick, me.IgnoreDist = ray, false, false
	if (me.cachedPick != nil) { if me.cachedPick.Pick(me); me.IsPick { me.IgnoreDist = true; return } }
	for _, obj := range me.SceneObjects {
		if (obj != me.cachedPick) {
			if obj.Pick(me); me.IsPick {
				me.cachedPick, me.IgnoreDist = obj, true
				return
			}
		}
	}
	me.IgnoreDist = true
}

type TRay struct {
	Pos, Dir, DirEpsilon, InvDir num.Vec3
	DirLength float64
	DirXGeq0, DirXLeq0, DirYGeq0, DirYLeq0, DirZGeq0, DirZLeq0 bool
	// DirXGt0, DirXLt0, DirYGt0, DirYLt0, DirZGt0, DirZLt0 bool

	invMagnitude, magnitude float64
	dirMagnitude num.Vec3
}

func (me *TRay) SetFrom (ray *TRay) {
	me.Pos, me.Dir, me.DirEpsilon, me.DirLength, me.DirXGeq0, me.DirXLeq0, me.DirYGeq0, me.DirYLeq0, me.DirZGeq0, me.DirZLeq0, me.invMagnitude, me.magnitude, me.dirMagnitude, me.InvDir = ray.Pos, ray.Dir, ray.DirEpsilon, ray.DirLength, ray.DirXGeq0, ray.DirXLeq0, ray.DirYGeq0, ray.DirYLeq0, ray.DirZGeq0, ray.DirZLeq0, ray.invMagnitude, ray.magnitude, ray.dirMagnitude, ray.InvDir
}

func (me *TRay) UpdateDirMagnitude () {
	me.DirLength = me.Dir.Length()
	me.magnitude = math.Sqrt(me.DirLength)
	me.invMagnitude = 1 / me.magnitude
	me.dirMagnitude.SetFromMult1(&me.Dir, me.invMagnitude)
	me.InvDir.SetFromInv(&me.Dir)
	me.DirXGeq0, me.DirXLeq0, me.DirYGeq0, me.DirYLeq0, me.DirZGeq0, me.DirZLeq0 = me.Dir.X >= 0, me.Dir.X <= 0, me.Dir.Y >= 0, me.Dir.Y <= 0, me.Dir.Z >= 0, me.Dir.Z <= 0
	me.DirEpsilon.SetFromMult1(&me.Dir, num.Epsilon)
	// me.DirXGt0, me.DirXLt0, me.DirYGt0, me.DirYLt0, me.DirZGt0, me.DirZLt0 = me.Dir.X > 0, me.Dir.X < 0, me.Dir.Y > 0, me.Dir.Y < 0, me.Dir.Z > 0, me.Dir.Z < 0
}

var (
	NilColor = TColor {}
)

func ColorToUint8 (val float64) uint8 {
	if val <= 0 { return 0 } else if val >= 1 { return 255 }
	return uint8(255 * val)
}

func ColorToUint8s (r, g, b, a float64, pixMap []uint8, offset int) {
	if r <= 0 { pixMap[offset + 0] = 0 } else if r >= 1 { pixMap[offset + 0] = 255 } else { pixMap[offset + 0] = uint8(255 * r) }
	if g <= 0 { pixMap[offset + 1] = 0 } else if g >= 1 { pixMap[offset + 1] = 255 } else { pixMap[offset + 1] = uint8(255 * g) }
	if b <= 0 { pixMap[offset + 2] = 0 } else if b >= 1 { pixMap[offset + 2] = 255 } else { pixMap[offset + 2] = uint8(255 * b) }
	if a <= 0 { pixMap[offset + 3] = 0 } else if a >= 1 { pixMap[offset + 3] = 255 } else { pixMap[offset + 3] = uint8(255 * a) }
}

func MixAlpha (a1, a2 uint8) uint8 {
	var i1, i2, i3 int = int(a1), int(a2), 0
	if i3 = i1 + i2; i3 > 255 { return 255 } else if i3 < 0 { return 0 }
	return uint8(i3)
}

func MixColorsCheap (r1, r2, a1, a2 float64) float64 {
	return (r1 * a1) + (r2 * a2)
}

func MixRGBA (r1, r2, a1, a2 uint8) uint8 {
	var f1, f2 float64 = float64(r1), float64(r2)
	var fa1, fa2 float64 = float64(a1) / 256, float64(a2) / 256
	var i1, i2, i3 int = int(f1 * fa1), int(f2 * fa2), 0
	if i3 = i1 + i2; i3 > 255 { return 255 } else if i3 < 0 { return 0 }
	return uint8(i3)
}

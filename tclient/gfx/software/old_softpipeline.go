package softpipeline

import (
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"tclient/gfx/voxels"
	num "tshared/numutil"
)

type renderTile struct {
	rayDir, rayPos, rayTmp, rayTmp1, rayTmp2, rayTmp3 num.Dvec3
	minT, t1, t2, t3, t4, t5, t6, numSteps, fx, fy, fracx, fracy, fracz, tmpY float64
	thisAlpha uint8
	voxel color.RGBA
	tmp bool
	rayIntX, rayIntY, rayIntZ, lrayIntX, lrayIntY, lrayIntZ int
	sampledR, sampledG, sampledB uint8
	colx, coly, colz color.RGBA
	srcr, srcg, srcb, srca, dstr, dstg, dstb, dsta, outr, outg, outb, outa float64
	str string
}

func (me *renderTile) makeRayInts () bool {
	if me.rayPos.X < Minrx { Minrx = me.rayPos.X }
	if me.rayPos.Y < Minry { Minry = me.rayPos.Y }
	if me.rayPos.Z < Minrz { Minrz = me.rayPos.Z }
	if me.rayPos.X > Maxrx { Maxrx = me.rayPos.X }
	if me.rayPos.Y > Maxry { Maxry = me.rayPos.Y }
	if me.rayPos.Z > Maxrz { Maxrz = me.rayPos.Z }
	me.rayIntX, me.rayIntY, me.rayIntZ = int(me.rayPos.X), int(me.rayPos.Y), int(me.rayPos.Z)
	return (me.rayIntX >= 0) && (me.rayIntX < SceneSize) && (me.rayIntY >= 0) && (me.rayIntY < SceneSize) && (me.rayIntZ >= 0) && (me.rayIntZ < SceneSize)
}

func mixA (a1, a2 uint8) uint8 {
	var i1, i2, i3 int = int(a1), int(a2), 0
	if i3 = i1 + i2; i3 > 255 { return 255 } else if i3 < 0 { return 0 }
	return uint8(i3)
}

func (me *renderTile) mixCol (src, dst, nu *color.RGBA) {
	me.srcr, me.srcg, me.srcb, me.srca = float64(src.R) / 255, float64(src.G) / 255, float64(src.B) / 255, float64(src.A) / 255
	me.dstr, me.dstg, me.dstb, me.dsta = float64(dst.R) / 255, float64(dst.G) / 255, float64(dst.B) / 255, float64(dst.A) / 255
	if me.outa = me.srca + (me.dsta * (1 - me.srca)); me.outa == 0 {
		me.outr, me.outg, me.outb = 0, 0, 0
	} else {
		me.outr = ((me.srcr * me.srca) + (me.dstr * me.dsta * (1 - me.srca))) / me.outa
		me.outg = ((me.srcg * me.srca) + (me.dstg * me.dsta * (1 - me.srca))) / me.outa
		me.outb = ((me.srcb * me.srca) + (me.dstb * me.dsta * (1 - me.srca))) / me.outa
	}
	nu.R, nu.G, nu.B, nu.A = uint8(me.outr * 255), uint8(me.outg * 255), uint8(me.outb * 255), uint8(me.outa * 255)
}

func mixRGB (r1, r2, a1, a2 uint8) uint8 {
	var f1, f2 float64 = float64(r1), float64(r2)
	var fa1, fa2 float64 = float64(a1) / 256, float64(a2) / 256
	var i1, i2, i3 int = int(f1 * fa1), int(f2 * fa2), 0
	if i3 = i1 + i2; i3 > 255 { return 255 } else if i3 < 0 { return 0 }
	return uint8(i3)
}

func (me *renderTile) getSample (x, y, z int, col *color.RGBA) {
	if (x >= SceneSize) || (y >= SceneSize) || (z >= SceneSize) {
		col.R, col.G, col.B, col.A = 0, 0, 0, 0
	} else if colorCube {
		if (num.IsEveni(x) && num.IsEveni(y)) || ((z > 0) && num.IsEveni(z)) {
			col.R, col.G, col.B, col.A = uint8(x / 2), uint8(y / 2), uint8(z / 2), 255
		} else {
			col.R, col.G, col.B, col.A = uint8(x), uint8(y), uint8(z), 255
		}
	} else {
		me.voxel = Scene[x][y][z]
		if me.voxel.A >= 1 {
			if correctAlpha {
				me.mixCol(col, &me.voxel, col)
			} else {
				col.R = mixRGB(col.R, me.voxel.R, col.A, me.voxel.A) // num.Mixb(me.voxel.R, me.voxel.A, col.R)
				col.G = mixRGB(col.G, me.voxel.G, col.A, me.voxel.A) // num.Mixb(me.voxel.G, me.voxel.A, col.G)
				col.B = mixRGB(col.B, me.voxel.B, col.A, me.voxel.A) // num.Mixb(me.voxel.B, me.voxel.A, col.B)
				col.A = mixA(col.A, me.voxel.A) // num.Mixb(1, col.A, me.voxel.A)
			}
		}
	}
}

func (me *renderTile) CastRay (x, y int, col *color.RGBA) {
	me.fx, me.fy = float64(x), float64(y)
	col.R, col.G, col.B, col.A = 0, 0, 0, 0
	me.rayPos.X, me.rayPos.Y, me.rayPos.Z = ((me.fx / width) - 0.5), ((me.fy / height) - 0.5), 0
	me.rayPos.MultMat(cmat1)
	me.rayTmp.X, me.rayTmp.Y, me.rayTmp.Z = 0, 0, planeDist
	me.rayTmp.MultMat(cmat1)
	// if (CamRot.Y != 0) && ((x == 0) || (x == 39) || (x == 79) || (x == 119) || (x == 159)) && ((y == 0) || (y == 44) || (y == 89)) { log.Printf("[%v,%v] 0,0,%v ==> %+v (for %+v)", x, y, planeDist, me.rayTmp, me.rayDir) }
	me.rayDir.X, me.rayDir.Y, me.rayDir.Z = me.rayPos.X, me.rayPos.Y, me.rayPos.Z - me.rayTmp.Z
	me.rayDir.Normalize()
	me.rayDir.MultMat(pmat)
	me.rayPos.Add(CamPos)

	// me.rayDir.X, me.rayDir.Y, me.rayDir.Z = -((me.fx / width) - 0.5), -((me.fy / height) - 0.5), planeDist

	if true {
		// if ((x == 0) || (x == 159)) && ((y == 0) || (y == 89)) { log.Printf("RAYPOS[%v,%v]=%+v", x, y, me.rayPos) }
		me.numSteps = 0
		for (col.A < 255) && me.rayPos.AllInRange(vmin, vmax) && (me.numSteps <= (vboth)) {
			me.numSteps++
			if me.rayPos.AllInRange(0, fs) {
				if col.A == 0 {
					if (int(me.rayPos.X) == 0) || (int(me.rayPos.X) == (SceneSize - 1)) {
						col.R, col.G, col.B, col.A = 0, 0, 64, 64
					} else if (int(me.rayPos.Y) == 0) || (int(me.rayPos.Y) == (SceneSize - 1)) {
						col.R, col.G, col.B, col.A = 0, 64, 0, 64
					} else {
						col.R, col.G, col.B, col.A = 32, 32, 32, 48
					}
				}
				if samples {
					_, me.fracx = math.Modf(me.rayPos.X); _, me.fracy = math.Modf(me.rayPos.Y); _, me.fracz = math.Modf(me.rayPos.Z)
					me.rayIntX, me.rayIntY, me.rayIntZ = int(me.rayPos.X), int(me.rayPos.Y), int(me.rayPos.Z)
					me.lrayIntX, me.lrayIntY, me.lrayIntZ = int(me.rayPos.X + 1), int(me.rayPos.Y + 1), int(me.rayPos.Z + 1)
					me.getSample(me.rayIntX, me.rayIntY, me.rayIntZ, col)
					me.getSample(me.lrayIntX, me.rayIntY, me.rayIntZ, &me.colx)
					me.getSample(me.rayIntX, me.lrayIntY, me.rayIntZ, &me.coly)
					me.getSample(me.rayIntX, me.rayIntY, me.lrayIntZ, &me.colz)
					//me.mixAll()
				} else {
					me.rayIntX, me.rayIntY, me.rayIntZ = int(me.rayPos.X), int(me.rayPos.Y), int(me.rayPos.Z)
					me.getSample(me.rayIntX, me.rayIntY, me.rayIntZ, col)
				}
			}
			me.rayPos.Add(me.rayDir)
		}
	}
}

var (
	// yaw=leftright=rotY pitch=updown=rotX
	AspectRatio float64 = 0
	FieldOfView float64 = 60
	Chan = make(chan int)
	CamPos, CamRot, CamRad, CamLookAt, CamUp, CamAxisX, CamAxisY, CamAxisZ num.Dvec3
	CamTurnLeft, CamTurnRight, CamMoveBack, CamMoveFwd, CamMoveLeft, CamMoveRight, CamMoveUp, CamMoveDown float64
	Width, Height int
	Scene [][][]color.RGBA
	SceneSize = 64
	NumTiles = 16
)

var (
	// 	me.rayAxisY.X, me.rayAxisY.Y, me.rayAxisY.Z = 0, 1, 0 // CamPos.X, me.rayPos.Y, me.rayPos.Z
	axisX, axisY, axisZ = &num.Dvec3 { 1, 0, 0 }, &num.Dvec3 { 0, 1, 0 }, &num.Dvec3 { 0, 0, 1 }
	axisXm, axisYm, axisZm = &num.Dvec3 { -1, 0, 0 }, &num.Dvec3 { 0, -1, 0 }, &num.Dvec3 { 0, 0, -1 }
	plane00, plane10, plane11, plane01, origin, dx, dy num.Dvec3
	vmin, vmax, vboth, fs, planeDist, width, height, deltaTime float64
	pmat, bmat, cmat1, cmat2, rmat num.Dmat4
	tileWidth, tileHeight int
	lastTick int64
	colorCube bool = false
	correctAlpha = false
	renderTiles = []*renderTile {}
	samples = false
	Maxrx, Maxry, Maxrz, Minrx, Minry, Minrz float64
	yes = 67.6125
	tmpFloats = make([]float64, 20)
	updateCam = false
)

func PostRender () {
}

func PreRender (nowTick int64) {
	deltaTime = float64(60 * (nowTick - lastTick)) / (1000000000)
	lastTick = nowTick
	if (CamTurnLeft != 0) || (CamTurnRight != 0) {
		CamRot.Y = math.Mod(CamRot.Y + (float64(CamTurnRight - CamTurnLeft) / 1), 360)
		updateCam = true
		// rmat = num.NewDmat4Identity().Mult(num.NewDmat4RotationZ(num.DegToRad(CamRot.Y)))
	}
	if updateCam || (CamMoveFwd != 0) || (CamMoveBack != 0) || (CamMoveLeft != 0) || (CamMoveRight != 0) || (CamMoveUp != 0) || (CamMoveDown != 0) {
		CamPos.X += deltaTime * (((CamMoveFwd - CamMoveBack) * math.Sin(CamRad.Y) * math.Abs(math.Cos(CamRad.X)) * 0.5) - ((CamMoveLeft - CamMoveRight) * math.Cos(CamRad.Y) * 0.25))
		CamPos.Z += deltaTime * (((CamMoveFwd - CamMoveBack) * math.Cos(CamRad.Y) * math.Abs(math.Cos(CamRad.X)) * 0.5) + ((CamMoveLeft - CamMoveRight) * math.Sin(CamRad.Y) * 0.25))
		CamPos.Y += deltaTime * ((-(CamMoveFwd - CamMoveBack) * 0.5 * math.Sin(CamRad.X)) + ((CamMoveUp - CamMoveDown) * math.Cos(CamRad.X) * 0.5))
		updateCam = true
	}
	if updateCam { UpdateCam() }
}

func Reinit (w, h int) {
	var no = 60.0
	fs = float64(SceneSize)
	Width, Height = w, h
	width, height = float64(w), float64(h)
	AspectRatio = width / height
	vmin = -(fs * 8)
	vmax = fs * 8
	vboth = vmax + math.Abs(vmin)
	planeDist = 1 / math.Tan(no * 0.5)
	tileWidth, tileHeight = int(math.Ceil(width / float64(NumTiles))), int(math.Ceil(height / float64(NumTiles)))
	pmat = num.NewDmat4Identity().Mult(num.NewDmat4Perspective(yes, height / width, 0.1, 500))
	if Scene == nil {
		CamUp.X, CamUp.Y, CamUp.Z = 0, 1, 0
		CamPos.X, CamPos.Y, CamPos.Z = fs / 2, (fs / 4) * 2, -(fs * 4) // 96, 32, -159 // 
		UpdateCam()
		log.Printf("Loading volume...")
		Scene = voxels.LoadVolume("bunny.raw", SceneSize)
	}
	lastTick = time.Now().UnixNano()
}

func Render (w, h int, srt *image.RGBA) {
	var cc, wait = 0, 0
	var tx, ty, tw, th int
	var rt *renderTile
	tw, th = tileWidth, tileHeight
	for ty = 0; ty < NumTiles; ty++ {
		for tx = 0; tx < NumTiles; tx++ {
			if len(renderTiles) <= wait {
				rt = &renderTile {}
				// rt.rmat = make([][]float64, 3)
				// rt.rmat[0] = make([]float64, 3)
				// rt.rmat[1] = make([]float64, 3)
				// rt.rmat[2] = make([]float64, 3)
				renderTiles = append(renderTiles, rt)
			}
			go RenderTile(tx * tw, num.Mini((tx + 1) * tw, w), ty * th, num.Mini((ty + 1) * th, h), renderTiles[wait], srt)
			wait++
		}
	}
	for (cc < wait) { cc += (<- Chan) }
}

func RenderTile (xmin, xmax, ymin, ymax int, rt *renderTile, srt *image.RGBA) {
	var x, y, poff int
	var col color.RGBA
	for y = ymin; y < ymax; y++ {
		for x = xmin; x < xmax; x++ {
			rt.CastRay(x, y, &col)
			poff = srt.PixOffset(x, y)
			srt.Pix[poff + 0] = col.R
			srt.Pix[poff + 1] = col.G
			srt.Pix[poff + 2] = col.B
			srt.Pix[poff + 3] = col.A
		}
	}
	Chan <- 1
}

func UpdateCam () {
	updateCam = false
	CamRad.X, CamRad.Y, CamRad.Z = num.DegToRad(CamRot.X), num.DegToRad(CamRot.Y), num.DegToRad(CamRot.Z)
	rmat = num.NewDmat4RotationY(CamRad.Y)
	CamLookAt.X, CamLookAt.Y, CamLookAt.Z = 0, 0, 1
	CamLookAt.MultMat(rmat)
	CamLookAt.Add(CamPos)
	CamAxisZ = CamLookAt.Sub(CamPos)
	CamAxisZ.Normalize()
	CamAxisX = CamUp.Cross(CamAxisZ)
	CamAxisY = CamAxisX.Cross(CamAxisZ.SwapSign())
	cmat1 = num.NewDmat4LookAt2(CamPos, CamAxisX, CamAxisY, CamAxisZ, tmpFloats)
	// origin.X, origin.Y, origin.Z = 0, 0, planeDist
	// origin.MultMat2(cmat1)
	// plane00.X, plane00.Y, plane00.Z = -0.5, 0.5, 0
	// plane00.MultMat2(cmat1)
	// plane10.X, plane10.Y, plane10.Z = 0.5, 0.5, 0
	// plane10.MultMat2(cmat1)
	// plane11.X, plane11.Y, plane11.Z = 0.5, -0.5, 0
	// plane11.MultMat2(cmat1)
	// plane01.X, plane01.Y, plane01.Z = -0.5, -0.5, 0
	// plane01.MultMat2(cmat1)
	// dx = plane10.Sub(plane00).Mult1(1 / width)
	// dy = plane01.Sub(plane00).Mult1(1 / height)
}

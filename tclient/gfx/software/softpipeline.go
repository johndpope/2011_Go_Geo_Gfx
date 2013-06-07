package softpipeline

import (
	"image"
	"log"
	"math"
	"runtime"

	gfx "tclient/gfx/gfxutil"
	"tclient/gfx/voxels"
	num "tshared/numutil"
)

type TScene struct {
	CamPos, CamRot, CamRad, CamSin, CamCos num.Vec3
	Range, FieldOfView float64
	Fog bool
	ColAmbient gfx.TColor
	Objects []*gfx.TObject
	Lights []*gfx.TLight
}

func (me *TScene) UpdateCamRot () {
	me.CamRad.SetFromDegToRad(&me.CamRot)
	me.CamCos.SetFromCos(&me.CamRad)
	me.CamSin.SetFromSin(&me.CamRad)
}

func NewScene () *TScene {
	var me = &TScene {}
	me.Objects = []*gfx.TObject {}
	me.Lights = []*gfx.TLight {}
	me.Range = 8192
	return me
}

type TThread struct {
	picker *gfx.TPicker
	inVol, isDebug bool
	lightCol, tempCol, volCol gfx.TColor
	cols, fcols []gfx.TColor
	fweights []float64
	nodeIndices, nodeIndices2 []int
	xmin, xmax, ymin, ymax, rx, ry, rpo, rec, vx, vy, vz, fxfloor, fyfloor, fzfloor, fxceil, fyceil, fzceil, step, maxSteps int
	lum, fx, fy, srrf, tmp, fxfrac, fyfrac, fzfrac, boxDist, snrLastDist, snrMaxLevel float64
	srpd, srpr, srd, srld []float64
	prays, srays, lrays []*gfx.TRay
	traverser *voxels.TOctreeTraverser
	ray, oldRay gfx.TRay
	snray *gfx.TRay
	vec, nodeExit num.Vec3
	light *gfx.TLight
	debug string
	numIndices, numIndices2, fori, forj int
	voxelFilter *voxels.TVoxelFilter
}

func (me *TThread) Light () {
	me.lightCol = Scene.ColAmbient
	me.lrays[me.rec].Pos = me.prays[me.rec].Pos
	for _, me.light = range Scene.Lights {
		me.lrays[me.rec].Dir.SetFromSub(&me.light.RotPos, &me.prays[me.rec].Pos)
		me.lrays[me.rec].UpdateDirMagnitude()
		if me.picker.PickCached(me.lrays[me.rec]); me.picker.IsPick {
			me.srld[me.rec] = (me.lrays[me.rec].Dir.Dot(&me.prays[me.rec].Dir) / me.lrays[me.rec].DirLength) * me.light.Intensity
			if me.srld[me.rec] > 0 {
				me.lightCol.R += (me.light.Col.R * me.srld[me.rec])
				me.lightCol.G += (me.light.Col.G * me.srld[me.rec])
				me.lightCol.B += (me.light.Col.B * me.srld[me.rec])
			}
		}
	}
}

func (me *TThread) RenderPixel () {
	me.fx, me.fy = float64(me.rx) + 0.5, float64(me.ry) + 0.5
	me.isDebug = (me.rx == 160) && (me.ry == 90) && (OctreeMaxLevel > 2)
	me.ray.Pos.X, me.ray.Pos.Y, me.ray.Pos.Z = (me.fx * planeStepWidth) - planeWidthHalf, (me.fy * planeStepHeight) - planeHeightHalf, planePosZ
	me.ray.Dir.SetFromRotation(me.ray.Pos, Scene.CamCos, Scene.CamSin)
	me.ray.Pos.X, me.ray.Pos.Y, me.ray.Pos.Z = Scene.CamPos.X, Scene.CamPos.Y, Scene.CamPos.Z
	me.rec = 0
	if RootOctree != nil { me.ray.Dir.Normalize() }
	me.ray.UpdateDirMagnitude()
	if RootOctree != nil {
		me.cols[0].Reset()
		me.traverser.Ray = &me.ray
		me.traverser.Traverse()
		me.cols[0].SetFrom(&me.traverser.Color)
	} else {
		me.SendRay(&me.ray)
	}
	if me.cols[0].A == 0 {
		me.cols[0].R, me.cols[0].G, me.cols[0].B, me.cols[0].A = 0.4, 0.4, 0.4, 1
	} else if me.cols[0].A < 1 {
		me.tempCol.R, me.tempCol.G, me.tempCol.B, me.tempCol.A = 0.4, 0.4, 0.4, 1
		me.volCol.Mix(&me.cols[0], &me.tempCol)
		me.cols[0] = me.volCol
	}
}

func (me *TThread) RenderTile () {
	me.traverser.MaxLevel = OctreeMaxLevel
	for me.ry = me.ymin; me.ry < me.ymax; me.ry++ {
		for me.rx = me.xmin; me.rx < me.xmax; me.rx++ {
			me.RenderPixel()
			gfx.ColorToUint8s(me.cols[0].R, me.cols[0].G, me.cols[0].B, me.cols[0].A, renderTarget.Pix, renderTarget.PixOffset(me.rx, me.ry))
		}
	}
	threadChan <- 1
}

// func (me *TThread) Sample (ray *gfx.TRay) {
// 	if Filtered && (me.voxelFilter.Vol != nil) {
// 		me.voxelFilter.Linear2(&ray.Pos)
// 		me.tempCol = me.voxelFilter.OutCol
// 	} else if me.octNode.Brick != nil {
// 		me.vx, me.vy, me.vz = int(ray.Pos.X * me.octNode.Brick.InvScale), int(ray.Pos.Y * me.octNode.Brick.InvScale), int(ray.Pos.Z * me.octNode.Brick.InvScale)
// 		if me.vx >= 8 || me.vy >= 8 || me.vz >= 8 || me.vx < 0 || me.vy < 0 || me.vz < 0 {
// 			log.Panicf("SAMPLE %v,%v,%v for (%v) * %v,%v,%v", me.vx, me.vy, me.vz, me.octNode.Brick.InvScale, ray.Pos.X, ray.Pos.Y, ray.Pos.Z)
// 		} else {
// 			me.tempCol = me.octNode.Brick.Col[me.vx][me.vy][me.vz]
// 		}
// 	} else {
// 		me.tempCol = gfx.TColor { 1, 1, 0, 1 } // me.octNode.Col
// 	}
// }

func (me *TThread) SendRay (ray *gfx.TRay) {
	me.cols[me.rec].Reset()
	for (me.cols[me.rec].A == 0) {
		if me.picker.Pick(ray); !me.picker.IsPick { break }
		me.cols[me.rec].R += me.picker.Col.R
		me.cols[me.rec].G += me.picker.Col.G
		me.cols[me.rec].B += me.picker.Col.B
		me.cols[me.rec].A, me.srpr[me.rec], me.srpd[me.rec] = 1, me.picker.Refl, me.picker.Dist
		me.prays[me.rec].Pos, me.srays[me.rec].Pos, me.prays[me.rec].Dir = me.picker.Entry, me.picker.Entry, me.picker.Dir
		if (me.rec < maxRec) && (me.srpr[me.rec] > 0) {
			me.srd[me.rec] = 2 * me.prays[me.rec].Dir.Dot(&ray.Dir)
			me.srays[me.rec].Dir.SetFromSubMult1(&ray.Dir, &me.prays[me.rec].Dir, me.srd[me.rec])
			me.srays[me.rec].UpdateDirMagnitude()
			me.rec++
			me.SendRay(me.srays[me.rec - 1])
			me.rec--
			me.cols[me.rec].R += (me.cols[me.rec + 1].R * me.srpr[me.rec])
			me.cols[me.rec].G += (me.cols[me.rec + 1].G * me.srpr[me.rec])
			me.cols[me.rec].B += (me.cols[me.rec + 1].B * me.srpr[me.rec])
		}
		if len(Scene.Lights) > 0 {
			me.Light()
			me.cols[me.rec].R *= me.lightCol.R
			me.cols[me.rec].G *= me.lightCol.G
			me.cols[me.rec].B *= me.lightCol.B
		}
		ray.Pos.SetFromAdd(&me.prays[me.rec].Pos, &ray.Dir)
		ray.Pos.Add(&ray.Dir)
	}
	if Scene.Fog && (me.rec == 0) && (me.cols[me.rec].A > 0) {
		me.srrf = (Scene.Range - me.srpd[me.rec]) / Scene.Range
		me.tmp = 1 - me.srrf
		me.cols[me.rec].R = (me.cols[me.rec].R * me.srrf) + (0.9 * me.tmp)
		me.cols[me.rec].G = (me.cols[me.rec].G * me.srrf) + (0.6 * me.tmp)
		me.cols[me.rec].B = (me.cols[me.rec].B * me.srrf) + (0.3 * me.tmp)
	}
}

var (
	Scene *TScene
	RootOctree, CurOctree *voxels.TOctree
	OctreeMaxLevel, Filtered, MaxSteps, NumThreads, Speed, VolSize = 0, false, 0, 6, 1.0, 512.0
	Width, Height, HeightWidthRatio, WidthHeightRatio, CamTurnLeft, CamTurnRight, CamTurnUp, CamTurnDown, CamMoveBack, CamMoveFwd, CamMoveLeft, CamMoveRight, CamMoveUp, CamMoveDown float64
	NowTick int64

	threads []*TThread
	curThread *TThread
	width, height, tileWidth, tileHeight, rtx, rty, rcc, rwait int
	renderTarget *image.RGBA
	deltaTime, sceneZoom float64
	lastTick int64
	animLights, updateCam, maxRec, threadChan, volSize = true, true, 3, make(chan int), int(VolSize)
	planeWidth, planeHeight, planePosZ, planeStepX, planeStepY, planeStepWidth, planeStepHeight, planeRange = 0.1, 0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	planeWidthHalf, planeHeightHalf = planeWidth / 2, planeHeight / 2
)

func PostRender () {
}

func PreRender () {
	deltaTime = float64(60 * (NowTick - lastTick)) / (1000000000 * Speed)
	if animLights {
		Scene.Lights[0].SetRotation(0, math.Mod(Scene.Lights[0].Rot.Y + (deltaTime), 360))
		Scene.Lights[1].SetRotation(0, math.Mod(Scene.Lights[1].Rot.Y + (deltaTime), 360))
		// Scene.Lights[2].SetRotation(0, math.Mod(Scene.Lights[2].Rot.Y + (deltaTime), 360))
		// Scene.Lights[3].SetRotation(0, math.Mod(Scene.Lights[3].Rot.Y + (deltaTime), 360))
	}
	if (CamTurnLeft != 0) || (CamTurnRight != 0) || (CamTurnUp != 0) || (CamTurnDown != 0) {
		Scene.CamRot.Y = math.Mod(Scene.CamRot.Y + (float64(CamTurnRight - CamTurnLeft) * deltaTime * 3), 360)
		Scene.CamRot.X = math.Mod(Scene.CamRot.X + (float64(CamTurnDown - CamTurnUp) * deltaTime * 3), 360)
		Scene.UpdateCamRot()
		updateCam = true
	}
	if updateCam || (CamMoveFwd != 0) || (CamMoveBack != 0) || (CamMoveLeft != 0) || (CamMoveRight != 0) || (CamMoveUp != 0) || (CamMoveDown != 0) {
		Scene.CamPos.X += deltaTime * 12 * (((CamMoveFwd - CamMoveBack) * Scene.CamSin.Y * math.Abs(Scene.CamCos.X) * 0.5) - ((CamMoveLeft - CamMoveRight) * Scene.CamCos.Y * 0.25))
		Scene.CamPos.Z += deltaTime * 12 * (((CamMoveFwd - CamMoveBack) * Scene.CamCos.Y * math.Abs(Scene.CamCos.X) * 0.5) + ((CamMoveLeft - CamMoveRight) * Scene.CamSin.Y * 0.25))
		Scene.CamPos.Y += deltaTime * 12 * ((-(CamMoveFwd - CamMoveBack) * 0.5 * Scene.CamSin.X) - ((CamMoveDown - CamMoveUp) * Scene.CamCos.X * 0.5))
		updateCam = false
	}
	lastTick = NowTick
}

func Reinit (w, h int, target *image.RGBA) {
	var rt *TThread
	width, height = w, h
	Width, Height = float64(w), float64(h)
	WidthHeightRatio = Width / Height
	planeHeight = planeWidth / WidthHeightRatio
	planeHeightHalf = planeHeight / 2
	planeStepX, planeStepY = 1 / Width, 1 / Height
	planeStepWidth, planeStepHeight = planeStepX * planeWidth, planeStepY * planeHeight
	renderTarget, HeightWidthRatio, tileWidth, tileHeight = target, Height / Width, int(math.Ceil(Width / float64(NumThreads))), int(math.Ceil(Height / float64(NumThreads)))
	if Scene == nil {
		log.Printf("Loading scene...")
		RootOctree = voxels.NewOctree(voxels.NewVolume("/ssd2/ScanModels/dragon512.raw", volSize), 8); RootOctree.Print(0)
		MaxSteps = int(VolSize * 1.5)
		Scene = NewScene()
		Scene.CamPos.X, Scene.CamPos.Y, Scene.CamPos.Z = 21, 39, 35 // VolSize / 2, VolSize / 2, -(VolSize * 2)
		Scene.ColAmbient = gfx.TColor { 0.66, 0.66, 0.66, 1 }
		Scene.Fog = true
		Scene.Objects = []*gfx.TObject {
			gfx.NewSphere(gfx.TColor { 0.33, 0.5, 0.33, 1 }, num.Vec3 { 0, -8198, 0 }, 8192, 0.33),
			gfx.NewSphere(gfx.TColor { 0.66, 0.66, 0, 1 }, num.Vec3 { -8, 0, 0 }, 2, 0.25),
			gfx.NewSphere(gfx.TColor { 0, 0, 0.66, 1 }, num.Vec3 { 2.5, -5, 0 }, 1.5, 0.75),
			gfx.NewSphere(gfx.TColor { 0.44, 0.22, 0, 1 }, num.Vec3 { 0, 0, 16 }, 2.5, 1),
			gfx.NewSphere(gfx.TColor { 0, 0, 0.33, 1 }, num.Vec3 { 0, 0, -18 }, 2, 0.75),
			gfx.NewSphere(gfx.TColor { 0.33, 0, 0, 1 }, num.Vec3 { 2.5, 5, 0 }, 1.5, 0.25),
			gfx.NewBox(gfx.TColor { 0.44, 0.22, 0, 1 }, num.Vec3 { 0, 0, 0 }, num.Vec3 { 4, 4, 4 }, 0.5),
			gfx.NewBox(gfx.TColor { 0.05, 0.05, 0.05, 1 }, num.Vec3 { -256, -4, 4 }, num.Vec3 { 512, 1, 8 }, 0.5),
		}
		Scene.Lights = []*gfx.TLight {
			gfx.NewLight(num.Vec3 { -200, 200, 200 }, num.Vec3 { 20.1, 20.1, 20.1 }, gfx.TColor { 0.8, 0.8, 0.8, 1 }, 256),
			// NewLight(num.Vec3 { 20, 20, 20 }, num.Vec3 { 0.4, 0.4, 0.4 }, gfx.TColor { 0.6, 0.6, 0.6, 1 }, 20),
			// NewLight(num.Vec3 { 10, 30, -10 }, num.Vec3 { 0.7,  0.7, 0.7 }, gfx.TColor { 0.4, 0.4, 0.4, 1 }, 20),
			gfx.NewLight(num.Vec3 { 600, 600, -600 }, num.Vec3 { 81, 81, 81 }, gfx.TColor { 0.9, 0.9, 0.9, 1 }, 512),
		}
		Scene.UpdateCamRot()
		threads = make([]*TThread, NumThreads * NumThreads)
		for t := 0; t < len(threads); t++ {
			rt = &TThread {}
			rt.picker = gfx.NewPicker(Scene.Objects, Scene.Range)
			rt.cols, rt.srpd, rt.srpr, rt.srd, rt.srld = make([]gfx.TColor, maxRec + 2), make([]float64, maxRec + 2), make([]float64, maxRec + 2), make([]float64, maxRec + 2), make([]float64, maxRec + 2)
			rt.srays, rt.prays, rt.lrays = make([]*gfx.TRay, maxRec + 2), make([]*gfx.TRay, maxRec + 2), make([]*gfx.TRay, maxRec + 2)
			rt.fcols = make([]gfx.TColor, 8)
			rt.fweights = make([]float64, 8)
			rt.voxelFilter = voxels.NewVoxelFilter(nil)
			rt.voxelFilter.SubVolSize = 8
			rt.numIndices, rt.numIndices2 = 8, 8
			rt.nodeIndices, rt.nodeIndices2 = make([]int, rt.numIndices), make([]int, rt.numIndices)
			for i := 0; i < rt.numIndices; i++ { rt.nodeIndices[i], rt.nodeIndices2[i] = i, i }
			if RootOctree != nil {
				rt.snrMaxLevel = float64(RootOctree.MaxLevel)
				rt.traverser = voxels.NewOctreeTraverser(RootOctree, Scene.Lights)
			}
			for i := 0; i < len(rt.srays); i++ { rt.prays[i], rt.srays[i], rt.lrays[i] = &gfx.TRay {}, &gfx.TRay {}, &gfx.TRay {} }
			threads[t] = rt
		}
	}
	Scene.FieldOfView = 33.75 * math.Min(2, WidthHeightRatio) //   math.Max(1, WidthHeightRatio))
	sceneZoom = 1 / math.Tan(num.DegToRad(Scene.FieldOfView) * 0.5)
	planePosZ = planeWidthHalf * sceneZoom
	planeRange = planeWidth * Scene.Range
	runtime.GC()
}

func Render () {
	rcc, rwait = 0, 0
	if RootOctree != nil { if OctreeMaxLevel > RootOctree.MaxLevel { OctreeMaxLevel = RootOctree.MaxLevel } else if OctreeMaxLevel < 0 { OctreeMaxLevel = 0 } }
	for rty = 0; rty < NumThreads; rty++ {
		for rtx = 0; rtx < NumThreads; rtx++ {
			curThread = threads[rwait]
			curThread.maxSteps, curThread.xmin, curThread.xmax, curThread.ymin, curThread.ymax = MaxSteps, rtx * tileWidth, num.Mini((rtx + 1) * tileWidth, width), rty * tileHeight, num.Mini((rty + 1) * tileHeight, height)
			go curThread.RenderTile()
			rwait++
		}
	}
	for (rcc < rwait) { rcc += (<- threadChan) }
}

package main

import (
	"runtime"

	"github.com/paul-lalonde/Go-SDL/sdl"

	gfxengine "tclient/gfx/engine"
)

func onExit () {
	runtime.UnlockOSThread()
	if gfxengine.HasInitGl {
		gfxengine.CleanUp()
	}
	if gfxengine.HasInitSdl {
		sdl.Quit()
	}
}

func onKeyDown (evt *sdl.KeyboardEvent) {
}

func onKeyUp (evt *sdl.KeyboardEvent) {
}

// func testBox () {
// 	var box = gfx.NewBox(gfx.TColor {}, num.Vec3 { 0, 0, 0 }, num.Vec3 { 4, 4, 4 }, 0)
// 	var picker = gfx.NewPicker([]*gfx.TObject {}, 9999999)
// 	var ray = gfx.TRay {}
// 	var zero float64 = 0.125
// 	var testBoxDir = func (dx, dy, dz float64) {
// 		ray.Dir = num.Vec3 { dx, dy, dz }
// 		ray.UpdateDirMagnitude()
// 		picker.Ray = &ray
// 		if box.Pick(picker); !picker.IsPick {
// 			log.Printf("POS(%+v) DIR(%+v) NOPICK", ray.Pos, ray.Dir)
// 		} else {
// 			log.Printf("POS(%+v) DIR(%+v) PICK entry=%+v DIST=%v", ray.Pos, ray.Dir, picker.Entry, picker.Dist)
// 		}
// 	}
// 	ray.Pos = num.Vec3 { 2, 2, -8 }
// 	testBoxDir(zero, zero, 1)
// 	ray.Pos = num.Vec3 { 2, 2, 8 }
// 	testBoxDir(zero, zero, -1)
// 	ray.Pos = num.Vec3 { -6, 2, 2 }
// 	testBoxDir(1, zero, zero)
// 	ray.Pos = num.Vec3 { 6, 2, 2 }
// 	testBoxDir(-1, zero, zero)
// 	ray.Pos = num.Vec3 { 2, -7, 2 }
// 	testBoxDir(zero, 1, zero)
// 	ray.Pos = num.Vec3 { 2, 7, 2 }
// 	testBoxDir(zero, -1, zero)
// 	log.Printf("BORDERS....")
// 	ray.Pos = num.Vec3 { 2, 2, zero }
// 	testBoxDir(zero, zero, 1)
// 	ray.Pos = num.Vec3 { 2, 2, 4 }
// 	testBoxDir(zero, zero, -1)
// 	ray.Pos = num.Vec3 { zero, 2, 2 }
// 	testBoxDir(1, zero, zero)
// 	ray.Pos = num.Vec3 { 4, 2, 2 }
// 	testBoxDir(-1, zero, zero)
// 	ray.Pos = num.Vec3 { 2, zero, 2 }
// 	testBoxDir(zero, 1, zero)
// 	ray.Pos = num.Vec3 { 2, 4, 2 }
// 	testBoxDir(zero, -1, zero)
// 	ray.Pos = num.Vec3 { 2, 4, 2 }
// 	testBoxDir(zero, 1, zero)
// 	ray.Pos = num.Vec3 { 2, 4, 2 }
// 	testBoxDir(zero, 1, zero)
// 	log.Printf("INSIDE....")
// 	ray.Pos = num.Vec3 { 2, 2, 2 }
// 	testBoxDir(zero, zero, 1)
// 	ray.Pos = num.Vec3 { 2, 2, 2 }
// 	testBoxDir(zero, zero, -1)
// }

// func testRot () {
// 	type tline struct { x1, z1, x2, z2 float64 }
// 	log.Printf("SIN(90)=%v", math.Sin(num.DegToRad(90)))
// 	log.Printf("COS(45)=%v", math.Cos(num.DegToRad(45)))
// 	log.Printf("SIN(45)=%v", math.Sin(num.DegToRad(45)))
// 	log.Printf("COS(45)=%v", math.Cos(45))
// 	log.Printf("SIN(45)=%v", math.Sin(45))
// 	var lines = []tline {
// 		tline { -10, 0, 10, 0 },
// 		tline { -10, 2, 10, 2 },
// 		tline { 0, 0, 16, 0 },
// 		tline { -8, 12, -8, -12 },
// 	}
// 	var rline, ln tline
// 	var roty, rotr float64 = 0, 0
// 	for roty = 0; roty < 360; roty += 45 {
// 		rotr = num.DegToRad(roty)
// 		for l := 0; l < len(lines); l++ {
// 			ln = lines[l]
// 			rline.x1 = (ln.z1 * math.Sin(rotr)) + (ln.x1 * math.Cos(rotr))
// 			rline.z1 = (ln.z1 * math.Cos(rotr)) - (ln.x1 * math.Sin(rotr))
// 			rline.x2 = (ln.z2 * math.Sin(rotr)) + (ln.x2 * math.Cos(rotr))
// 			rline.z2 = (ln.z2 * math.Cos(rotr)) - (ln.x2 * math.Sin(rotr))
// 			log.Printf("LINE[%+v] ROTY%v = [%+v]", lines[l], roty, rline)
// 		}
// 	}
// 	// me.rayDir.X -= ((math.Sin(CamRot.Y) * math.Abs(math.Cos(CamRot.X)) * 0.5) - (math.Cos(CamRot.Y) * 0.25))
// 	// me.rayDir.Z -= ((math.Cos(CamRot.Y) * math.Abs(math.Cos(CamRot.X)) * 0.5) + (math.Sin(CamRot.Y) * 0.25))

// }

// func testRecursion () {
// 	var root = voxels.NewOctree(voxels.NewVolume("bunny.raw", 64), 8)
// 	var rl, ml = 0, root.MaxLevel - 1
// 	var node = make([]*voxels.TOctree, ml + 1)
// 	var walkNodeRec, walkNodeIt func ();
// 	walkNodeIt = func () {
// 		var queue = []*voxels.TOctree { node[0] }
// 		var cur *voxels.TOctree
// 		var processed = map[*voxels.TOctree]bool {}
// 		var l int
// 		for {
// 			if l = len(queue); l < 1 {
// 				break
// 			}
// 			cur = queue[l - 1]
// 			queue = queue[:l - 1]
// 			if (cur.Level == ml) || (cur.ChildNodes == nil) {
// 				log.Printf("EXIT LEVEL %v", cur.Level);
// 			} else if !processed[cur] {
// 				log.Printf("ENTER LEVEL %v", cur.Level)
// 				log.Printf("PRESTUFF %v", cur.Level)
// 				for i := 0; i < 3; i++ {
// 					switch i {
// 					case 0:
// 						log.Printf("PRECASE %v.%v", cur.Level, i)
// 						queue = append(queue, cur.ChildNodes[cur.Level + i])
// 						log.Printf("POSTCASE %v.%v", cur.Level, i)
// 					case 1:
// 						log.Printf("PRECASE %v.%v", cur.Level, i)
// 						queue = append(queue, cur.ChildNodes[cur.Level + i])
// 						log.Printf("POSTCASE %v.%v", cur.Level, i)
// 					case 2:
// 						log.Printf("PRECASE %v.%v", cur.Level, i)
// 						queue = append(queue, cur.ChildNodes[cur.Level + i])
// 						log.Printf("POSTCASE %v.%v", cur.Level, i)
// 					}
// 				}
// 			}
// 			processed[cur] = true
// 		}
// 	}
// 	walkNodeRec = func () {
// 		log.Printf("ENTER LEVEL %v", rl)
// 		if (node[rl].Level == ml) || (node[rl].ChildNodes == nil) { log.Printf("EXIT LEVEL %v", rl); return }
// 		log.Printf("PRESTUFF %v", rl)
// 		for i := 0; i < 3; i++ {
// 			switch i {
// 			case 0:
// 				log.Printf("PRECASE %v.%v", rl, i)
// 				node[rl + 1] = node[rl].ChildNodes[rl + i]; rl++; walkNodeRec(); rl--
// 				log.Printf("POSTCASE %v.%v", rl,  i)
// 			case 1:
// 				log.Printf("PRECASE %v.%v", rl, i)
// 				node[rl + 1] = node[rl].ChildNodes[rl + i]; rl++; walkNodeRec(); rl--
// 				log.Printf("POSTCASE %v.%v", rl,  i)
// 			case 2:
// 				log.Printf("PRECASE %v.%v", rl, i)
// 				node[rl + 1] = node[rl].ChildNodes[rl + i]; rl++; walkNodeRec(); rl--
// 				log.Printf("POSTCASE %v.%v", rl,  i)
// 			}
// 		}
// 	}
// 	if true {
// 		rl, node[0] = 0, root
// 		log.Printf("\n\n=========>RECURSIVE ML=%v:", ml)
// 		walkNodeRec()
// 	}
// 	if true {
// 		rl, node[0] = 0, root
// 		log.Printf("\n\n=========>ITERATIVE ML=%v:", ml)
// 		walkNodeIt()
// 	}
// }

// func testTree () {
// 	var root = voxels.NewOctree(voxels.NewVolume("bunny.raw", 64), 8)
// 	var trav = voxels.NewOctreeTraverser(root)
// 	var ray = gfx.TRay {}
// 	var i int
// 	var zero, outmax, outmin float64 = 0, 68, -8
// 	var testTrav func (dx, dy, dz float64)
// 	// outmax, outmin = 60, 4
// 	testTrav = func (dx, dy, dz float64) {
// 		ray.Dir = num.Vec3 { dx, dy, dz }
// 		ray.UpdateDirMagnitude()
// 		log.Printf("RAY=(%+v)(%+v)", ray.Pos, ray.Dir)
// 		trav.Traverse(&ray)
// 		if trav.NumMatches == 0 {
// 			log.Printf("NOMATCHES")
// 		} else {
// 			for i = 0; i < trav.NumMatches; i++ {
// 				log.Printf("MATCH #%v:\tL=%v (%v,%v,%v) BOX=[%+v]", i, trav.OctNodes[i].Level, trav.OctNodes[i].NodeX, trav.OctNodes[i].NodeY, trav.OctNodes[i].NodeZ, *trav.OctNodes[i].Box)
// 			}
// 		}
// 	}
// 	log.Printf("\n\n\n\n\n\n\n\nMAXLEVEL=%v", root.MaxLevel)
// 	trav.MaxLevel = 1
// 	if true {
// 		log.Printf("\n\n=====> front ---------> 0 left-bottom")
// 		ray.Pos = num.Vec3 { 21, 24, outmin }
// 		testTrav(zero, zero, 1)
// 		log.Printf("\n\n=====> front ---------> 2 left-top")
// 		ray.Pos = num.Vec3 { 21, 39, 35 }
// 		testTrav(zero, zero, 1)
// 		log.Printf("\n\n=====> front ---------> 4 right-bottom")
// 		ray.Pos = num.Vec3 { 36, 24, outmin }
// 		testTrav(zero, zero, 1)
// 		log.Printf("\n\n=====> front ---------> 6 right-top")
// 		ray.Pos = num.Vec3 { 36, 39, outmin }
// 		testTrav(zero, zero, 1)

// 		log.Printf("\n\n=====> back ---------> 1 left-bottom")
// 		ray.Pos = num.Vec3 { 28, 24, outmax }
// 		testTrav(zero, zero, -1)
// 		log.Printf("\n\n=====> back ---------> 3 left-top")
// 		ray.Pos = num.Vec3 { 28, 46, outmax }
// 		testTrav(zero, zero, -1)
// 		log.Printf("\n\n=====> back ---------> 5 right-bottom")
// 		ray.Pos = num.Vec3 { 36, 24, outmax }
// 		testTrav(zero, zero, -1)
// 		log.Printf("\n\n=====> back ---------> 7 right-top")
// 		ray.Pos = num.Vec3 { 36, 46, outmax }
// 		testTrav(zero, zero, -1)
// 	}
// 	if false {
// 		log.Printf("\n\n=====> left ---------> 0 front-bottom")
// 		ray.Pos = num.Vec3 { outmin, 24, 28 }
// 		testTrav(1, zero, zero)
// 		log.Printf("\n\n=====> left ---------> 2 front-top")
// 		ray.Pos = num.Vec3 { outmin, 46, 28 }
// 		testTrav(1, zero, zero)
// 		log.Printf("\n\n=====> left ---------> 1 back-bottom")
// 		ray.Pos = num.Vec3 { outmin, 24, 36 }
// 		testTrav(1, zero, zero)
// 		log.Printf("\n\n=====> left ---------> 3 back-top")
// 		ray.Pos = num.Vec3 { outmin, 46, 36 }
// 		testTrav(1, zero, zero)

// 		log.Printf("\n\n=====> right ---------> 4 front-bottom")
// 		ray.Pos = num.Vec3 { outmax, 24, 28 }
// 		testTrav(-1, zero, zero)
// 		log.Printf("\n\n=====> right ---------> 6 front-top")
// 		ray.Pos = num.Vec3 { outmax, 46, 28 }
// 		testTrav(-1, zero, zero)
// 		log.Printf("\n\n=====> right ---------> 5 back-bottom")
// 		ray.Pos = num.Vec3 { outmax, 24, 36 }
// 		testTrav(-1, zero, zero)
// 		log.Printf("\n\n=====> right ---------> 7 back-top")
// 		ray.Pos = num.Vec3 { outmax, 46, 36 }
// 		testTrav(-1, zero, zero)
// 	}
// 	if false {
// 		log.Printf("\n\n=====> bottom ---------> 0 left-front")
// 		ray.Pos = num.Vec3 { 28, outmin, 24 }
// 		testTrav(zero, 1, zero)
// 		log.Printf("\n\n=====> bottom ---------> 1 left-back")
// 		ray.Pos = num.Vec3 { 28, outmin, 46 }
// 		testTrav(zero, 1, zero)
// 		log.Printf("\n\n=====> bottom ---------> 4 right-front")
// 		ray.Pos = num.Vec3 { 36, outmin, 24 }
// 		testTrav(zero, 1, zero)
// 		log.Printf("\n\n=====> bottom ---------> 5 right-back")
// 		ray.Pos = num.Vec3 { 36, outmin, 46 }
// 		testTrav(zero, 1, zero)

// 		log.Printf("\n\n=====> top ---------> 2 left-front")
// 		ray.Pos = num.Vec3 { 28, outmax, 24 }
// 		testTrav(zero, -1, zero)
// 		log.Printf("\n\n=====> top ---------> 3 left-back")
// 		ray.Pos = num.Vec3 { 28, outmax, 46 }
// 		testTrav(zero, -1, zero)
// 		log.Printf("\n\n=====> top ---------> 6 right-front")
// 		ray.Pos = num.Vec3 { 36, outmax, 24 }
// 		testTrav(zero, -1, zero)
// 		log.Printf("\n\n=====> top ---------> 7 right-back")
// 		ray.Pos = num.Vec3 { 36, outmax, 46 }
// 		testTrav(zero, -1, zero)
// 	}
// }

func main () {
	runtime.LockOSThread()
	runtime.GOMAXPROCS(16)
	gfxengine.Init()
	defer onExit()
	gfxengine.RefreshWindowCaption()
	gfxengine.Loop(onKeyDown, onKeyUp)
}

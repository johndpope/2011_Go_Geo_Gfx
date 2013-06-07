package voxels

import (
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"os"

	gfx "tclient/gfx/gfxutil"
	num "tshared/numutil"
)

type TOctree struct {
	Level, MaxLevel, Total, TotalBricks, TotalSolids, RootX, RootY, RootZ, NodeX, NodeY, NodeZ int
	Brick *TVolume
	Col gfx.TColor
	ChildNodes []*TOctree
	Box *gfx.TObject
}

func NewOctree (vol *TVolume, brickSize int) *TOctree {
	return NewOctreeNode(vol, nil, nil, brickSize, 0, 0, 0, 0, 0, 0)
}

func NewOctreeNode (vol *TVolume, parent *TOctree, root *TOctree, brickSize, nodeX, nodeY, nodeZ, rootX, rootY, rootZ int) *TOctree {
	var me = &TOctree {}
	var boxSize float64
	var allSolid = true
	var lastColor gfx.TColor
	var i = 0
	me.NodeX, me.NodeY, me.NodeZ, me.RootX, me.RootY, me.RootZ = nodeX, nodeY, nodeZ, rootX, rootY, rootZ
	if parent == nil {
		root, boxSize, me.Box, me.Total, me.Level, me.MaxLevel = me, vol.Box.Extent, vol.Box, 1, 0, 0
	} else {
		me.Level = parent.Level + 1
		boxSize = parent.Box.Extent * 0.5
		me.Box = gfx.NewBox(gfx.NilColor, num.Vec3 { parent.Box.Pos.X + (boxSize * float64(nodeX)), parent.Box.Pos.Y + (boxSize * float64(nodeY)), parent.Box.Pos.Z + (boxSize * float64(nodeZ)) }, num.Vec3 { boxSize, boxSize, boxSize }, 0)
	}
	if boxSize > float64(brickSize) {
		me.ChildNodes = make([]*TOctree, 8)
		root.Total += 8
		for x := 0; x < 2; x++ {
			for y := 0; y < 2; y++ {
				for z := 0; z < 2; z++ {
					me.ChildNodes[i] = NewOctreeNode(vol, me, root, brickSize, x, y, z, rootX + (int(boxSize / 2) * x), rootY + (int(boxSize / 2) * y), rootZ + (int(boxSize / 2) * z))
					i++
				}
			}
		}
		lastColor = me.ChildNodes[0].Col
		for _, cn := range me.ChildNodes { if (cn.Brick != nil) || !cn.Col.Equals(&lastColor) { allSolid = false } }
		if allSolid {
			me.ChildNodes = nil
			root.TotalSolids++
			me.Brick, me.Col = nil, lastColor
		} else {
			root.TotalBricks++
			me.Brick, me.Col = vol.MakeBrick(brickSize, rootX, rootY, rootZ, int(boxSize))
		}
	} else {
		root.MaxLevel = me.Level
		if me.Brick, me.Col = vol.SubVolume(me.Box, rootX, rootY, rootZ, int(boxSize)); me.Brick == nil {
			root.TotalSolids++
		} else {
			root.TotalBricks++
		}
	}
	return me
}

func (me *TOctree) Print (indent int) {
	if indent >= 0 {
		var s = ""
		for i := 0; i < indent; i++ { s += "\t" }
		log.Printf(s + "L%v C=%v B=%+v", me.Level, len(me.ChildNodes), me.Box)
		if me.Level <= 2 {
			for _, cn := range me.ChildNodes {
				cn.Print(indent + 1)
			}
		}
	}
	if indent <= 0 {
		log.Printf("LEVELS: 0..%v. TOTALS -- nodes: %v bricks: %v solids: %v", me.MaxLevel, me.Total, me.TotalBricks, me.TotalSolids)
	}
}

type TOctreeTraverser struct {
	Color gfx.TColor
	Root *TOctree
	Ray *gfx.TRay
	Level, MaxLevel int
	Lights []*gfx.TLight

	foundEntry bool
	tempCol, tempCol2, tempCol3, lightCol gfx.TColor
	tnode *TOctree
	travNode []*TOctree
	lightRay gfx.TRay
	brickPos, rayPos, rayDir, rayInv, voxelNormal, lightVec num.Vec3
	vx, vy, vz, nvx, nvy, nvz, a, rl, n0, n1, n2 int
	light *gfx.TLight
	nx, ny, nz, tmin, tmax, tmp, levLevel, maxLevel, lightDiffuse, lightDot float64
	cn []int
	tx0, ty0, tz0, tx1, ty1, tz1, mx, my, mz []float64
}

func NewOctreeTraverser (root *TOctree, lights []*gfx.TLight) *TOctreeTraverser {
	var ml = root.MaxLevel + 1
	var me = &TOctreeTraverser {}
	me.Lights = lights
	me.MaxLevel, me.Root = root.MaxLevel, root
	me.tx0, me.ty0, me.tz0, me.tx1, me.ty1, me.tz1, me.mx, me.my, me.mz = make([]float64, ml), make([]float64, ml), make([]float64, ml), make([]float64, ml), make([]float64, ml), make([]float64, ml), make([]float64, ml), make([]float64, ml), make([]float64, ml)
	me.maxLevel, me.cn, me.travNode = 1 / float64(me.MaxLevel), make([]int, ml), make([]*TOctree, ml)
	return me
}

func (me *TOctreeTraverser) KdRestart () {
	/*
	tMin = tMax = sceneMin
	tHit = infinity
	for tMax < sceneMax {
		node = root
		tMin = tMax
		tMax = sceneMax
		for !node.isLeaf() {

		}
		// now node.IsLeaf!
		if node.isBrick {
			col.add(node.sampleBrick())
		} else {
			col.add(node.color)
		}
	}
	*/
}

func (me *TOctreeTraverser) Traverse () {
	me.Color.Reset()
	me.rl, me.rayPos, me.rayDir, me.rayInv, me.a = 0, me.Ray.Pos, me.Ray.Dir, me.Ray.InvDir, 0
	if me.rayDir.X == 0 { me.rayDir.X = num.Epsilon; me.rayInv.X = 1 / me.rayDir.X }
	if me.rayDir.Y == 0 { me.rayDir.Y = num.Epsilon; me.rayInv.Y = 1 / me.rayDir.Y }
	if me.rayDir.Z == 0 { me.rayDir.Z = num.Epsilon; me.rayInv.Z = 1 / me.rayDir.Z }
	if me.rayDir.X < 0 { me.rayPos.X = me.Root.Box.Size.X - me.rayPos.X; me.rayDir.X = -me.rayDir.X; me.a = me.a | 4; me.rayInv.X = -me.rayInv.X }
	if me.rayDir.Y < 0 { me.rayPos.Y = me.Root.Box.Size.Y - me.rayPos.Y; me.rayDir.Y = -me.rayDir.Y; me.a = me.a | 2; me.rayInv.Y = -me.rayInv.Y }
	if me.rayDir.Z < 0 { me.rayPos.Z = me.Root.Box.Size.Z - me.rayPos.Z; me.rayDir.Z = -me.rayDir.Z; me.a = me.a | 1; me.rayInv.Z = -me.rayInv.Z }
	me.tx0[me.rl], me.ty0[me.rl], me.tz0[me.rl] = (me.Root.Box.Min.X - me.rayPos.X) * me.rayInv.X, (me.Root.Box.Min.Y - me.rayPos.Y) * me.rayInv.Y, (me.Root.Box.Min.Z - me.rayPos.Z) * me.rayInv.Z
	me.tx1[me.rl], me.ty1[me.rl], me.tz1[me.rl] = (me.Root.Box.Max.X - me.rayPos.X) * me.rayInv.X, (me.Root.Box.Max.Y - me.rayPos.Y) * me.rayInv.Y, (me.Root.Box.Max.Z - me.rayPos.Z) * me.rayInv.Z
	me.tmin, me.tmax = math.Max(me.tx0[me.rl], math.Max(me.ty0[me.rl], me.tz0[me.rl])), math.Min(me.tx1[me.rl], math.Min(me.ty1[me.rl], me.tz1[me.rl]))
	if ((me.tmin < me.tmax) && (me.tmax > 0)) {
		me.rayDir.SetFromMult1(&me.Ray.Dir, 0.125)
		me.travNode[me.rl] = me.Root
		me.cn[0] = -1
		for me.rl >= 0 {
			me.tnode = me.travNode[me.rl]
			if (me.tx1[me.rl] < 0) || (me.ty1[me.rl] < 0) || (me.tz1[me.rl] < 0) {
				me.rl--
			} else if (me.tnode.Level == me.MaxLevel) || (me.tnode.ChildNodes == nil) {
				me.tempCol.Reset()
				if me.tnode.Brick != nil {
					if me.Ray.Pos.AllInside(&me.tnode.Box.Min, &me.tnode.Box.Max) {
						me.brickPos.SetFrom(&me.Ray.Pos)
					} else {
						me.tmin = math.Max(me.tx0[me.rl], math.Max(me.ty0[me.rl], me.tz0[me.rl]))
						me.brickPos.SetFromAddMult1(&me.Ray.Pos, &me.Ray.Dir, me.tmin)
					}
					me.vx = int(((me.brickPos.X - me.tnode.Box.Min.X) * me.tnode.Brick.InvScale))
					me.vy = int(((me.brickPos.Y - me.tnode.Box.Min.Y) * me.tnode.Brick.InvScale))
					me.vz = int(((me.brickPos.Z - me.tnode.Box.Min.Z) * me.tnode.Brick.InvScale))
					for (me.tempCol.A < 1) && (me.vx >= 0) && (me.vx <= me.tnode.Brick.Size) && (me.vy >= 0) && (me.vy <= me.tnode.Brick.Size) && (me.vz >= 0) && (me.vz <= me.tnode.Brick.Size) {
						if (me.vx < me.tnode.Brick.Size) && (me.vy < me.tnode.Brick.Size) && (me.vz < me.tnode.Brick.Size) {
							me.tempCol2.SetFrom(&me.tnode.Brick.Col[me.vx][me.vy][me.vz])
							if me.tnode.Level == me.Root.MaxLevel {
								me.lightCol.R, me.lightCol.G, me.lightCol.B, me.lightCol.A = 0.66, 0.66, 0.66, 1
								me.voxelNormal = me.tnode.Brick.Normals[me.vx][me.vy][me.vz]
								me.voxelNormal.SwapSigns()
								me.lightRay.Pos = me.brickPos
								for _, me.light = range me.Lights {
									me.lightRay.Dir.SetFromSub(&me.light.RotPos, &me.lightRay.Pos)
									me.lightRay.UpdateDirMagnitude()
									me.lightDiffuse = (me.lightRay.Dir.Dot(&me.voxelNormal) / me.lightRay.DirLength) * me.light.Intensity
									if me.lightDiffuse > 0 {
										me.lightCol.R, me.lightCol.G, me.lightCol.B = me.lightCol.R + (me.light.Col.R * me.lightDiffuse), me.lightCol.G + (me.light.Col.G * me.lightDiffuse), me.lightCol.B + (me.light.Col.B * me.lightDiffuse)
									}
									// me.lightDot = me.voxelNormal.Dot(&me.LightDir)
									// me.lightShade = math.Max(me.lightDot, 0) * 0.8
									// me.lightVec.SetFromSubMult1(&me.LightDir, &me.voxelNormal, 2 * me.lightDot)
									// me.lightVec.SwapSigns()
									// me.lightShade += (0.2 * (math.Pow(math.Max(0, me.lightVec.Dot(&me.LightDir)), 10)))
									// me.lightShade = math.Min(me.lightShade, 1)
									// me.tempCol2.Mult(me.lightShade)
								}
								me.tempCol2.Mult(&me.lightCol)
							}
							me.tempCol3.Blend(&me.tempCol, &me.tempCol2)
							me.tempCol.SetFrom(&me.tempCol3)
						}
						BrickStep:
							me.brickPos.Add(&me.Ray.Dir)
							me.nvx = int((me.brickPos.X - me.tnode.Box.Min.X) * me.tnode.Brick.InvScale)
							me.nvy = int((me.brickPos.Y - me.tnode.Box.Min.Y) * me.tnode.Brick.InvScale)
							me.nvz = int((me.brickPos.Z - me.tnode.Box.Min.Z) * me.tnode.Brick.InvScale)
						if (me.nvx == me.vx) && (me.nvy == me.vy) && (me.nvz == me.vz) {
							goto BrickStep
						} else {
							me.vx, me.vy, me.vz = me.nvx, me.nvy, me.nvz
						}
					}
					// me.tempCol.R, me.tempCol.G, me.tempCol.B, me.tempCol.A = float64(me.tnode.NodeX) * me.levLevel, float64(me.tnode.NodeY) * me.levLevel, float64(me.tnode.NodeZ) * me.levLevel, 0.5
				} else if me.tnode.Col.A > 0 {
					me.tempCol.SetFrom(&me.tnode.Col)
				// } else {
					// me.levLevel = me.maxLevel * float64(me.tnode.Level)
					// me.tempCol.R, me.tempCol.G, me.tempCol.B, me.tempCol.A = /*float64(me.tnode.NodeX) * me.levLevel, float64(me.tnode.NodeY) * me.levLevel, float64(me.tnode.NodeZ) * me.levLevel*/ 0.1, 0.1, 0.1, 0.1
				}
				me.tempCol2.Blend(&me.Color, &me.tempCol)
				me.Color.SetFrom(&me.tempCol2)
				if me.Color.A >= 1 { return }
				me.rl--
			} else {
				if me.cn[me.rl] < 0 {
					// FIRST NODE
					me.cn[me.rl] = 0
					me.mx[me.rl] = 0.5 * (me.tx0[me.rl] + me.tx1[me.rl])
					me.my[me.rl] = 0.5 * (me.ty0[me.rl] + me.ty1[me.rl])
					me.mz[me.rl] = 0.5 * (me.tz0[me.rl] + me.tz1[me.rl])
					if (me.tx0[me.rl] > me.ty0[me.rl]) && (me.tx0[me.rl] > me.tz0[me.rl]) {
						if me.my[me.rl] < me.tx0[me.rl] { me.cn[me.rl] = me.cn[me.rl] | 2 }	//	1	1	2
						if me.mz[me.rl] < me.tx0[me.rl] { me.cn[me.rl] = me.cn[me.rl] | 1 }	//	0	2	1
					}
					if (me.ty0[me.rl] > me.tx0[me.rl]) && (me.ty0[me.rl] > me.tz0[me.rl]) {
						if me.mx[me.rl] < me.ty0[me.rl] { me.cn[me.rl] = me.cn[me.rl] | 4 }	//	2	0	4
						if me.mz[me.rl] < me.ty0[me.rl] { me.cn[me.rl] = me.cn[me.rl] | 1 }	//	0	2	1
					}
					if (me.tz0[me.rl] > me.tx0[me.rl]) && (me.tz0[me.rl] > me.ty0[me.rl]) {
						if me.mx[me.rl] < me.tz0[me.rl] { me.cn[me.rl] = me.cn[me.rl] | 4 }	//	2	0	4
						if me.my[me.rl] < me.tz0[me.rl] { me.cn[me.rl] = me.cn[me.rl] | 2 }	//	1	1	2
					}
				} else {
					// NEXT NODE
					switch me.cn[me.rl] {
						case 0: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = me.mx[me.rl], me.my[me.rl], me.mz[me.rl], 4, 2, 1
						case 1: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = me.mx[me.rl], me.my[me.rl], me.tz1[me.rl], 5, 3, 8
						case 2: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = me.mx[me.rl], me.ty1[me.rl], me.mz[me.rl], 6, 8, 3
						case 3: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = me.mx[me.rl], me.ty1[me.rl], me.tz1[me.rl], 7, 8, 8
						case 4: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = me.tx1[me.rl], me.my[me.rl], me.mz[me.rl], 8, 6, 5
						case 5: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = me.tx1[me.rl], me.my[me.rl], me.tz1[me.rl], 8, 7, 8
						case 6: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = me.tx1[me.rl], me.ty1[me.rl], me.mz[me.rl], 8, 8, 7
						case 7: me.nx, me.ny, me.nz, me.n0, me.n1, me.n2 = 0, 0, 0, 8, 8, 8
					}
					me.cn[me.rl] = 8; if (me.nx < me.ny) && (me.nx < me.nz) { me.cn[me.rl] = me.n0 }; if (me.ny < me.nx) && (me.ny < me.nz) { me.cn[me.rl] = me.n1 }; if (me.nz < me.nx) && (me.nz < me.ny) { me.cn[me.rl] = me.n2 }
				}
				if me.cn[me.rl] < 8 {
					me.travNode[me.rl + 1] = me.tnode.ChildNodes[me.cn[me.rl] ^ me.a]
					me.cn[me.rl + 1] = -1
					switch me.cn[me.rl] {
						case 0: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.tx0[me.rl], me.ty0[me.rl], me.tz0[me.rl], me.mx[me.rl], me.my[me.rl], me.mz[me.rl]
						case 1: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.tx0[me.rl], me.ty0[me.rl], me.mz[me.rl], me.mx[me.rl], me.my[me.rl], me.tz1[me.rl]
						case 2: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.tx0[me.rl], me.my[me.rl], me.tz0[me.rl], me.mx[me.rl], me.ty1[me.rl], me.mz[me.rl]
						case 3: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.tx0[me.rl], me.my[me.rl], me.mz[me.rl], me.mx[me.rl], me.ty1[me.rl], me.tz1[me.rl]
						case 4: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.mx[me.rl], me.ty0[me.rl], me.tz0[me.rl], me.tx1[me.rl], me.my[me.rl], me.mz[me.rl]
						case 5: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.mx[me.rl], me.ty0[me.rl], me.mz[me.rl], me.tx1[me.rl], me.my[me.rl], me.tz1[me.rl]
						case 6: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.mx[me.rl], me.my[me.rl], me.tz0[me.rl], me.tx1[me.rl], me.ty1[me.rl], me.mz[me.rl]
						case 7: me.tx0[me.rl + 1], me.ty0[me.rl + 1], me.tz0[me.rl + 1], me.tx1[me.rl + 1], me.ty1[me.rl + 1], me.tz1[me.rl + 1] = me.mx[me.rl], me.my[me.rl], me.mz[me.rl], me.tx1[me.rl], me.ty1[me.rl], me.tz1[me.rl]
					}
					me.rl++
				} else {
					me.rl--
				}
			}
		}
	}
}

type TVolume struct {
	Col [][][]gfx.TColor
	Normals [][][]num.Vec3
	Box *gfx.TObject
	Size int
	InvScale, Scale float64
}

func NewVolume (filePath string, volSize int) *TVolume {
	var fsize = float64(volSize)
	var normVec num.Vec3
	var x, y, z int
	var prev, next uint8
	var max = volSize - 1
	var raw = LoadVolumeRaw(filePath, volSize)
	var invSize, inv256 = 1 / float64(volSize), 1.0 / 256.0
	var me = &TVolume {}
	me.InvScale, me.Scale = 1, 1
	me.Box = gfx.NewBox(gfx.NilColor, num.Vec3 { 0, 0, 0 }, num.Vec3 { fsize, fsize, fsize }, 0)
	me.Col, me.Normals, me.Size = make([][][]gfx.TColor, volSize), make([][][]num.Vec3, volSize), volSize
	for x = 0; x < volSize; x++ {
		me.Col[x], me.Normals[x] = make([][]gfx.TColor, volSize), make([][]num.Vec3, volSize)
		for y = 0; y < volSize; y++ {
			// log.Printf("ALLOC %v,%v...", x, y)
			me.Col[x][y], me.Normals[x][y] = make([]gfx.TColor, volSize), make([]num.Vec3, volSize)
		}
	}
	for x = 0; x < volSize; x++ {
		for y = 0; y < volSize; y++ {
			for z = 0; z < volSize; z++ {
				if raw[x][y][z] > 0 {
					me.Col[x][y][z].R = float64(x) * invSize
					me.Col[x][y][z].G = float64(y) * invSize
					me.Col[x][y][z].B = float64(z) * invSize
					me.Col[x][y][z].A = float64(raw[x][y][z]) * inv256 // float64(raw[x][y][z].A) / 256
					if (x > 0) { prev = raw[x - 1][y][z] } else { prev = 0 }
					if (x < max) { next = raw[x + 1][y][z] } else { next = 0 }
					normVec.X = (float64(next) * inv256) - (float64(prev) * inv256)
					if (y > 0) { prev = raw[x][y - 1][z] } else { prev = 0 }
					if (y < max) { next = raw[x][y + 1][z] } else { next = 0 }
					normVec.Y = (float64(next) * inv256) - (float64(prev) * inv256)
					if (z > 0) { prev = raw[x][y][z - 1] } else { prev = 0 }
					if (z < max) { next = raw[x][y][z + 1] } else { next = 0 }
					normVec.Z = (float64(next) * inv256) - (float64(prev) * inv256)
					normVec.Normalize()
					me.Normals[x][y][z] = normVec
				}
			}
		}
	}
	return me
}

func (me *TVolume) MakeBrick (brickSize, volX, volY, volZ, subVolSize int) (*TVolume, gfx.TColor) {
	var x, y, z int
	var nu = &TVolume {}
	var filter = NewVoxelFilter(me)
	var lastColor = gfx.NilColor
	var solid = true
	filter.N, filter.SubVolSize = brickSize, subVolSize
	nu.Box = me.Box
	nu.Col, nu.Size = make([][][]gfx.TColor, brickSize), brickSize
	nu.InvScale = float64(brickSize) / float64(subVolSize)
	nu.Scale = float64(subVolSize) / float64(brickSize)
	for x = 0; x < brickSize; x++ {
		nu.Col[x] = make([][]gfx.TColor, brickSize)
		for y = 0; y < brickSize; y++ {
			nu.Col[x][y] = make([]gfx.TColor, brickSize)
		}
	}
	for x = 0; x < brickSize; x++ {
		for y = 0; y < brickSize; y++ {
			for z = 0; z < brickSize; z++ {
				filter.LinearN(x, y, z, volX, volY, volZ)
				if (x != 0) && (y != 0) && (z != 0) && (filter.OutCol.A != 0) { solid = false }
				lastColor, nu.Col[x][y][z] = filter.OutCol, filter.OutCol
			}
		}
	}
	if solid { nu = nil }
	return nu, lastColor
}

func (me *TVolume) SubVolume (box *gfx.TObject, sx, sy, sz, volSize int) (*TVolume, gfx.TColor) {
	var nu, lastColor, thisColor, solid, x, y, z = &TVolume {}, gfx.NilColor, gfx.NilColor, true, 0, 0, 0
	nu.Box, nu.Size, nu.InvScale, nu.Scale = box, volSize, 1, 1
	nu.Col, nu.Normals = make([][][]gfx.TColor, volSize), make([][][]num.Vec3, volSize)
	for x = 0; x < volSize; x++ {
		nu.Col[x], nu.Normals[x] = make([][]gfx.TColor, volSize), make([][]num.Vec3, volSize)
		for y = 0; y < volSize; y++ {
			nu.Col[x][y] = make([]gfx.TColor, volSize)
			nu.Normals[x][y] = make([]num.Vec3, volSize)
		}
	}
	lastColor = me.Col[sx][sy][sz]
	for x = 0; x < volSize; x++ {
		for y = 0; y < volSize; y++ {
			for z = 0; z < volSize; z++ {
				if thisColor = me.Col[sx + x][sy + y][sz + z]; thisColor.A != 0 { solid = false }
				lastColor, nu.Col[x][y][z] = thisColor, thisColor
				nu.Normals[x][y][z] = me.Normals[sx + x][sy + y][sz + z]
			}
		}
	}
	if solid { nu = nil }
	return nu, lastColor
}

type TVolumeRgba8 [][][]color.RGBA

type TVolumeRaw [][][]uint8

type TVoxelFilter struct {
	OutCol gfx.TColor
	Vol *TVolume
	N, SubVolSize int
	StrictAlpha bool
	xceil, yceil, zceil, xfloor, yfloor, zfloor, n3, vs, vs3, x, y, z, i, numc int
	xfrac, yfrac, zfrac, alpha, nweight float64
	weights []float64
	ncols, cols []gfx.TColor
}

func NewVoxelFilter (vol *TVolume) *TVoxelFilter {
	var me = &TVoxelFilter {}
	me.Vol, me.weights, me.cols = vol, make([]float64, 8), make([]gfx.TColor, 8)
	if vol != nil {
		me.SubVolSize = vol.Size
	}
	return me
}

func (me *TVoxelFilter) Linear2 (pos *num.Vec3) {
	me.xceil, me.yceil, me.zceil = int(math.Ceil(pos.X * me.Vol.InvScale)), int(math.Ceil(pos.Y * me.Vol.InvScale)), int(math.Ceil(pos.Z * me.Vol.InvScale))
	me.xfloor, me.yfloor, me.zfloor = int(math.Floor(pos.X * me.Vol.InvScale)), int(math.Floor(pos.Y * me.Vol.InvScale)), int(math.Floor(pos.Z * me.Vol.InvScale))
	_, me.xfrac = math.Modf(pos.X * me.Vol.InvScale)
	_, me.yfrac = math.Modf(pos.Y * me.Vol.InvScale)
	_, me.zfrac = math.Modf(pos.Z * me.Vol.InvScale)

	me.weights[0] = (1 - me.xfrac) * (1 - me.yfrac) * (1 - me.zfrac)
	me.weights[1] = (1 - me.xfrac) * (1 - me.yfrac) * me.zfrac
	me.weights[2] = (1 - me.xfrac) * me.yfrac * (1 - me.zfrac)
	me.weights[3] = (1 - me.xfrac) * me.yfrac * me.zfrac
	me.weights[4] = me.xfrac * (1 - me.yfrac) * (1 - me.zfrac)
	me.weights[5] = me.xfrac * (1 - me.yfrac) * me.zfrac
	me.weights[6] = me.xfrac * me.yfrac * (1 - me.zfrac)
	me.weights[7] = me.xfrac * me.yfrac * me.zfrac

	if me.xceil >= me.SubVolSize { me.weights[4], me.weights[5], me.weights[6], me.weights[7] = -1, -1, -1, -1 }
	if me.yceil >= me.SubVolSize { me.weights[2], me.weights[3], me.weights[6], me.weights[7] = -1, -1, -1, -1 }
	if me.zceil >= me.SubVolSize { me.weights[1], me.weights[3], me.weights[5], me.weights[7] = -1, -1, -1, -1 }

	if me.weights[0] == -1 { me.cols[0].Reset() } else { me.cols[0] = me.Vol.Col[me.xfloor][me.yfloor][me.zfloor] }
	if me.weights[1] == -1 { me.cols[1].Reset() } else { me.cols[1] = me.Vol.Col[me.xfloor][me.yfloor][me.zceil] }
	if me.weights[2] == -1 { me.cols[2].Reset() } else { me.cols[2] = me.Vol.Col[me.xfloor][me.yceil][me.zfloor] }
	if me.weights[3] == -1 { me.cols[3].Reset() } else { me.cols[3] = me.Vol.Col[me.xfloor][me.yceil][me.zceil] }
	if me.weights[4] == -1 { me.cols[4].Reset() } else { me.cols[4] = me.Vol.Col[me.xceil][me.yfloor][me.zfloor] }
	if me.weights[5] == -1 { me.cols[5].Reset() } else { me.cols[5] = me.Vol.Col[me.xceil][me.yfloor][me.zceil] }
	if me.weights[6] == -1 { me.cols[6].Reset() } else { me.cols[6] = me.Vol.Col[me.xceil][me.yceil][me.zfloor] }
	if me.weights[7] == -1 { me.cols[7].Reset() } else { me.cols[7] = me.Vol.Col[me.xceil][me.yceil][me.zceil] }

	me.alpha = (me.cols[0].A + me.cols[1].A + me.cols[2].A + me.cols[3].A + me.cols[4].A + me.cols[5].A + me.cols[6].A + me.cols[7].A) / 8
	if me.alpha == 0 {
		me.OutCol.Reset()
	} else {
		me.OutCol.R = ((me.cols[0].R * me.weights[0] * me.cols[0].A) + (me.cols[1].R * me.weights[1] * me.cols[1].A) + (me.cols[2].R * me.weights[2] * me.cols[2].A) + (me.cols[3].R * me.weights[3] * me.cols[3].A) + (me.cols[4].R * me.weights[4] * me.cols[4].A) + (me.cols[5].R * me.weights[5] * me.cols[5].A) + (me.cols[6].R * me.weights[6] * me.cols[6].A) + (me.cols[7].R * me.weights[7] * me.cols[7].A)) / me.alpha
		me.OutCol.G = ((me.cols[0].G * me.weights[0] * me.cols[0].A) + (me.cols[1].G * me.weights[1] * me.cols[1].A) + (me.cols[2].G * me.weights[2] * me.cols[2].A) + (me.cols[3].G * me.weights[3] * me.cols[3].A) + (me.cols[4].G * me.weights[4] * me.cols[4].A) + (me.cols[5].G * me.weights[5] * me.cols[5].A) + (me.cols[6].G * me.weights[6] * me.cols[6].A) + (me.cols[7].G * me.weights[7] * me.cols[7].A)) / me.alpha
		me.OutCol.B = ((me.cols[0].B * me.weights[0] * me.cols[0].A) + (me.cols[1].B * me.weights[1] * me.cols[1].A) + (me.cols[2].B * me.weights[2] * me.cols[2].A) + (me.cols[3].B * me.weights[3] * me.cols[3].A) + (me.cols[4].B * me.weights[4] * me.cols[4].A) + (me.cols[5].B * me.weights[5] * me.cols[5].A) + (me.cols[6].B * me.weights[6] * me.cols[6].A) + (me.cols[7].B * me.weights[7] * me.cols[7].A)) / me.alpha
		me.OutCol.A = (me.cols[0].A * me.weights[0]) + (me.cols[1].A * me.weights[1]) + (me.cols[2].A * me.weights[2]) + (me.cols[3].A * me.weights[3]) + (me.cols[4].A * me.weights[4]) + (me.cols[5].A * me.weights[5]) + (me.cols[6].A * me.weights[6]) + (me.cols[7].A * me.weights[7])
	}
}

func (me *TVoxelFilter) LinearN (bx, by, bz, vx, vy, vz int) {
	var tcol gfx.TColor
	me.vs = me.SubVolSize / me.N // 4
	me.vs3 = me.vs * me.vs * me.vs // 64
	me.n3 = me.N * me.N * me.N // 512
	if len(me.ncols) < me.vs3 { me.ncols = make([]gfx.TColor, me.vs3) }
	me.numc, me.alpha = 0, 0
	me.OutCol.Reset()
	vx, vy, vz = vx + (me.vs * bx), vy + (me.vs * by), vz + (me.vs * bz)
	me.i = 0
	for me.x = 0; me.x < me.vs; me.x++ {
		for me.y = 0; me.y < me.vs; me.y++ {
			for me.z = 0; me.z < me.vs; me.z++ {
				// me.i = CubicIndex(me.x, me.y, me.z, me.vs, me.vs)
				tcol = me.Vol.Col[vx + me.x][vy + me.y][vz + me.z]
				me.ncols[me.i] = tcol
				me.alpha += me.ncols[me.i].A
				if me.StrictAlpha || (me.ncols[me.i].A != 0) { me.numc++ }
				me.i++
			}
		}
	}
	me.nweight = 1 / float64(me.numc)
	me.alpha = me.alpha / float64(me.numc)
	if me.alpha != 0 {
		for me.i = 0; me.i < len(me.ncols); me.i++ {
			if me.StrictAlpha || (me.ncols[me.i].A > 0) {
				me.OutCol.R += (me.ncols[me.i].R * me.nweight * me.ncols[me.i].A)
				me.OutCol.G += (me.ncols[me.i].G * me.nweight * me.ncols[me.i].A)
				me.OutCol.B += (me.ncols[me.i].B * me.nweight * me.ncols[me.i].A)
				me.OutCol.A += (me.ncols[me.i].A * me.nweight)
			}
		}
		me.OutCol.R, me.OutCol.G, me.OutCol.B = me.OutCol.R / me.alpha, me.OutCol.G / me.alpha, me.OutCol.B / me.alpha
		if me.OutCol.A > 1 { me.OutCol.A = 1 }
		if me.OutCol.A <= 0 { me.OutCol.Reset() }
	}
}

func CubicIndex (x, y, z, xsize, ysize int) int {
	return (((z * xsize) + x) * ysize) + y
}

func LoadVolumeRaw (filePath string, size int) TVolumeRaw {
	var x, y, z, i int
	var f float64
	var fs = float64(size)
	var fs2  = fs * fs
	var file *os.File
	var data []byte
	var bval byte
	var err error
	var vol = make([][][]uint8, size)
	for x = 0; x < size; x++ {
		vol[x] = make([][]uint8, size)
		for y = 0; y < size; y++ {
			vol[x][y] = make([]uint8, size)
		}
	}
	if file, err = os.Open(filePath); err != nil { panic(err) }
	defer file.Close()
	if data, err = ioutil.ReadAll(file); err != nil {
		panic(err)
	} else {
		z = -1
		y = -1
		x = -1
		for i = 0; i < len(data); i++ {
			f = float64(i)
			bval = data[i]
			if num.IsInt(f / fs2) { z++; y = -1; x = -1 }
			if num.IsInt(f / fs) { y++; x = -1 }
			x++
			vol[x][y][z] = bval
		}
	}
	return TVolumeRaw(vol)
}

func LoadVolumeRgba8 (filePath string, size int) TVolumeRgba8 {
	var x, y, z, i int
	var f float64
	var fs = float64(size)
	var fsf = 256 / fs
	var fs2  = fs * fs
	var file *os.File
	var data []byte
	var bval byte
	var err error
	var vol = make([][][]color.RGBA, size)
	for x = 0; x < size; x++ {
		vol[x] = make([][]color.RGBA, size)
		for y = 0; y < size; y++ {
			vol[x][y] = make([]color.RGBA, size)
		}
	}
	if file, err = os.Open(filePath); err != nil { panic(err) }
	defer file.Close()
	if data, err = ioutil.ReadAll(file); err != nil {
		panic(err)
	} else {
		z = -1
		y = -1
		x = -1
		for i = 0; i < len(data); i++ {
			f = float64(i)
			bval = data[i]
			if num.IsInt(f / fs2) { z++; y = -1; x = -1 }
			if num.IsInt(f / fs) { y++; x = -1 }
			x++
			vol[x][y][z].R = uint8(float64(x) * fsf)
			vol[x][y][z].G = uint8(float64(y) * fsf)
			vol[x][y][z].B = uint8(float64(z) * fsf)
			vol[x][y][z].A = bval
		}
	}
	return TVolumeRgba8(vol)
}

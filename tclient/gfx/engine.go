package engine

import (
	"fmt"
	"log"
	"time"

	gl "github.com/chsc/gogl/gl42"
	"github.com/paul-lalonde/Go-SDL/sdl"

	"tclient/gfx/canvas"
	"tclient/gfx/glsl"
	"tclient/gfx/gltypes"
	"tclient/gfx/glutil"
	"tclient/gfx/pipeline"
	"tshared/coreutil"
	"tshared/stringutil"
)

const SCREEN_RETRO, SCREEN_DEFWIDTH, SCREEN_DEFHEIGHT = 1, 1440, 810 // 640, 360 // 960, 540 // 1280, 720 // 

var (
	Fps = 0
	FullScreen, HasInitGl, HasInitSdl = false, false, false
	KeySymTmpTexUp, KeySymTmpTexDn uint32 = sdl.K_F1, sdl.K_F2
	// KeySymOctreeLevelUp, KeySymOctreeLevelDown uint32 = sdl.K_F1, sdl.K_F2
	Looping = true
	KeysPressed = map[uint32]bool {}
	MouseX, MouseY float64
	Screen *sdl.Surface
	Canvases = []*canvas.Canvas {}
	Canvas0 *canvas.Canvas
	Shaders *gltypes.ShaderManager
	StartTime int64
	TmpTexIndex = 8
	TmpTexes = make([]gl.Uint, 9)
)

func glInit () {
	var err error
	StartTime = time.Now().UnixNano()
	if err = gl.Init(); err != nil { panic(err) }
	gl.ClearColor(1, 0, 0, 1)
}

func AddCanvas (newVerts []gl.Float, width, height int, retroFactor float64) int {
	var canvas canvas.Canvas
	var l int
	canvas.TmpTex = TmpTexes[TmpTexIndex]
	canvas.Index = gl.Int(len(Canvases))
	canvas.Init(width, height, retroFactor)
	canvas.Shaders = Shaders
	Canvases = append(Canvases, &canvas)
	if l = len(Canvases); l == 1 { Canvas0 = &canvas }
	return l
}

func CleanUp () {
	SetWindowCaption("Exiting...")
	for len(Canvases) > 0 { RemoveCanvas(len(Canvases) - 1) }
	pipeline.CleanUp(false)
	Shaders.CleanUp()
}

func Init () {
	var sdlInit int
	if sdlInit = sdl.Init(sdl.INIT_VIDEO); sdlInit != 0 { log.Panicf("Could not initialize SDL (return code %d), error message: %s\n", sdlInit, sdl.GetError()) }
	HasInitSdl = true
	ReinitVideo(SCREEN_DEFWIDTH, SCREEN_DEFHEIGHT, true, false)
	SetWindowCaption("Init OpenGL...")
	glInit()
	HasInitGl = true
	SetWindowCaption("Compiling shaders...")
	Shaders = gltypes.NewShaderManager()
	RecompileShaders()
	SetWindowCaption("Init rendering pipeline...")
	pipeline.Reinit()
	LoadTmpTex(8)
	AddCanvas(nil, int(float64(Screen.W)), int(float64(Screen.H)), SCREEN_RETRO)
	ReinitVideo(SCREEN_DEFWIDTH, SCREEN_DEFHEIGHT, false, true)
}

func LoadTmpTex (index int) {
	if index < 0 { index = len(TmpTexes) - 1 } else if index >= len(TmpTexes) { index = 0 }
	TmpTexIndex = index
	if (TmpTexes[index] == 0) && (index == 0) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/blob.exr_1040x1040.float", 1040, 1040) }
	if (TmpTexes[index] == 0) && (index == 1) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/candle.exr_1000x810.float", 1000, 810) }
	if (TmpTexes[index] == 0) && (index == 2) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/cannon.exr_780x566.float", 780, 566) }
	if (TmpTexes[index] == 0) && (index == 5) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/desk.exr_644x874.float", 644, 874) }
	if (TmpTexes[index] == 0) && (index == 6) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/memorial.exr_512x768.float", 512, 768) }
	if (TmpTexes[index] == 0) && (index == 3) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/prisms.exr_1200x865.float", 1200, 865) }
	if (TmpTexes[index] == 0) && (index == 4) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/still.exr_1240x846.float", 1240, 846) }
	if (TmpTexes[index] == 0) && (index == 7) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/tree.exr_928x906.float", 928, 906) }
	if (TmpTexes[index] == 0) && (index == 8) { TmpTexes[index] = glutil.MakeTextureFromImageFloatsFile("/ssd2/exr/west.exr_1214x732.float", 1214, 732) }
	if Canvas0 != nil { Canvas0.TmpTex = TmpTexes[TmpTexIndex] }
}

func Loop (onKeyDown, onKeyUp func(*sdl.KeyboardEvent)) {
	var nowTime = time.Now()
	var lastSecond, lastTime = nowTime, nowTime
	var nowNano int64 = 0
	var durSec time.Duration
	var sdlEvt sdl.Event
	var winw, winh int
	var retro int = int(Canvas0.RetroFactor)
	var canvas0RetroFactor = 1 / Canvas0.RetroFactor
	gl.Enable(gl.FRAMEBUFFER_SRGB)
	for Looping {
		nowTime = time.Now()
		pipeline.TimeLast, pipeline.TimeNow = lastTime, nowTime
		pipeline.TimeSecsElapsed = nowTime.Sub(lastTime).Seconds()
		lastTime = nowTime
		nowNano = nowTime.UnixNano()
		Canvas0.TimeElapsed = nowNano - StartTime
		durSec = nowTime.Sub(lastSecond)
		if sdlEvt = sdl.PollEvent(); sdlEvt != nil {
			switch event := sdlEvt.(type) {
				case *sdl.ResizeEvent:
					winw, winh = int(event.W), int(event.H)
					if retro > 0 {
						for (winw % retro) != 0 { winw-- }
						for (winh % retro) != 0 { winh-- }
						if (winw / retro) < 32 { winw = 32 * retro }
						if (winh / retro) < 32 { winh = 32 * retro }
					}
					ReinitVideo(winw, winh, true, true)
					RefreshWindowCaption()
				case *sdl.QuitEvent:
					Looping = false
					return
				case *sdl.KeyboardEvent:
					if event.Type == sdl.KEYUP {
						KeysPressed[event.Keysym.Sym] = false
						if event.Keysym.Sym == KeySymTmpTexUp { LoadTmpTex(TmpTexIndex + 1) }
						if event.Keysym.Sym == KeySymTmpTexDn { LoadTmpTex(TmpTexIndex - 1) }
						// if event.Keysym.Sym == KeySymOctreeLevelUp { softpipeline.OctreeMaxLevel++ }
						// if event.Keysym.Sym == KeySymOctreeLevelDown { softpipeline.OctreeMaxLevel-- }
						onKeyUp(event)
					} else if event.Type == sdl.KEYDOWN {
						KeysPressed[event.Keysym.Sym] = true
						onKeyDown(event)
					}
				case *sdl.MouseMotionEvent:
					MouseX, MouseY = float64(event.X) * canvas0RetroFactor, float64(event.Y) * canvas0RetroFactor
					Canvas0.MouseX, Canvas0.MouseY = gl.Float(MouseX * Canvas0.InvWidth), gl.Float(MouseY * Canvas0.InvHeight)
			}
		}
		pipeline.CamMove, pipeline.CamTurn = false, false
		if KeysPressed[sdl.K_UP] { pipeline.CamMove, pipeline.CamMoveFwd = true, 1 } else { pipeline.CamMoveFwd = 0 }
		if KeysPressed[sdl.K_DOWN] { pipeline.CamMove, pipeline.CamMoveBack = true, 1 } else { pipeline.CamMoveBack = 0 }
		if KeysPressed[sdl.K_LEFT] { pipeline.CamTurn, pipeline.CamTurnLeft = true, 1 } else { pipeline.CamTurnLeft = 0 }
		if KeysPressed[sdl.K_RIGHT] { pipeline.CamTurn, pipeline.CamTurnRight = true, 1 } else { pipeline.CamTurnRight = 0 }
		if KeysPressed[sdl.K_PAGEUP] { pipeline.CamTurn, pipeline.CamTurnUp = true, 1 } else { pipeline.CamTurnUp = 0 }
		if KeysPressed[sdl.K_PAGEDOWN] { pipeline.CamTurn, pipeline.CamTurnDown = true, 1 } else { pipeline.CamTurnDown = 0 }
		// if KeysPressed[sdl.K_RSHIFT] { pipeline.SpeedUp = 10 } else { pipeline.SpeedUp = 1 }
		if KeysPressed[sdl.K_a] { pipeline.CamMove, pipeline.CamMoveLeft = true, 1 } else { pipeline.CamMoveLeft = 0 }
		if KeysPressed[sdl.K_d] { pipeline.CamMove, pipeline.CamMoveRight = true, 1 } else { pipeline.CamMoveRight = 0 }
		if KeysPressed[sdl.K_w] { pipeline.CamMove, pipeline.CamMoveUp = true, 1 } else { pipeline.CamMoveUp = 0 }
		if KeysPressed[sdl.K_s] { pipeline.CamMove, pipeline.CamMoveDown = true, 1 } else { pipeline.CamMoveDown = 0 }
		if durSec.Seconds() < 1 { Fps++ } else { RefreshWindowCaption(); Fps = 0; lastSecond = nowTime }

		pipeline.PreRender()
		// gl.Disable(gl.FRAMEBUFFER_SRGB)
		Canvas0.RenderContent()
		// Canvas0.RenderPostFx()
		// gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		// gl.Viewport(0, 0, gl.Sizei(Screen.W), gl.Sizei(Screen.H))
		// Canvas0.RenderSelf()

		sdl.GL_SwapBuffers()
		glutil.PanicIfErrors("Post Render Loop")
	}
}

func RecompileShaders (shaderNames ... string) {
	var vs string
	var glProg gl.Uint
	var glStatus gl.Int
	var glFShader, glVShader gl.Uint
	var shaderProg *gltypes.ShaderProgram
	var defines = map[string]interface{} { }
	defines["CANV_W"] = Screen.W
	defines["CANV_H"] = Screen.H
	for fsName, fsSrc := range glsl.FShaders {
		if (len(shaderNames) == 0) || stringutil.IsInSlice(shaderNames, fsName) {
			glFShader = gl.CreateShader(gl.FRAGMENT_SHADER)
			glutil.ShaderSource(fsName, glFShader, fsSrc, defines, false)
			gl.CompileShader(glFShader)
			if gl.GetShaderiv(glFShader, gl.COMPILE_STATUS, &glStatus); glStatus == 0 { log.Panicf("fshader %s: %s\n", fsName, glutil.ShaderInfoLog(glFShader, true)) }

			if vs = glsl.VShaders[fsName]; len(vs) <= 0 {
				vs = glsl.VShaders["ppquad"]
			}
			glVShader = gl.CreateShader(gl.VERTEX_SHADER)
			glutil.ShaderSource(fsName, glVShader, vs, defines, false)
			gl.CompileShader(glVShader)
			if gl.GetShaderiv(glVShader, gl.COMPILE_STATUS, &glStatus); glStatus == 0 { log.Panicf("vshader %s: %s\n", fsName, glutil.ShaderInfoLog(glVShader, true)) }

			glProg = gl.CreateProgram()
			gl.AttachShader(glProg, glFShader)
			if glVShader != 0 { gl.AttachShader(glProg, glVShader) }
			gl.LinkProgram(glProg)
			if gl.GetProgramiv(glProg, gl.LINK_STATUS, &glStatus); glStatus == 0 { log.Panicf("sprog %s: %s", fsName, glutil.ShaderInfoLog(glProg, false)) }

			shaderProg = gltypes.NewShaderProgram(fsName, glProg, glFShader, glVShader)
			shaderProg.UnifCamLook = glutil.ShaderLocationU(glProg, "uCamLook")
			shaderProg.UnifCamPos = glutil.ShaderLocationU(glProg, "uCamPos")
			shaderProg.UnifScreen = glutil.ShaderLocationU(glProg, "uScreen")
			shaderProg.UnifTex0 = glutil.ShaderLocationU(glProg, "uTex0")
			shaderProg.UnifTex1 = glutil.ShaderLocationU(glProg, "uTex1")
			shaderProg.UnifTex2 = glutil.ShaderLocationU(glProg, "uTex2")
			shaderProg.UnifTime = glutil.ShaderLocationU(glProg, "uTime")
			if fsName == "canvas" {
				if Shaders.Canvas != nil { Shaders.Canvas.CleanUp() }
				Shaders.Canvas = shaderProg
			} else if fsName == "postfx" {
				if Shaders.PostFx != nil { Shaders.PostFx.CleanUp() }
				Shaders.PostFx = shaderProg
			} else if fsName == "ppblur" {
				if Shaders.PostBlur != nil { Shaders.PostBlur.CleanUp() }
				Shaders.PostBlur = shaderProg
			} else if fsName == "ppbright" {
				if Shaders.PostBright != nil { Shaders.PostBright.CleanUp() }
				Shaders.PostBright = shaderProg
			} else if fsName == "pplum2" {
				if Shaders.PostLum2 != nil { Shaders.PostLum2.CleanUp() }
				Shaders.PostLum2 = shaderProg
			} else if fsName == "pplum3" {
				if Shaders.PostLum3 != nil { Shaders.PostLum3.CleanUp() }
				Shaders.PostLum3 = shaderProg
			} else if fsName == "texture" {
				if Shaders.Texture != nil { Shaders.Texture.CleanUp() }
				Shaders.Texture = shaderProg
			} else if fsName == "cast" {
				if Shaders.Cast != nil { Shaders.Cast.CleanUp() }
				Shaders.Cast = shaderProg
			}
		}
	}
}

func RefreshWindowCaption () {
	SetWindowCaption(fmt.Sprintf("%v×%v @ %vfps %vkmh P%v R%v", Canvas0.Width, Canvas0.Height, Fps, pipeline.SpeedKmh(), pipeline.CamPos.ToInts(),  pipeline.CamRot.ToInts()))
	// log.Printf("%v FPS", Fps)
}

func ReinitVideo (width, height int, reinitScreen, resizeCanvas bool) {
	var doFullScreen = reinitScreen && resizeCanvas && (width == 0) && (height == 0)
	var prevWidth, prevHeight int
	var oldScreen, newScreen *sdl.Surface
	if oldScreen = Screen; oldScreen != nil {
		prevWidth, prevHeight = int(oldScreen.W), int(oldScreen.H)
	} else {
		prevWidth, prevHeight = SCREEN_DEFWIDTH, SCREEN_DEFHEIGHT
	}
	if reinitScreen {
		if newScreen = sdl.SetVideoMode(width, height, 24, sdl.OPENGL | coreutil.Ifui(doFullScreen, sdl.FULLSCREEN, sdl.RESIZABLE)); newScreen == nil {
			if doFullScreen {
				ReinitVideo(prevWidth, prevHeight, true,  resizeCanvas)
				return
			} else {
				log.Panic("Could not initialize OpenGL surface.")
			}
		} else {
			Screen = newScreen
			if oldScreen != nil { oldScreen.Free() }
		}
	}
	if resizeCanvas {
		Canvas0.Resize(int(float64(width) / Canvas0.RetroFactor), int(float64(height) / Canvas0.RetroFactor))
	}
}

func RemoveCanvas (index int) {
	var canvas = Canvases[index]
	Canvases = append(Canvases[0:index], Canvases[index + 1:]...)
	canvas.CleanUp(true)
}

func SetWindowCaption (caption string) {
	sdl.WM_SetCaption(caption, "")
}

func ToggleFullscreen () {
	// sdl.WM_ToggleFullScreen(glutil.Screen)
	ReinitVideo(coreutil.Ifi(FullScreen, SCREEN_DEFWIDTH, 0), coreutil.Ifi(FullScreen, SCREEN_DEFHEIGHT, 0), true, true)
}

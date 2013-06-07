package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	gl "github.com/chsc/gogl/gl42"
	"github.com/paul-lalonde/Go-SDL/sdl"
)

type shaderProg struct {
	srcFrag, srcVert string
	glProg, glFragShader, glVertShader gl.Uint
	attrPos gl.Uint
	unifTex gl.Int
}

var (
	glTexClamp, glTexFilter gl.Int
	sdlScreen *sdl.Surface
	shaderTextureCreator, shaderTextureDisplay shaderProg
	doRtt = true
	noRtt = false
	volTex gl.Uint
	rttFrameBuf, rttFrameTex gl.Uint
	rttVertBuf, dispVertBuf gl.Uint
	rttVerts, dispVerts []gl.Float
	startTime, lastLoopTime, lastSecond time.Time
	texWidth, texHeight, winWidth, winHeight gl.Sizei
)

func init () {
	var x, y, z gl.Float
	texWidth, texHeight = 640, 360
	glTexClamp, glTexFilter = gl.CLAMP_TO_BORDER, gl.NEAREST
	shaderTextureDisplay.srcVert = "uniform sampler2D uTex; varying vec2 vPos; attribute vec3 aPos; void main (void) { vPos = (aPos + 1) / 2; gl_Position = vec4(aPos, 1.0); }"
	shaderTextureDisplay.srcFrag = "uniform sampler2D uTex; varying vec2 vPos; void main (void) { gl_FragColor = texture2D(uTex, vPos); }"
	shaderTextureCreator.srcVert = "varying vec2 vPos; attribute vec2 aPos; uniform sampler3D uTex; void main (void) { vPos = (aPos + 1) / 2; gl_Position = vec4(aPos, 0, 1.0); }"
	shaderTextureCreator.srcFrag = `
		uniform sampler3D uTex;
		// uniform sampler2D uTexEnd;
		// uniform sampler2D uTexStart;
		// uniform sampler1D uTexTransferFunc;
		varying vec2 vPos;

		float stepSize = 0.0123; // 0.001 

		void main (void) {
			// Get the end point of the ray (from the front-culled faces rendering)
			vec3 rayStart = vec3(vPos, 1); // .xyz;
			vec3 rayEnd = vec3(vPos, 0); // texture2D(uTexEnd, vPos).xyz;

			// Get a vector from back to front
			vec3 traverseVector = rayEnd - rayStart;

			// The maximum length of the ray
			float maxLength = length(traverseVector);

			// Construct a ray in the correct direction and of step length
			vec3 step = stepSize * normalize(traverseVector);
			vec3 rayStep = step;

			// The color accumulation buffer
			vec4 acc = vec4(0.0, 0.0, 0.0, 0.0);

			// Holds current voxel color
			vec4 voxelColor;

			vec4 texel;

			// Advance ray
			for (int i = 0; i < int(1 / stepSize); ++i) {
				if ((length(rayStep) >= maxLength) || (acc.a >= 0.99)) {
					acc.a = 1.0;
					break;
				}

				texel = texture3D(uTex, rayStep + rayStart);
				voxelColor = vec4(texel.w, texel.w / 1.5, texel.w / 2, texel.w); // texture1D(uTexTransferFunc, texture3D(uTex, rayStep + rayStart).w);

				// Accumulate RGB : acc.rgb = (voxelColor.rgb * voxelColor.a) + ((1.0 - voxelColor.a) * acc.rgb);
				acc.rgb = mix(acc.rgb, voxelColor.rgb, voxelColor.a);

				// Accumulate Opacity: acc.a = acc.a + (1.0 - acc.a)*voxelColor.a;
				acc.a = mix(voxelColor.a, 1.0, acc.a);

				rayStep += step;
			}
			gl_FragColor = acc; // vec4(1, acc.g, 0, 1);
			return;
		}`
	x, y, z = 1, 1, 1
	rttVerts = []gl.Float { -x, -y, x, -y, x, y, -x, y }
	dispVerts = []gl.Float { -x, -y, z, x, -y, z, x, y, z, -x, y, z }
}

// 256x256x178 http://www.volvis.org/

func main () {
	var now = time.Now()
	initialWidth, initialHeight := 1280, 720
	runtime.LockOSThread()
	if sdlInit := sdl.Init(sdl.INIT_VIDEO); sdlInit != 0 { panic("SDL init error") }
	reinitScreen(initialWidth, initialHeight)
	defer cleanExit(true, false)
	if err := gl.Init(); err != nil { panic(err) }
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.DEPTH)
	sdl.WM_SetCaption("Loading volume...", "")
	loadVolume()
	defer cleanExit(false, true)
	sdl.WM_SetCaption("Compiling shaders...", "")
	glSetupShaderProg(&shaderTextureCreator)
	glSetupShaderProg(&shaderTextureDisplay)
	glFillBuffer(rttVerts, &rttVertBuf)
	glFillBuffer(dispVerts, &dispVertBuf)
	gl.GenTextures(1, &rttFrameTex)
	gl.BindTexture(gl.TEXTURE_2D, rttFrameTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, glTexFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, glTexFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, glTexClamp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, glTexClamp)
	if doRtt && !noRtt {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, texWidth, texHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
		gl.GenFramebuffers(1, &rttFrameBuf)
		gl.BindFramebuffer(gl.FRAMEBUFFER, rttFrameBuf)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, rttFrameTex, 0)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	} else {
		if noRtt {
			rttFrameBuf = 0
		} else {
			glFillTextureFromImageFile("texture.jpg")
		}
	}
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.ClearColor(0.3, 0.2, 0.1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	var looping = true
	var fps = 0
	var durSec time.Duration
	startTime, lastLoopTime, lastSecond = now, now, now
	sdl.WM_SetCaption("Up & running!", "")
	for looping {
		now = time.Now()
		if durSec = now.Sub(lastSecond); durSec.Seconds() >= 1 {
			sdl.WM_SetCaption(fmt.Sprintf("%v×%v @ %vfps", sdlScreen.W, sdlScreen.H, fps), "")
			fps = 0
			lastSecond = now
		} else {
			fps++
		}
		if evt := sdl.PollEvent(); evt != nil {
			switch event := evt.(type) {
				case *sdl.ResizeEvent:
					reinitScreen(int(event.W), int(event.H))
				case *sdl.QuitEvent:
					looping = false
			}
		}
		if doRtt || noRtt { renderToTexture() }
		if !noRtt { renderToScreen() }
		sdl.GL_SwapBuffers()
		lastLoopTime = now
	}
	sdl.Quit()
}

func cleanExit (screen, gfx bool) {
	if gfx {
		glDisposeShaderProg(shaderTextureDisplay)
		glDisposeShaderProg(shaderTextureCreator)
		gl.DeleteFramebuffers(1, &rttFrameBuf)
		gl.DeleteTextures(1, &rttFrameTex)
		gl.DeleteTextures(1, &volTex)
		gl.DeleteBuffers(1, &rttVertBuf)
		gl.DeleteBuffers(1, &dispVertBuf)
	}
	if screen && (sdlScreen != nil) {
		sdlScreen.Free()
		sdlScreen = nil
	}
}

func glDisposeShaderProg (sp shaderProg) {
	gl.DetachShader(sp.glProg, sp.glFragShader)
	gl.DetachShader(sp.glProg, sp.glVertShader)
	gl.DeleteShader(sp.glFragShader)
	gl.DeleteShader(sp.glVertShader)
	gl.DeleteProgram(sp.glProg)
}

func glFillBuffer (vals []gl.Float, buf *gl.Uint) {
	gl.GenBuffers(1, buf)
	gl.BindBuffer(gl.ARRAY_BUFFER, *buf)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(vals) * 4), gl.Pointer(&vals[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

func glFillTextureFromImageFile (filePath string) {
	var file, err = os.Open(filePath)
	var img image.Image
	if err != nil { panic(err) }
	defer file.Close()
	if img, _, err = image.Decode(file); err != nil { panic(err) }
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ { for y := 0; y < h; y++ { rgba.Set(x, y, img.At(x, y)) } }
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.Sizei(w), gl.Sizei(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Pointer(&rgba.Pix[0]))
}

func glGetInfoLog (shaderOrProgram gl.Uint, isShader bool) string {
	var l = gl.Sizei(256)
	var s = gl.GLStringAlloc(l)
	defer gl.GLStringFree(s)
	if isShader { gl.GetShaderInfoLog(shaderOrProgram, l, nil, s) } else { gl.GetProgramInfoLog(shaderOrProgram, l, nil, s) }
	return gl.GoString(s)
}

func glGetLocation (glProg gl.Uint, name string, isAtt bool) gl.Uint {
	var a gl.Int
	var s = gl.GLString(name)
	defer gl.GLStringFree(s)
	if isAtt { a = gl.GetAttribLocation(glProg, s) } else { a = gl.GetUniformLocation(glProg, s) }
	if a < 0 { panic("Shader attribute or uniform bind error") }
	return gl.Uint(a)
}

func glSetShaderSource (shader gl.Uint, source string) {
	var src = gl.GLStringArray(source)
	defer gl.GLStringArrayFree(src)
	gl.ShaderSource(shader, gl.Sizei(len(src)), &src[0], nil)
}

func glSetupShaderProg (shader *shaderProg) {
	var glStatus gl.Int
	shader.glFragShader = gl.CreateShader(gl.FRAGMENT_SHADER)
	glSetShaderSource(shader.glFragShader, shader.srcFrag)
	gl.CompileShader(shader.glFragShader)
	if gl.GetShaderiv(shader.glFragShader, gl.COMPILE_STATUS, &glStatus); glStatus == 0 { panic("Frag shader compile error: " + glGetInfoLog(shader.glFragShader, true)) }
	shader.glVertShader = gl.CreateShader(gl.VERTEX_SHADER)
	glSetShaderSource(shader.glVertShader, shader.srcVert)
	gl.CompileShader(shader.glVertShader)
	if gl.GetShaderiv(shader.glVertShader, gl.COMPILE_STATUS, &glStatus); glStatus == 0 { panic("Vert shader compile error: " + glGetInfoLog(shader.glFragShader, true)) }
	shader.glProg = gl.CreateProgram()
	gl.AttachShader(shader.glProg, shader.glFragShader)
	gl.AttachShader(shader.glProg, shader.glVertShader)
	gl.LinkProgram(shader.glProg)
	if gl.GetProgramiv(shader.glProg, gl.LINK_STATUS, &glStatus); glStatus == 0 { panic("Shader program link error: " + glGetInfoLog(shader.glProg, false)) }
	shader.attrPos = glGetLocation(shader.glProg, "aPos", true)
	shader.unifTex = gl.Int(glGetLocation(shader.glProg, "uTex", false))
}

func loadVolume () {
	var file, err = os.Open("/home/roxor/apps/ImageVis3D/Data/BostonTeapot.raw")
	var bytes []byte
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if bytes, err = ioutil.ReadAll(file); err != nil {
		panic(err)
	}
	gl.GenTextures(1, &volTex)
	gl.BindTexture(gl.TEXTURE_3D, volTex)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, glTexFilter)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, glTexFilter)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_R, glTexClamp)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_S, glTexClamp)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_WRAP_T, glTexClamp)
	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.ALPHA, 256, 256, 178, 0, gl.ALPHA, gl.UNSIGNED_BYTE, gl.Pointer(&bytes[0]))
	gl.BindTexture(gl.TEXTURE_3D, 0)
}

func reinitScreen (width, height int) {
	winWidth, winHeight = gl.Sizei(width), gl.Sizei(height)
	if sdlScreen != nil { sdlScreen.Free() }
	sdlScreen = sdl.SetVideoMode(width, height, 24, sdl.OPENGL | sdl.RESIZABLE)
	if sdlScreen == nil { panic("SDL video init error") }
}

func renderToScreen () {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, winWidth, winHeight)
	gl.ClearColor(0.1, 0.3, 0.2, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(shaderTextureDisplay.glProg)
	gl.BindTexture(gl.TEXTURE_2D, rttFrameTex)
	gl.Uniform1i(shaderTextureDisplay.unifTex, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, dispVertBuf)
	gl.EnableVertexAttribArray(shaderTextureDisplay.attrPos)
	gl.VertexAttribPointer(shaderTextureDisplay.attrPos, 3, gl.FLOAT, gl.FALSE, 0, gl.Pointer(nil))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(shaderTextureDisplay.attrPos)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

func renderToTexture () {
	gl.BindFramebuffer(gl.FRAMEBUFFER, rttFrameBuf)
	gl.Viewport(0, 0, texWidth, texHeight)
	gl.ClearColor(0.1, 0.2, 0.3, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(shaderTextureCreator.glProg)
	gl.BindTexture(gl.TEXTURE_3D, volTex)
	gl.Uniform1i(shaderTextureCreator.unifTex, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, rttVertBuf)
	gl.EnableVertexAttribArray(shaderTextureCreator.attrPos)
	gl.VertexAttribPointer(shaderTextureCreator.attrPos, 2, gl.FLOAT, gl.FALSE, 0, gl.Pointer(nil))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(shaderTextureCreator.attrPos)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

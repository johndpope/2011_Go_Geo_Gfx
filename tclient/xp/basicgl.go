package main

import (
	gl "github.com/chsc/gogl/gl42"
	sdl "github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"

	"image"
	_ "image/jpeg"
	"os"
	"runtime"
)

type shaderProg struct {
	srcFrag, srcVert string
	glProg, glFragShader, glVertShader gl.Uint
	attrPos gl.Uint
	unifTex gl.Int
}

var (
	sdlScreen *sdl.Surface
	shaderTextureCreator, shaderTextureDisplay shaderProg
	doRtt = true
	rttFrameBuf, rttFrameTex gl.Uint
	rttVertBuf, dispVertBuf gl.Uint
	rttVerts, dispVerts []gl.Float
	texSize, winWidth, winHeight gl.Sizei
)

func init () {
	var x, y, z gl.Float
	texSize = 8
	shaderTextureCreator.srcFrag = "varying vec2 vPos; void main (void) { gl_FragColor = vec4(0.25, vPos, 1); }"
	shaderTextureCreator.srcVert = "varying vec2 vPos; attribute vec2 aPos; void main (void) { vPos = (aPos + 1) / 2; gl_Position = vec4(aPos, 0, 1.0); }"
	shaderTextureDisplay.srcFrag = "uniform sampler2D uTex; varying vec2 vPos; void main (void) { gl_FragColor = texture2D(uTex, vPos); }"
	shaderTextureDisplay.srcVert = "uniform sampler2D uTex; varying vec2 vPos; attribute vec3 aPos; void main (void) { vPos = (aPos + 1); gl_Position = vec4(aPos, 1.0); }"
	x, y, z = 1, 1, 1
	rttVerts = []gl.Float { -x, -y, x, -y, x, y, -x, y }
	dispVerts = []gl.Float { -x, -y, z, x, -y, z, x, y, z, -x, y, z }
}

func main () {
	initialWidth, initialHeight := 1280, 720
	runtime.LockOSThread()
	if sdlInit := sdl.Init(sdl.INIT_VIDEO); sdlInit != 0 { panic("SDL init error") }
	reinitScreen(initialWidth, initialHeight)
	defer cleanExit(true, false)
	if err := gl.Init(); err != nil { panic(err) }
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.DEPTH)
	defer cleanExit(false, true)
	glSetupShaderProg(&shaderTextureCreator, false)
	glSetupShaderProg(&shaderTextureDisplay, true)
	glFillBuffer(rttVerts, &rttVertBuf)
	glFillBuffer(dispVerts, &dispVertBuf)
	gl.GenTextures(1, &rttFrameTex)
	gl.BindTexture(gl.TEXTURE_2D, rttFrameTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	if doRtt {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, texSize, texSize, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
		gl.GenFramebuffers(1, &rttFrameBuf)
		gl.BindFramebuffer(gl.FRAMEBUFFER, rttFrameBuf)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, rttFrameTex, 0)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	} else {
		glFillTextureFromImageFile("texture.jpg")
	}
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.ClearColor(0.3, 0.6, 0.9, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ActiveTexture(gl.TEXTURE0)
	for {
		if evt := sdl.PollEvent(); evt != nil {
			switch event := evt.(type) {
				case *sdl.ResizeEvent:
					reinitScreen(int(event.W), int(event.H))
				case *sdl.QuitEvent:
					return
			}
		} else {
			if doRtt { renderToTexture() }
			renderToScreen()
			sdl.GL_SwapBuffers()
		}
	}

	sdl.Quit()
}

func renderToScreen () {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, winWidth, winHeight)
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
	gl.Viewport(0, 0, texSize, texSize)
	gl.ClearColor(0.9, 0.6, 0.3, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(shaderTextureCreator.glProg)
	gl.BindBuffer(gl.ARRAY_BUFFER, rttVertBuf)
	gl.EnableVertexAttribArray(shaderTextureCreator.attrPos)
	gl.VertexAttribPointer(shaderTextureCreator.attrPos, 2, gl.FLOAT, gl.FALSE, 0, gl.Pointer(nil))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(shaderTextureCreator.attrPos)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

func cleanExit (screen, gfx bool) {
	if gfx {
		glDisposeShaderProg(shaderTextureDisplay)
		glDisposeShaderProg(shaderTextureCreator)
		gl.DeleteFramebuffers(1, &rttFrameBuf)
		gl.DeleteTextures(1, &rttFrameTex)
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

func glSetupShaderProg (shader *shaderProg, unif bool) {
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
	if unif { shader.unifTex = gl.Int(glGetLocation(shader.glProg, "uTex", false)) }
}

func reinitScreen (width, height int) {
	winWidth, winHeight = gl.Sizei(width), gl.Sizei(height)
	if sdlScreen != nil { sdlScreen.Free() }
	sdlScreen = sdl.SetVideoMode(width, height, 24, sdl.OPENGL | sdl.RESIZABLE)
	if sdlScreen == nil { panic("SDL video init error") }
}

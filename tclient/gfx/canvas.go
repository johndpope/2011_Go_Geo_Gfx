package canvas

import (
	"log"
	"math"

	gl "github.com/chsc/gogl/gl42"

	"tclient/gfx/gltypes"
	"tclient/gfx/glutil"
	"tclient/gfx/pipeline"
)

type Canvas struct {
	MouseX, MouseY gl.Float
	Index gl.Int
	FrameBufContent, FrameTexContent, FrameBufPost, FrameTexPost, FrameBufLum2, FrameTexLum2, FrameBufBlurH, FrameTexBlurH, FrameBufBlurV, FrameTexBlurV, FrameBufBright, FrameTexBright gl.Uint
	FrameBufLum3, FrameTexLum3 []gl.Uint
	Width, Height gl.Sizei
	InvWidth, InvHeight float64
	Shaders *gltypes.ShaderManager
	RetroFactor float64
	TmpTex gl.Uint
	TimeElapsed int64

	lumMeasuring bool
	i, j, it, blurIterations, brightIterations int
	width, height, realWidth, realHeight, bloomWidth, bloomHeight gl.Float
	tmpFloats []gl.Float
	bloomSizeX, bloomSizeY gl.Sizei
	fw, fh float64
	lumChainLen int
	lumChainSize, lumChainMaxSize float64
	blurQuality float64
	frameBufBlurV, frameTexBlurV, lastTex, nextTex gl.Uint
}

func (me *Canvas) CleanUp (dispose bool) {
	gl.DeleteFramebuffers(1, &me.FrameBufContent)
	gl.DeleteFramebuffers(1, &me.FrameBufPost)
	gl.DeleteFramebuffers(1, &me.FrameBufLum2)
	gl.DeleteFramebuffers(1, &me.FrameBufBright)
	gl.DeleteFramebuffers(1, &me.FrameBufBlurH)
	gl.DeleteFramebuffers(1, &me.FrameBufBlurV)
	if len(me.FrameBufLum3) > 0 { gl.DeleteFramebuffers(gl.Sizei(len(me.FrameBufLum3)), &me.FrameBufLum3[0]) }
	gl.DeleteTextures(1, &me.FrameTexContent)
	gl.DeleteTextures(1, &me.FrameTexPost)
	gl.DeleteTextures(1, &me.FrameTexLum2)
	gl.DeleteTextures(1, &me.FrameTexBright)
	gl.DeleteTextures(1, &me.FrameTexBlurH)
	gl.DeleteTextures(1, &me.FrameTexBlurV)
	if len(me.FrameTexLum3) > 0 { gl.DeleteTextures(gl.Sizei(len(me.FrameTexLum3)), &me.FrameTexLum3[0]) }
}

func (me *Canvas) Init (width, height int, retroFactor float64) {
	me.tmpFloats = make([]gl.Float, 4)
	me.RetroFactor = retroFactor
	me.blurQuality = 0.25
	me.blurIterations = 0
	me.brightIterations = 0
	me.Resize(width, height)
}

func (me *Canvas) RenderContent () {
	// 	me.RenderContentInSoftware()
	// 	return
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0/*me.FrameBufContent*/)
	gl.Viewport(0, 0, me.Width, me.Height)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(me.Shaders.Cast.Program)
	// gl.Uniform1f(me.Shaders.Cast.UnifTime, gl.Float(float64(me.TimeElapsed) / 1000 / 1000000))
	gl.Uniform2f(me.Shaders.Cast.UnifScreen, me.width, me.height)
	gl.Uniform3f(me.Shaders.Cast.UnifCamLook, gl.Float(pipeline.CamLook.X), gl.Float(pipeline.CamLook.Y), gl.Float(pipeline.CamLook.Z))
	gl.Uniform3f(me.Shaders.Cast.UnifCamPos, gl.Float(pipeline.CamPos.X), gl.Float(pipeline.CamPos.Y), gl.Float(pipeline.CamPos.Z))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
}

func (me *Canvas) RenderContentTexture () {
	gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufContent)
	gl.Viewport(0, 0, me.Width, me.Height)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(me.Shaders.Texture.Program)
	gl.BindTexture(gl.TEXTURE_2D, me.TmpTex)
	gl.Uniform2f(me.Shaders.Texture.UnifScreen, me.width, -me.height)
	// gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
}

// func (me *Canvas) RenderContentInSoftware () {
// 	softpipeline.NowTick = me.NowTick
// 	softpipeline.PreRender()
// 	softpipeline.Render()
// 	softpipeline.PostRender()
// 	gl.BindTexture(gl.TEXTURE_2D, me.FrameTexContent)
// 	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, me.Width, me.Height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Pointer(&softwareRenderTarget.Pix[0]))
// 	gl.BindTexture(gl.TEXTURE_2D, 0)
// }

func (me *Canvas) RenderPostFx () {
/*
	gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufLum2)
	gl.Viewport(0,  0, gl.Sizei(me.fw * 0.25), gl.Sizei(me.fh * 0.25))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(me.Shaders.PostLum2.Program)
	gl.BindTexture(gl.TEXTURE_2D, me.FrameTexContent)
	gl.Uniform4f(me.Shaders.PostLum2.UnifScreen, gl.Float(1 / (me.fw * 0.25)), gl.Float(1 / (me.fh * 0.25)), me.width, me.height)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	me.lumChainSize = me.lumChainMaxSize
	gl.UseProgram(me.Shaders.PostLum3.Program)
	gl.BindTexture(gl.TEXTURE_2D, me.FrameTexLum2)
	gl.Uniform4f(me.Shaders.PostLum3.UnifScreen, gl.Float(1 / me.lumChainSize), gl.Float(1 / me.lumChainSize), gl.Float(me.fw * 0.25), gl.Float(me.fh * 0.25))
	for me.i = 0; me.i < me.lumChainLen; me.i++ {
		gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufLum3[me.i])
		gl.Viewport(0, 0, gl.Sizei(me.lumChainSize), gl.Sizei(me.lumChainSize))
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		gl.Uniform4f(me.Shaders.PostLum3.UnifScreen, gl.Float(1 / me.lumChainSize * 3), gl.Float(1 / me.lumChainSize * 3), gl.Float(me.lumChainSize), gl.Float(me.lumChainSize))
		gl.BindTexture(gl.TEXTURE_2D, me.FrameTexLum3[me.i])
		me.lumChainSize /= 3
	}

	if me.TmpTex != me.lastTex {
		me.lastTex = me.TmpTex
		gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufLum3[me.i - 1])
		gl.BindTexture(gl.TEXTURE_2D, me.FrameTexLum3[me.i - 1])
		gl.ReadPixels(0, 0, 1, 1, gl.RGBA, gl.FLOAT, gl.Pointer(&me.tmpFloats[0]))
		fMean := math.Max(0.0001, math.Exp(float64(me.tmpFloats[0])))
		fKey := math.Max(0, 1.5 - (1.5 / ((fMean * 0.1) + 1))) + 0.1;
		fExp := math.Max(0.0001, fKey * (1 / fMean))
		fExp2 := math.Pow(2, fExp)
		log.Printf("Readback: min=%f max=%f avg=%f exp(avg)=%f Key=%f exp=%f exp2=%f", me.tmpFloats[2], me.tmpFloats[1], me.tmpFloats[0], fMean, fKey, fExp, fExp2)
	}

/*
	if me.brightIterations > 0 {
		gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufBright)
		gl.Viewport(0,  0, me.bloomSizeX, me.bloomSizeY)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.UseProgram(me.Shaders.PostBright.Program)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, me.FrameTexLum3[me.i - 1])
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, me.FrameTexContent)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
		me.it, me.nextTex, me.frameBufBlurV, me.frameTexBlurV = me.brightIterations, me.FrameTexBright, me.FrameBufBright, me.FrameTexBright
		me.RenderPostFxBlur()
		me.nextTex = me.FrameTexBright
	}
	if me.blurIterations > 0 {
		me.it, me.nextTex, me.frameBufBlurV, me.frameTexBlurV = me.blurIterations, me.FrameTexContent, me.FrameBufBlurV, me.FrameTexBlurV
		me.RenderPostFxBlur()
	}
*/
	// gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufBright)
	// gl.Viewport(0, 0, me.bloomSizeX, me.bloomSizeY)
	// gl.Clear(gl.COLOR_BUFFER_BIT)
	// gl.UseProgram(me.Shaders.PostBright.Program)
	// gl.BindTexture(gl.TEXTURE_2D, me.FrameTexContent)
	// gl.Uniform2f(me.Shaders.PostBright.UnifScreen, me.bloomWidth, me.bloomHeight)
	// gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	// me.it, me.nextTex, me.frameBufBlurV, me.frameTexBlurV = 1, me.FrameTexBright, me.FrameBufBright, me.FrameTexBright
	// me.RenderPostFxBlur()

	gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufPost)
	gl.Viewport(0, 0, me.Width, me.Height)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(me.Shaders.PostFx.Program)
	gl.Uniform2f(me.Shaders.PostFx.UnifScreen, me.width, me.height)
	// gl.ActiveTexture(gl.TEXTURE1)
	// gl.BindTexture(gl.TEXTURE_2D, me.FrameTexLum3[me.i - 1])
	// gl.Uniform1i(me.Shaders.PostFx.UnifTex1, 1)
	// gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, me.FrameTexContent)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
}

func (me *Canvas) RenderPostFxBlur () {
	gl.UseProgram(me.Shaders.PostBlur.Program)
	for me.j = 0; me.j < me.it; me.j++ {
		gl.BindFramebuffer(gl.FRAMEBUFFER, me.FrameBufBlurH)
		gl.Viewport(0,  0, me.bloomSizeX, me.bloomSizeY)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.BindTexture(gl.TEXTURE_2D, me.nextTex)
		gl.Uniform2f(me.Shaders.PostBlur.UnifScreen, me.bloomWidth, 0)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

		gl.BindFramebuffer(gl.FRAMEBUFFER, me.frameBufBlurV)
		gl.Viewport(0,  0, me.bloomSizeX, me.bloomSizeY)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.BindTexture(gl.TEXTURE_2D, me.FrameTexBlurH)
		gl.Uniform2f(me.Shaders.PostBlur.UnifScreen, 0, me.bloomHeight)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
		me.nextTex = me.frameTexBlurV
	}
}

func (me *Canvas) RenderSelf () {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(me.Shaders.Canvas.Program)
	gl.Uniform2f(me.Shaders.Canvas.UnifScreen, me.realWidth, me.realHeight)
	gl.BindTexture(gl.TEXTURE_2D, /*me.FrameTexPost*/me.FrameTexContent)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
}

func (me *Canvas) Resize (newWidth, newHeight int) {
	var lumChainStartSize float64
	if (newWidth != 0) && (newHeight != 0) {
		log.Printf("\n")
		me.fw, me.fh = float64(newWidth), float64(newHeight)
		me.InvWidth, me.InvHeight = 1 / me.fw, 1 / me.fh
		me.width, me.height = gl.Float(me.InvWidth), gl.Float(me.InvHeight)
		me.realWidth, me.realHeight = gl.Float(1 / (me.fw * me.RetroFactor)), gl.Float(1 / (me.fh * me.RetroFactor))
		me.MouseX, me.MouseY = 0.5, 0.5
		me.Width, me.Height = gl.Sizei(newWidth), gl.Sizei(newHeight)
		me.CleanUp(false)
		me.bloomWidth, me.bloomHeight = gl.Float(1 / (me.fw * me.blurQuality)), gl.Float(1 / (me.fh * me.blurQuality))
		me.bloomSizeX, me.bloomSizeY = gl.Sizei(me.fw * me.blurQuality), gl.Sizei(me.fh * me.blurQuality)
		me.FrameBufContent, me.FrameTexContent = glutil.MakeRttFramebuffer(gl.RGB16F, me.Width, me.Height, 1, 0)
		me.FrameBufBright, me.FrameTexBright = glutil.MakeRttFramebuffer(gl.RGB16F, me.bloomSizeX, me.bloomSizeY, 1, 1)
		me.FrameBufBlurH, me.FrameTexBlurH = glutil.MakeRttFramebuffer(gl.RGB16F, me.bloomSizeX, me.bloomSizeY, 1, 1)
		me.FrameBufBlurV, me.FrameTexBlurV = glutil.MakeRttFramebuffer(gl.RGB16F, me.bloomSizeX, me.bloomSizeY, 1, 1)
		me.FrameBufLum2, me.FrameTexLum2 = glutil.MakeRttFramebuffer(gl.RGB16F, gl.Sizei(me.fw / 4), gl.Sizei(me.fh / 4), 1, 0)
		me.FrameBufPost, me.FrameTexPost = glutil.MakeRttFramebuffer(gl.RGB8, me.Width, me.Height, 1, 1)
		me.lumChainLen, lumChainStartSize = 0, math.Min(me.fw, me.fh) / 4
		for me.lumChainSize = 1; me.lumChainSize < lumChainStartSize; me.lumChainSize *= 3 { me.lumChainLen++ }
		me.lumChainSize /= 3
		me.lumChainMaxSize = me.lumChainSize
		me.FrameBufLum3, me.FrameTexLum3 = make([]gl.Uint, me.lumChainLen), make([]gl.Uint, me.lumChainLen)
		for i := 0; i < me.lumChainLen; i++ {
			me.FrameBufLum3[i], me.FrameTexLum3[i] = glutil.MakeRttFramebuffer(gl.RGB16F, gl.Sizei(me.lumChainSize), gl.Sizei(me.lumChainSize), 1, 0)
			me.lumChainSize /= 3
		}
		gl.ClearColor(0.85, 0.5, 0.5, 1)
		// softwareRenderTarget = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		// softpipeline.Reinit(newWidth, newHeight, softwareRenderTarget)
	}
}

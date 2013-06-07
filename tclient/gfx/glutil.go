package glutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"

	gl "github.com/chsc/gogl/gl42"

	"tclient/gfx/gltypes"
	"tshared/stringutil"
	num "tshared/numutil"
)

const (
	KB = 1024
	MB = KB * KB
	GB = MB * KB
	GPU_MEMORY_INFO_DEDICATED_VIDMEM_NVX gl.Enum = 0x9047
	GPU_MEMORY_INFO_TOTAL_AVAILABLE_MEMORY_NVX gl.Enum = 0x9048
	GPU_MEMORY_INFO_CURRENT_AVAILABLE_VIDMEM_NVX gl.Enum = 0x9049
	GPU_MEMORY_INFO_EVICTION_COUNT_NVX gl.Enum = 0x904A
	GPU_MEMORY_INFO_EVICTED_MEMORY_NVX gl.Enum = 0x904B
	RENDERBUFFER_FREE_MEMORY_ATI gl.Enum = 0x87FD
	TEXTURE_FREE_MEMORY_ATI gl.Enum = 0x87FC
	VBO_FREE_MEMORY_ATI gl.Enum = 0x87FB
)

var (
	BufMode gl.Enum = gl.STATIC_DRAW
	FillWithZeroes = false
	QuadVerts2D = []gl.Float { -1, -1, 1, -1, 1, 1, -1, 1 }
	TexFilter gl.Int = gl.NEAREST
	TexWrap gl.Int = gl.CLAMP_TO_BORDER
)

var (
	extensionPrefixes = []string { "GL_ARB_", "GL_ATI_", "GL_S3_", "GL_EXT_", "GL_IBM_", "GL_KTX_", "GL_NV_", "GL_NVX_", "GL_OES_", "GL_SGIS_", "GL_SGIX_", "GL_SUN_", "GL_APPLE_" }
	extensions []string = nil
	maxTexBufSize, maxTexSize1D, maxTexSize2D, maxTexSize3D gl.Int = 0, 0, 0, 0
)

func Extension (name string) bool {
	if strings.HasPrefix(name, "GL_") { return stringutil.InSliceAt(Extensions(), name) >= 0 }
	for _, ep := range extensionPrefixes { if stringutil.InSliceAt(Extensions(), ep + name) >= 0 { return true } }
	return false
}

func Extensions () []string {
	var ub *gl.Ubyte
	if extensions == nil {
		ub = gl.GetString(gl.EXTENSIONS)
		extensions = stringutil.Split(gl.GoStringUb(ub), " ")
	}
	return extensions
}

func FindTexSize2D (maxSize, numTexels, minSize float64) (float64, float64) {
	var wh float64
	if math.Floor(numTexels) != numTexels { log.Panicf("AAAAH %v", numTexels) }
	if numTexels <= maxSize { return numTexels, 1 }
	for h := 2.0; h < maxSize; h ++ {
		for w := 2.0; w < maxSize; w ++ {
			wh = w * h
			if wh == numTexels {
				if minSize > 0 { for (math.Mod(w, 2) == 0) && (math.Mod(h, 2) == 0) && ((w / 2) >= minSize) && ((h / 2) >= minSize) { w, h = w / 2, h / 2 } }
				for ((h * 2) < w) && (math.Mod(w, 2) == 0) { w, h = w / 2, h * 2 }
				if minSize > 0 { for (math.Mod(w, 2) == 0) && (math.Mod(h, 2) == 0) && ((w / 2) >= minSize) && ((h / 2) >= minSize) { w, h = w / 2, h / 2 } }
				return w, h
			} else if wh > numTexels { break }
		}
	}
	return 0, 0
}

func Integerv (name gl.Enum) gl.Int {
	var ret gl.Int
	gl.GetIntegerv(name, &ret)
	PanicIfErrors("Integerv(n=%v)", name)
	return ret
}

func Integervs (name gl.Enum, num uint) []gl.Int {
	var ret = make([]gl.Int, num)
	gl.GetIntegerv(name, &ret[0])
	PanicIfErrors("Integervs(n=%v)", name)
	return ret
}

func LastErrors (filterBy ... gl.Enum) []gl.Enum {
	var errs []gl.Enum = nil
	var err gl.Enum
	for {
		if err = gl.GetError(); err == 0 {
			break;
		} else if (len(filterBy) == 0) || (gltypes.InSliceAt(filterBy, err) >= 0) {
			if errs == nil { errs = []gl.Enum { err } } else { errs = append(errs, err) }
		}
	}
	return errs
}

func MakeArrayBuffer (glPtr *gl.Uint, size uint64, sl interface{}, isLen, makeTex bool) gl.Uint {
	var ptr = gl.Pointer(nil)
	var glTex gl.Uint = 0
	var glTexFormat gl.Enum = gl.R8UI
	var sizeFactor, sizeTotal uint64 = 1, 0
	var tm = false
	var handle = func (sf uint64, glPtr gl.Pointer, le int, tf gl.Enum) {
		tm = true; if le > 1 { ptr = glPtr }; if size == 0 { size = uint64(le); isLen = true }; if isLen { sizeFactor = sf }; if tf != 0 { glTexFormat = tf }
	}
	if (sl == nil) && FillWithZeroes { sl = make([]uint8, size) }
	gl.GenBuffers(1, glPtr)
	gl.BindBuffer(gl.ARRAY_BUFFER, *glPtr)
	if sl != nil {
		if tv, tb := sl.([]uint8); tb { handle(1, gl.Pointer(&tv[0]), len(tv), gl.R8UI) }
		if tv, tb := sl.([]uint16); tb { handle(2, gl.Pointer(&tv[0]), len(tv), gl.R16UI) }
		if tv, tb := sl.([]uint32); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32UI) }
		if tv, tb := sl.([]uint64); tb { handle(8, gl.Pointer(&tv[0]), len(tv), gl.RG32UI) }
		if tv, tb := sl.([]int8); tb { handle(1, gl.Pointer(&tv[0]), len(tv), gl.R8I) }
		if tv, tb := sl.([]int16); tb { handle(2, gl.Pointer(&tv[0]), len(tv), gl.R16I) }
		if tv, tb := sl.([]int32); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32I) }
		if tv, tb := sl.([]int64); tb { handle(8, gl.Pointer(&tv[0]), len(tv), gl.RG32I) }
		if tv, tb := sl.([]float32); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32F) }
		if tv, tb := sl.([]float64); tb { handle(8, gl.Pointer(&tv[0]), len(tv), gl.RG32F) }
		if tv, tb := sl.([]gl.Bitfield); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R8UI) }
		if tv, tb := sl.([]gl.Byte); tb { handle(1, gl.Pointer(&tv[0]), len(tv), gl.R8I) }
		if tv, tb := sl.([]gl.Ubyte); tb { handle(1, gl.Pointer(&tv[0]), len(tv), gl.R8UI) }
		if tv, tb := sl.([]gl.Ushort); tb { handle(2, gl.Pointer(&tv[0]), len(tv), gl.R16UI) }
		if tv, tb := sl.([]gl.Short); tb { handle(2, gl.Pointer(&tv[0]), len(tv), gl.R16I) }
		if tv, tb := sl.([]gl.Uint); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32UI) }
		if tv, tb := sl.([]gl.Uint64); tb { handle(8, gl.Pointer(&tv[0]), len(tv), gl.RG32UI) }
		if tv, tb := sl.([]gl.Int); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32I) }
		if tv, tb := sl.([]gl.Int64); tb { handle(8, gl.Pointer(&tv[0]), len(tv), gl.RG32I) }
		if tv, tb := sl.([]gl.Clampd); tb { handle(8, gl.Pointer(&tv[0]), len(tv), gl.RG32F) }
		if tv, tb := sl.([]gl.Clampf); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32F) }
		if tv, tb := sl.([]gl.Float); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32F) }
		if tv, tb := sl.([]gl.Half); tb { handle(2, gl.Pointer(&tv[0]), len(tv), gl.R16F) }
		if tv, tb := sl.([]gl.Double); tb { handle(8, gl.Pointer(&tv[0]), len(tv), gl.RG32F) }
		if tv, tb := sl.([]gl.Enum); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32I) }
		if tv, tb := sl.([]gl.Sizei); tb { handle(4, gl.Pointer(&tv[0]), len(tv), gl.R32UI) }
		if tv, tb := sl.([]gl.Char); tb { handle(1, gl.Pointer(&tv[0]), len(tv), gl.R8UI) }
		if !tm { log.Panicf("MakeArrayBuffer() -- slice type unsupported:\n%+v", sl) }
	}
	sizeTotal = size * sizeFactor
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(sizeTotal), ptr, BufMode)
	if makeTex {
		if sizeTotal > MaxTextureBufferSize() { log.Panicf("Texture buffer size (%vMB) would exceed your GPU's maximum texture buffer size (%vMB)", sizeTotal / MB, maxTexBufSize / MB) }
		gl.GenTextures(1, &glTex)
		gl.BindTexture(gl.TEXTURE_BUFFER, glTex)
		gl.TexBuffer(gl.TEXTURE_BUFFER, glTexFormat, *glPtr)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	PanicIfErrors("MakeArrayBuffer()")
	return glTex
}

func MakeArrayBuffer32 (glPtr *gl.Uint, sliceOrLen interface{}, makeTex bool) gl.Uint {
	if le, isLe := sliceOrLen.(uint64); isLe {
		return MakeArrayBuffer(glPtr, le * 4, nil, false, makeTex)
	}
	return MakeArrayBuffer(glPtr, 0, sliceOrLen, true, makeTex)
}

func MakeAtomicCounters (glPtr *gl.Uint, num gl.Sizei) {
	gl.GenBuffers(1, glPtr)
	gl.BindBuffer(gl.ATOMIC_COUNTER_BUFFER, *glPtr)
	gl.BufferData(gl.ATOMIC_COUNTER_BUFFER, gl.Sizeiptr(4 * num), gl.Pointer(nil), BufMode)
	gl.BindBuffer(gl.ATOMIC_COUNTER_BUFFER, 0)
}

func MakeRttFramebuffer (texFormat gl.Enum, width, height, mipLevels gl.Sizei, anisoFiltering gl.Int) (gl.Uint, gl.Uint) {
	var glPtrTex, glPtrFrameBuf gl.Uint
	var glMagFilter, glMinFilter gl.Int = gl.LINEAR, gl.LINEAR
	if anisoFiltering < 1 { glMagFilter, glMinFilter = gl.NEAREST, gl.NEAREST }
	if mipLevels > 1 { if anisoFiltering < 1 { glMinFilter = gl.NEAREST_MIPMAP_NEAREST } else { glMinFilter = gl.LINEAR_MIPMAP_LINEAR } }
	gl.GenTextures(1, &glPtrTex)
	gl.BindTexture(gl.TEXTURE_2D, glPtrTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, glMagFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, glMinFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	if anisoFiltering > 1 { gl.TexParameteri(gl.TEXTURE_2D, 0x84FE, anisoFiltering) }  // max 16
	gl.TexStorage2D(gl.TEXTURE_2D, mipLevels, texFormat, width, height)
	gl.GenFramebuffers(1, &glPtrFrameBuf)
	gl.BindFramebuffer(gl.FRAMEBUFFER, glPtrFrameBuf)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, glPtrTex, 0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	PanicIfErrors("MakeRttFramebuffer")
	return glPtrFrameBuf, glPtrTex
}

func MakeTexture (glPtr *gl.Uint, dimensions uint8, texFormat gl.Enum, width, height, depth gl.Sizei) {
	MakeTextureForTarget(glPtr, dimensions, width, height, depth, 0, texFormat, true, false)
}

func MakeTextureForTarget (glPtr *gl.Uint, dimensions uint8, width, height, depth gl.Sizei, texTarget gl.Enum, texFormat gl.Enum, panicIfErrors, reuseGlPtr bool) {
	var is3d, is2d = (dimensions == 3), (dimensions == 2)
	if texTarget == 0 { texTarget = gltypes.Ife(is3d, gl.TEXTURE_3D, gltypes.Ife(is2d, gl.TEXTURE_2D, gl.TEXTURE_1D)) }
	if texFormat == 0 { texFormat = gl.RGBA8 }
	if width == 0 { panic("MakeTextureForTarget() needs at least width") }
	if height == 0 { height = width }
	if depth == 0 { depth = height }
	if (!reuseGlPtr) || (*glPtr == 0) { gl.GenTextures(1, glPtr) }
	gl.BindTexture(texTarget, *glPtr)
	gl.TexParameteri(texTarget, gl.TEXTURE_MAG_FILTER, TexFilter)
	gl.TexParameteri(texTarget, gl.TEXTURE_MIN_FILTER, TexFilter)
	gl.TexParameteri(texTarget, gl.TEXTURE_WRAP_S, TexWrap)
	if is2d || is3d { gl.TexParameteri(texTarget, gl.TEXTURE_WRAP_T, TexWrap) }
	if is3d { gl.TexParameteri(texTarget, gl.TEXTURE_WRAP_R, TexWrap) }
	if is3d {
		gl.TexStorage3D(texTarget, 1, texFormat, width, height, depth)
	} else if is2d {
		gl.TexStorage2D(texTarget, 1, texFormat, width, height)
	} else {
		gl.TexStorage1D(texTarget, 1, texFormat, width)
	}
	gl.BindTexture(texTarget, 0)
	if panicIfErrors { PanicIfErrors("MakeTextureForTarget(dim=%v w=%v h=%v d=%v)", dimensions, width, height, depth) }
}

func MakeTextureFromImageFile (filePath string) gl.Uint {
	var file, err = os.Open(filePath)
	var img image.Image
	var tex gl.Uint
	if err != nil { panic(err) }
	defer file.Close()
	img, _, err = image.Decode(file)
	if err != nil { panic(err) }
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	sw, sh := gl.Sizei(w), gl.Sizei(h)
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ { for y := 0; y < h; y++ { rgba.Set(x, y, img.At(x, h - 1 - y)) } }
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.RGBA8, sw, sh)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, sw, sh, gl.RGBA, gl.UNSIGNED_BYTE, gl.Pointer(&rgba.Pix[0]))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	PanicIfErrors("MakeTextureFromImageFile(%v)", filePath)
	return tex
}

func MakeTextureFromImageFloatsFile (filePath string, w, h int) gl.Uint {
	var file, err = os.Open(filePath)
	var tex gl.Uint
	var pix = make([]gl.Float, w * h * 3)
	var fVal float32
	var raw []uint8
	var buf *bytes.Buffer
	var i int
	if err != nil { panic(err) }
	defer file.Close()
	raw, err = ioutil.ReadAll(file)
	if err != nil { panic(err) }
	buf = bytes.NewBuffer(raw)
	for i = 0; (err == nil) && (i < len(pix)); i++ {
		if err = binary.Read(buf, binary.LittleEndian, &fVal); err == io.EOF {
			err = nil; break
		} else if err == nil {
			pix[i] = gl.Float(fVal)
		}
	}
	if err != nil { panic(err) }
	sw, sh := gl.Sizei(w), gl.Sizei(h)
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.RGB16F, sw, sh)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, sw, sh, gl.RGB, gl.FLOAT, gl.Pointer(&pix[0]))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	PanicIfErrors("MakeTextureFromImageFloatsFile(%v)", filePath)
	return tex
}

func MakeTextures (num gl.Sizei, glPtr []gl.Uint, dimensions uint8, texFormat gl.Enum, width, height, depth gl.Sizei) {
	var is3d, is2d = (dimensions == 3), (dimensions == 2)
	var texTarget gl.Enum = gltypes.Ife(is3d, gl.TEXTURE_3D, gltypes.Ife(is2d, gl.TEXTURE_2D, gl.TEXTURE_1D))
	gl.GenTextures(num, &glPtr[0])
	PanicIfErrors("MakeTextures.GenTextures(num=%v)", num)
	if width == 0 { panic("MakeTextures() needs at least width") }
	if height == 0 { height = width }
	if depth == 0 { depth = height }
	for i := 0; i < len(glPtr); i++ {
		gl.BindTexture(texTarget, glPtr[i])
		gl.TexParameteri(texTarget, gl.TEXTURE_MAG_FILTER, TexFilter)
		gl.TexParameteri(texTarget, gl.TEXTURE_MIN_FILTER, TexFilter)
		gl.TexParameteri(texTarget, gl.TEXTURE_WRAP_R, TexWrap)
		gl.TexParameteri(texTarget, gl.TEXTURE_WRAP_S, TexWrap)
		gl.TexParameteri(texTarget, gl.TEXTURE_WRAP_T, TexWrap)
		if is3d {
			gl.TexStorage3D(texTarget, 1, texFormat, width, height, depth)
		} else if is2d {
			gl.TexStorage2D(texTarget, 1, texFormat, width, height)
		} else {
			gl.TexStorage1D(texTarget, 1, texFormat, width)
		}
		gl.BindTexture(texTarget, 0)
		PanicIfErrors("MakeTextures.Loop(i=%v dim=%v w=%v h=%v d=%v)", i, dimensions, width, height, depth)
	}
}

func MaxTextureBufferSize () uint64 {
	if maxTexBufSize == 0 {
		gl.GetIntegerv(gl.MAX_TEXTURE_BUFFER_SIZE, &maxTexBufSize)
		PanicIfErrors("MaxTextureBufferSize()")
	}
	return uint64(maxTexBufSize)
}

func MaxTextureSize1D () gl.Int {
	if maxTexSize1D == 0 {
		gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxTexSize1D)
		PanicIfErrors("MaxTextureSize1D()")
	}
	return maxTexSize1D
}

func MaxTextureSize2D () gl.Int {
	if maxTexSize2D == 0 {
		gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxTexSize2D)
		PanicIfErrors("MaxTextureSize2D()")
	}
	return maxTexSize2D
}

func MaxTextureSize3D () gl.Int {
	if maxTexSize3D == 0 {
		gl.GetIntegerv(gl.MAX_3D_TEXTURE_SIZE, &maxTexSize3D)
		PanicIfErrors("MaxTextureSize3D()")
	}
	return maxTexSize3D
}

func PanicIfErrors (step string, fmtArgs ... interface{}) {
	var errs = LastErrors()
	var ln string
	if len(errs) > 0 {
		if len(fmtArgs) > 0 { step = fmt.Sprintf(step, fmtArgs ...) }
		log.Printf("OpenGL error(s) at step %s...", step)
		for _, err := range errs {
			switch err {
			case gl.INVALID_ENUM:
				ln = "GL_INVALID_ENUM"
			case gl.INVALID_VALUE:
				ln = "GL_INVALID_VALUE"
			case gl.INVALID_OPERATION:
				ln = "GL_INVALID_OPERATION"
			case gl.OUT_OF_MEMORY:
				ln = "GL_OUT_OF_MEMORY"
			case gl.INVALID_FRAMEBUFFER_OPERATION:
				ln = "GL_INVALID_FRAMEBUFFER_OPERATION"
			default:
				ln = fmt.Sprintf("%v", err)
			}
			log.Printf("\t\t%s", ln)
		}
		log.Panicf("Aborting.")
	}
}

func ReadAtomicCounterValues (ac gl.Uint, vals []gl.Uint) {
	var ptr *gl.Uint
	gl.BindBuffer(gl.ATOMIC_COUNTER_BUFFER, ac)
	for i := 0; i < len(vals); i++ {
		ptr = (*gl.Uint)(gl.MapBufferRange(gl.ATOMIC_COUNTER_BUFFER, gltypes.OffsetIntPtr(nil, gl.Sizei(i * 4)), gltypes.SizeOfGlUint, gl.MAP_READ_BIT))
		vals[i] = *ptr
		gl.UnmapBuffer(gl.ATOMIC_COUNTER_BUFFER)
	}
	gl.BindBuffer(gl.ATOMIC_COUNTER_BUFFER, 0)
	PanicIfErrors("ReadAtomicCounter(%v)", ac)
}

func ResetAtomicCounters (glPtr gl.Uint, num gl.Sizei, value gl.Uint) {
	var ptr *gl.Uint
	var i gl.Sizei
	gl.BindBuffer(gl.ATOMIC_COUNTER_BUFFER, glPtr)
	for i = 0; i < num; i++ {
		ptr = (*gl.Uint)(gl.MapBufferRange(gl.ATOMIC_COUNTER_BUFFER, gltypes.OffsetIntPtr(nil, i * 4), gltypes.SizeOfGlUint, gl.MAP_WRITE_BIT | gl.MAP_INVALIDATE_BUFFER_BIT | gl.MAP_UNSYNCHRONIZED_BIT))
		*ptr = value
		gl.UnmapBuffer(gl.ATOMIC_COUNTER_BUFFER)
	}
	gl.BindBuffer(gl.ATOMIC_COUNTER_BUFFER, 0)
}

func ShaderInfoLog (shaderOrProgram gl.Uint, isShader bool) string {
	var l = gl.Sizei(256)
	var s = gl.GLStringAlloc(l)
	defer gl.GLStringFree(s)
	if isShader { gl.GetShaderInfoLog(shaderOrProgram, l, nil, s) } else { gl.GetProgramInfoLog(shaderOrProgram, l, nil, s) }
	PanicIfErrors("ShaderInfoLog(s=%v)", isShader)
	return gl.GoString(s)
}

func ShaderLocation (glProg gl.Uint, name string, isAtt bool) gl.Int {
	var loc gl.Int
	var s = gl.GLString(name)
	defer gl.GLStringFree(s)
	if isAtt { loc = gl.GetAttribLocation(glProg, s) } else { loc = gl.GetUniformLocation(glProg, s) }
	// if loc < 0 { log.Panicf("sprog att/uni %s: bind error %d", name, loc) }
	// PanicIfErrors("ShaderLocation(n=%v a=%v)", name, isAtt)
	LastErrors()
	return loc
}

func ShaderLocationA (glProg gl.Uint, name string) gl.Uint {
	return gl.Uint(ShaderLocation(glProg, name, true))
}

func ShaderLocationU (glProg gl.Uint, name string) gl.Int {
	return ShaderLocation(glProg, name, false)
}

func ShaderSource (name string, shader gl.Uint, source string, defines map[string]interface{}, logPrint bool) {
	var src []*gl.Char
	var i, l = 1, len(defines)
	var lines = make([]string, (l * 5) + 3)
	var joined string
	lines[0] = "#version 420 core\n"
	for dk, dv := range defines {
		lines[i + 0] = "#define "
		lines[i + 1] = dk
		lines[i + 2] = " "
		lines[i + 3] = fmt.Sprintf("%v", dv)
		lines[i + 4] = "\n"
		i = i + 5
	}
	lines[i] = "#line 1\n"
	lines[i + 1] = source
	joined = strings.Join(lines, "")
	src = gl.GLStringArray(lines ...)
	defer gl.GLStringArrayFree(src)
	gl.ShaderSource(shader, gl.Sizei(len(src)), &src[0], nil)
	PanicIfErrors("ShaderSource(name=%v source=%v)", name, joined)
	if logPrint { log.Printf("\n\n------------------------------\n%s\n\n", joined) }
}

func VramAvailable (val uint64) uint64 {
	var tmp uint64
	var getMem = func (n gl.Enum) uint64 { return uint64(Integerv(n)) * 1024 }
	if Extension("gpu_memory_info") {
		tmp = getMem(GPU_MEMORY_INFO_DEDICATED_VIDMEM_NVX)
		val = tmp
		if tmp = getMem(GPU_MEMORY_INFO_TOTAL_AVAILABLE_MEMORY_NVX); tmp < val { val = tmp }
		if tmp = getMem(GPU_MEMORY_INFO_CURRENT_AVAILABLE_VIDMEM_NVX); tmp < val { val = tmp }
	} else if Extension("meminfo") {
		val = getMem(VBO_FREE_MEMORY_ATI) + getMem(TEXTURE_FREE_MEMORY_ATI) + getMem(RENDERBUFFER_FREE_MEMORY_ATI)
	}
	return val
	/*
	var glPtr gl.Uint
	var errs []gl.Enum
	var ms uint64 = 0
	PanicIfErrors("Vram().Pre")
	for ms = 16; ms < uint64(MaxTextureSize()); ms++ {
		MakeTextureForTarget(&glPtr, 3, gl.Sizei(ms), 0, 0, gl.TEXTURE_3D, false, true)
		if errs = LastErrors(gl.OUT_OF_MEMORY); len(errs) > 0 {
			ms--
			break
		}
	}
	gl.DeleteTextures(1, &glPtr)
	return ms * ms * ms * 4
	*/
}

func WriteFloats4 (vecX, vecY, vecZ, vecW float64, index int, glFloats []gl.Float) {
	glFloats[index + 0], glFloats[index + 1], glFloats[index + 2], glFloats[index + 3] = gl.Float(vecX), gl.Float(vecY), gl.Float(vecZ), gl.Float(vecW)
}

func WriteFloatsVec3 (vec num.Vec3, index int, glFloats []gl.Float) {
	glFloats[index + 0], glFloats[index + 1], glFloats[index + 2] = gl.Float(vec.X), gl.Float(vec.Y), gl.Float(vec.Z)
}

func WriteFloatsVec4 (vec num.Vec3, vecW float64, index int, glFloats []gl.Float) {
	glFloats[index + 0], glFloats[index + 1], glFloats[index + 2], glFloats[index + 3] = gl.Float(vec.X), gl.Float(vec.Y), gl.Float(vec.Z), gl.Float(vecW)
}

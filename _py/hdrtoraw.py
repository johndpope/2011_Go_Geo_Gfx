import OpenEXR
import array
import Imath
import Image
import sys

def main(exrFilePath):
    exrFile = OpenEXR.InputFile(exrFilePath)
    dw = exrFile.header()['dataWindow']
    pt = Imath.PixelType(Imath.PixelType.FLOAT)
    size = (dw.max.x - dw.min.x + 1, dw.max.y - dw.min.y + 1)
    width, height = size[0], size[1]
    print size
    rstr = exrFile.channel("R", pt)
    gstr = exrFile.channel("G", pt)
    bstr = exrFile.channel("B", pt)
    rgbf = [Image.fromstring("F", size, exrFile.channel(c, pt)) for c in "RGB"]
    for im in rgbf:
        print im.getextrema()

    return
    red, green, blue = array.array("f", rstr), array.array("f", gstr), array.array("f", bstr)
    floats = array.array("f")
    for i in range(0, len(red)):
    	floats.append(red[i])
    	floats.append(green[i])
    	floats.append(blue[i])
    outFile = open(exrFilePath + "_" + str(width) + "x" + str(height) + ".float", "wb")
    floats.tofile(outFile)
    outFile.close()

    return
    rgbf = [Image.fromstring("F", size, exrFile.channel(c, pt)) for c in "RGB"]
    extrema = [im.getextrema() for im in rgbf]
    darkest = min([lo for (lo,hi) in extrema])
    lighest = max([hi for (lo,hi) in extrema])
    scale = 255 / (lighest - darkest)
    def normalize_0_255(v):
        return (v * scale) + darkest
    rgb8 = [im.point(normalize_0_255).convert("L") for im in rgbf]
    Image.merge("RGB", rgb8).save(exrFilePath + "_" + str(width) + "x" + str(height) + ".jpg")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print "usage: exr2jpg <exrfile>"
    main(sys.argv[1])

package captcha

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	math_tools "github.com/MHSarmadi/Umbra/Server/math"
)

func GenerateNumericCaptcha(number string) ([]byte, error) {
	const (
		scale  = 3
		height = 90
	)
	var width = 55*len(number) + 20

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(img, img.Bounds(), &image.Uniform{
		C: color.RGBA{
			uint8(math_tools.RandomInt32(200, 255)),
			uint8(math_tools.RandomInt32(200, 255)),
			uint8(math_tools.RandomInt32(200, 255)),
			255,
		},
	}, image.Point{}, draw.Src)

	x := 12
	y := 12

	for _, ch := range number {
		glyph, ok := DigitFont[ch]
		if !ok {
			continue
		}

		drawDigit16x20(
			img,
			glyph,
			x,
			y+int(math_tools.RandomInt32(-13, 17)),
			scale,
			color.RGBA{
				uint8(math_tools.RandomInt32(0, 90)),
				uint8(math_tools.RandomInt32(0, 90)),
				uint8(math_tools.RandomInt32(0, 90)),
				255,
			},
		)

		x += 18*scale + int(math_tools.RandomInt32(-6, 6))
	}

	addRandomNoise(img)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func drawDigit16x20(
	img *image.RGBA,
	glyph [20]uint16,
	x, y int,
	scale int,
	col color.Color,
) {
	for row := range 20 {
		line := glyph[row]

		for bit := range 16 {
			if (line>>(15-bit))&1 == 1 {
				for dy := range scale {
					for dx := range scale {
						img.Set(
							x+bit*scale+dx,
							y+row*scale+dy,
							col,
						)
					}
				}
			}
		}
	}
}

func addRandomNoise(img *image.RGBA) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	for x := range w {
		for y := range h {
			if math_tools.RandomInt32(0, 10) > 5 {
				rgba := img.RGBAAt(x, y)
				rgba.A = uint8(math_tools.RandomInt32(100, 255)) - rgba.A
				rgba.G = uint8(math_tools.RandomInt32(100, 255)) - rgba.G
				rgba.B = uint8(math_tools.RandomInt32(100, 255)) - rgba.B
				img.Set(x, y, rgba)
			}
		}
	}
}

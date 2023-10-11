package atlas_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/plugins/spritetools/atlas"
	"golang.org/x/image/colornames"
)

func fill(w, h int, c color.Color) (pd *pixel.PictureData) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, c)
		}
	}
	return pixel.PictureDataFromImage(img)
}

// func colorEq(i2 image.Image, w, h int, c color.Color) (err error) {
// 	var i1 = fill(w, h, c)

// 	if !i1.Bounds().Size().Eq(i2.Bounds().Size()) {
// 		return fmt.Errorf("Image sizes are not the same: Expected: %s, Got: %s", i1.Bounds().Size(), i2.Bounds().Size())
// 	}

// 	for x := 0; x < i1.Bounds().Dx(); x++ {
// 		for y := 0; y < i1.Bounds().Dy(); y++ {
// 			r1, g1, b1, a1 := i1.At(x, y).RGBA()
// 			r2, g2, b2, a2 := i2.At(x, y).RGBA()
// 			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
// 				return fmt.Errorf("At: (%d, %d), Expected: (%v, %v, %v, %v), Got: (%v, %v, %v, %v)", x, y, r1, g1, b1, a1, r2, b2, g2, a2)
// 			}
// 		}
// 	}

// 	return nil
// }

func TestNewPacker(t *testing.T) {
	t.Run("Test", func(t *testing.T) {
		pack := atlas.NewAtlas[int](atlas.PackerCfg{})
		colors := []struct {
			col  color.Color
			w, h int
		}{
			{
				col: colornames.Black,
				w:   64,
				h:   64,
			},
			{
				col: colornames.Aliceblue,
				w:   128,
				h:   32,
			},
			{
				col: colornames.Navy,
				w:   256,
				h:   256,
			},
			{
				col: colornames.Salmon,
				w:   8,
				h:   192,
			},
			{
				col: colornames.Orchid,
				w:   1024,
				h:   512,
			},
			{
				col: colornames.Olive,
				w:   999,
				h:   999,
			},
			{
				col: colornames.Oldlace,
				w:   15,
				h:   800,
			},
		}
		for i, c := range colors {
			pack.Insert(i, fill(c.w, c.h, c.col))
		}
		if err := pack.Pack(); err != nil {
			t.Error(err)
		}
		if err := pack.Export("test.png"); err != nil {
			t.Error(err)
		}
		// for i, c := range colors {
		// 	img := pack.SubImage(i)
		// 	if err := colorEq(img, c.w, c.h, c.col); err != nil {
		// 		t.Errorf("%d is not expected: %s", i, err.Error())
		// 	}
		// }
	})
}

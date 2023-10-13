package atlas

import (
	"embed"
	"fmt"
	"image"
	"io"
	"os"

	// need the following to automatically register for image.decode
	_ "image/jpeg"
	_ "image/png"
)

func panicif(b bool, f string, a ...any) {
	if b {
		panic(fmt.Sprintf(f, a...))
	}
}

func panicerr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func area(r image.Rectangle) int {
	return r.Dx() * r.Dy()
}

func rect(x, y, w, h int) image.Rectangle {
	return image.Rect(x, y, x+w, y+h)
}

func loadSpriteReader(r io.ReadCloser, err error) (i image.Image, e error) {
	if err != nil {
		return
	}
	defer r.Close()
	i, _, err = image.Decode(r)
	return
}

func loadEmbedSprite(fs embed.FS, file string) (i image.Image, err error) {
	return loadSpriteReader(fs.Open(file))
}

func loadSprite(file string) (i image.Image, err error) {
	return loadSpriteReader(os.Open(file))
}

// split is the actual algorithm for splitting a given space (by j in spcs) to fit the given width and height.
// Will return an empty rectangle if a space wasn't available
// This function is based on this project (https://github.com/TeamHypersomnia/rectpack2D)
func split(spcs spaces, j int, bw, bh int) (found image.Rectangle, newSpcs spaces) {
	sp := spcs[j]
	spw, sph := sp.Dx(), sp.Dy()
	switch {
	// Perfect match
	case bw == spw && bh == sph:
		found = sp
		spcs = append(spcs[:j], spcs[j+1:]...)
	// Perfect width, split height
	case bw == spw && bh < sph:
		h := sph - bh
		found = rect(sp.Min.X, sp.Min.Y, spw, bh)
		spcs = append(spcs[:j], spcs[j+1:]...)
		spcs = append(spcs, rect(sp.Min.X, sp.Min.Y+bh, spw, h))
	// Perfect height, split width
	case bw < spw && bh == sph:
		w := spw - bw
		found = rect(sp.Min.X, sp.Min.Y, bw, sph)
		spcs = append(spcs[:j], spcs[j+1:]...)
		spcs = append(spcs, rect(sp.Min.X+bw, sp.Min.Y, w, sph))
	// Split both
	case bw < spw && bh < sph:
		w := spw - bw
		h := sph - bh
		found = rect(sp.Min.X, sp.Min.Y, bw, bh)
		var r1, r2 image.Rectangle

		// Maximize the leftover size
		r1 = rect(sp.Min.X+bw, sp.Min.Y, w, bh)
		r2 = rect(sp.Min.X, sp.Min.Y+bh, spw, h)

		spcs = append(spcs[:j], spcs[j+1:]...)
		spcs = append(spcs, r1, r2)
	}
	newSpcs = spcs
	return
}

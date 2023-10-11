package atlas

import (
	"github.com/gopxl/pixel/v2"
	"golang.org/x/exp/constraints"
)

type queuedData[K comparable] struct {
	id  K
	pic *pixel.PictureData
}

// container for the leftover space after split
type createdSplits struct {
	hasSmall, hasBig bool
	count            int
	smaller, bigger  pixel.Rect
}

// adds the given leftover spaces to this container
func splits(rects ...pixel.Rect) (s *createdSplits) {
	s = &createdSplits{
		count:    len(rects),
		hasSmall: true,
		smaller:  rects[0],
	}

	if s.count == 2 {
		s.hasBig = true
		s.bigger = rects[1]
	}

	return
}

// helper function to create rectangles
func rect[T constraints.Float | constraints.Integer](x, y, w, h T) pixel.Rect {
	return pixel.Rwh(float64(x), float64(y), float64(w), float64(h))
}

func area(r pixel.Rect) float64 {
	return r.W() * r.H()
}

// helper to split existing space
func split(img, space pixel.Rect) (s *createdSplits, err error) {
	w := space.W() - img.W()
	h := space.H() - img.H()

	if w < 0 || h < 0 {
		return nil, ErrSplitFailed
	} else if w == 0 && h == 0 {
		// perfectly fit case
		return &createdSplits{}, nil
	} else if w > 0 && h == 0 {
		r := rect(space.Min.X+img.W(), space.Min.Y, w, img.H())
		return splits(r), nil
	} else if w == 0 && h > 0 {
		r := rect(space.Min.X, space.Min.Y+img.H(), img.W(), h)
		return splits(r), nil
	}

	var smaller, larger pixel.Rect
	if w > h {
		smaller = rect(space.Min.X, space.Min.Y+img.H(), img.W(), h)
		larger = rect(space.Min.X+img.W(), space.Min.Y, w, space.H())
	} else {
		smaller = rect(space.Min.X+img.W(), space.Min.Y, w, img.H())
		larger = rect(space.Min.X, space.Min.Y+img.H(), space.W(), h)
	}

	return splits(smaller, larger), nil
}

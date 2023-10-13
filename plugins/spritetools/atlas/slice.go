package atlas

import (
	"image"

	"github.com/gopxl/pixel/v2"
)

// A SliceId represents a texture in the atlas added by Atlas.Slice.
// This differs from a TextureId in that it's meant to be drawn with a frame offset (a sub image).
type SliceId TextureId

// Returns a TextureId representing the given frame of the slice
func (s SliceId) Frame(frame uint32) TextureId {
	return TextureId{id: s.id + frame, atlas: s.atlas}
}

func (s SliceId) BoundsOf(frame uint32) image.Rectangle {
	return s.Frame(frame).BoundsOf()
}

func (s SliceId) Draw(m pixel.Matrix, frame uint32) {
	s.Frame(frame).Draw(m)
}

func (s SliceId) Im(frame uint32, size pixel.Vec) {
	s.Frame(frame).Im(size)
}

func (s SliceId) ImV(frame uint32, size pixel.Vec, tint, border pixel.RGBA) {
	s.Frame(frame).ImV(size, tint, border)
}

func (s SliceId) ImButton(id string, frame uint32, size pixel.Vec) bool {
	return s.Frame(frame).ImButton(id, size)
}

func (s SliceId) ImButtonV(id string, frame uint32, size pixel.Vec, framePadding int, bg, tint pixel.RGBA) bool {
	return s.Frame(frame).ImButtonV(id, size, framePadding, bg, tint)
}

package atlas

import (
	"image"

	"github.com/gopxl/pixel/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type TextureId struct {
	id    uint32
	atlas *Atlas
}

func (t TextureId) BoundsOf() image.Rectangle {
	s, has := t.atlas.idMap[t.id]
	panicif(!has, "id: %v does not exit in atlas", t.id)
	return image.Rect(0, 0, s.rect.Dx(), s.rect.Dy())
}

func (t TextureId) Draw(m pixel.Matrix) {
	panicif(!t.atlas.clean, "Packer is dirty, call Packer.MakeTexture() first")

	l, has := t.atlas.idMap[t.id]
	panicif(!has, "id [%v] does not exist in packer", t.id)
	r := l.rect

	frame := pixel.R(float64(r.Min.X), float64(r.Min.Y), float64(r.Dx()), float64(r.Dy()))
	t.atlas.internal[l.index].Draw(frame, m)
}

func (t TextureId) Im(size pixel.Vec) {
	t.ImV(size, pixel.Col(255, 255, 255, 255), pixel.Col(0, 0, 0, 0))
}

func (t TextureId) ImV(size pixel.Vec, tint, border pixel.RGBA) {
	panicif(!t.atlas.clean, "Packer is dirty, call Packer.MakeTexture() first")

	l, has := t.atlas.idMap[t.id]
	panicif(!has, "id [%v] does not exist in packer", t.id)
	r := l.rect

	w, h := 1/t.atlas.internal[l.index].Width(), 1/t.atlas.internal[l.index].Height()
	corn := pixel.R(r.Min.X, r.Min.Y, r.Dx(), r.Dy()).Scale(w, h, w, h).Corners()

	imgui.ImageV(t.atlas.internal[l.index].TextureID(), size.Scale(float32(r.Dx()), float32(r.Dy())).Im(), corn[pixel.CornerUL].Im(), corn[pixel.CornerLR].Im(), tint.IM(), border.IM())
}

// push/popID is necessary because the texture ID is used for identification
//
//	and since this is an atlas, multiple things will be using the same texture
func (t TextureId) ImButton(id string, size pixel.Vec) bool {
	return t.ImButtonV(id, size, -1, pixel.Col(0, 0, 0, 0), pixel.Col(255, 255, 255, 255))
}

func (t TextureId) ImButtonV(id string, size pixel.Vec, framePadding int, bg, tint pixel.RGBA) bool {
	panicif(!t.atlas.clean, "Packer is dirty, call Packer.MakeTexture() first")

	l, has := t.atlas.idMap[t.id]
	panicif(!has, "id [%v] does not exist in packer", t.id)
	r := l.rect

	w, h := 1/t.atlas.internal[l.index].Width(), 1/t.atlas.internal[l.index].Height()
	corn := pixel.R(r.Min.X, r.Min.Y, r.Dx(), r.Dy()).Scale(w, h, w, h).Corners()
	imgui.PushID(id)
	defer imgui.PopID()
	return imgui.ImageButtonV(t.atlas.internal[l.index].TextureID(), size.Scale(float32(r.Dx()), float32(r.Dy())).Im(), corn[pixel.CornerUL].Im(), corn[pixel.CornerLR].Im(), framePadding, bg.IM(), tint.IM())
}

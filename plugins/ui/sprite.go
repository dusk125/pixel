package ui

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/inkyblackness/imgui-go"

	"github.com/gopxl/pixel/v2"
)

const (
	WrappedNone = iota
	WrappedSprite
	WrappedBatch
	WrappedCanvas
)

type wrapper struct {
	Type  int
	Value interface{}
}

func Sprite(sprite *pixel.Sprite) imgui.TextureID {
	return imgui.TextureID(unsafe.Pointer(&wrapper{
		Type:  WrappedSprite,
		Value: sprite,
	}))
}

func (ui *UI) AddSprite(id int, sprite *pixel.Sprite) imgui.TextureID {
	if err := ui.packer.Insert(id, sprite.Picture().(*pixel.PictureData)); err != nil {
		log.Fatalln(err)
	}
	return imgui.TextureID(ui.packer.IdOf(name))
}

func (ui *UI) AddSpriteFromFile(path string) (id imgui.TextureID, sprite *pixel.Sprite) {
	return ui.AddSpriteFromFileV(filepath.Base(path[:len(path)-4]), path)
}

func (ui *UI) AddSpriteFromFile(id int, path string) (id imgui.TextureID, sprite *pixel.Sprite) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalln(err)
	}

	data := pixel.PictureDataFromImage(img)
	sprite = pixel.NewSprite(data, data.Bounds())
	id = ui.AddSprite(name, sprite)

	return
}

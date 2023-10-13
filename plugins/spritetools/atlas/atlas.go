package atlas

import (
	"embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"sort"

	"github.com/gopxl/pixel/v2"
)

type embedEntry struct {
	is bool
	fs embed.FS
}

type newEntry struct {
	id     uint32
	path   string
	bounds image.Rectangle
	frame  image.Point
	embed  embedEntry
}

type loc struct {
	index int
	rect  image.Rectangle
}

type spaces []image.Rectangle

type sheet struct {
	size   image.Rectangle
	spaces spaces
}

type Atlas struct {
	adding   []newEntry
	internal []*pixel.PictureData
	clean    bool
	idMap    map[uint32]loc
	id       uint32
}

// func (a *Atlas) Debug() {
// 	if imgui.BeginChild("Atlas") {
// 		if imgui.BeginTabBar("##tabs") {
// 			for i, sprite := range a.internal {
// 				imgui.PushIDInt(i)
// 				if imgui.BeginTabItem(fmt.Sprintf("%v", i)) {
// 					imgui.Textf("%v", sprite.Bounds())
// 					imgui.Image(sprite.TextureID(), sprite.Size().Im().Times(0.5))
// 				}
// 				imgui.EndTabItem()
// 				imgui.PopID()
// 			}
// 		}
// 		imgui.EndTabBar()
// 	}
// 	imgui.EndChild()
// }

func (a *Atlas) Dump() {
	for i, t := range a.internal {
		f, err := os.Create(fmt.Sprintf("%v.png", i))
		if err != nil {
			log.Println(i, err)
			continue
		}
		defer f.Close()

		if err := png.Encode(f, t.Image()); err != nil {
			log.Println(i, err)
			continue
		}
	}
}

func (a *Atlas) AddEmbed(fs embed.FS, path string) (id TextureId) {
	img, err := loadEmbedSprite(fs, path)
	panicerr(err)
	bounds := img.Bounds()
	id = TextureId{id: a.id, atlas: a}
	a.id++
	a.adding = append(a.adding, newEntry{id: id.id, path: path, bounds: bounds, embed: embedEntry{is: true, fs: fs}})
	a.clean = false
	return
}

func (a *Atlas) AddFile(path string) (id TextureId) {
	img, err := loadSprite(path)
	panicerr(err)
	bounds := img.Bounds()
	id = TextureId{id: a.id, atlas: a}
	a.id++
	a.adding = append(a.adding, newEntry{id: id.id, path: path, bounds: bounds})
	a.clean = false
	return
}

func (a *Atlas) Slice(path string, cellSize image.Point) (id SliceId) {
	img, err := loadSprite(path)
	panicerr(err)
	bounds := img.Bounds()
	panicif(bounds.Dx()%cellSize.X != 0 || bounds.Dy()%cellSize.Y != 0, "Texture size (%v,%v) must be multiple of cellSize (%v,%v)", bounds.Dx(), bounds.Dy(), cellSize.X, cellSize.Y)

	id = SliceId{id: a.id, atlas: a}
	a.id += uint32((bounds.Dx() / cellSize.X) * (bounds.Dy() / cellSize.Y))
	a.adding = append(a.adding, newEntry{id: id.id, path: path, bounds: bounds, frame: cellSize})
	a.clean = false
	return
}

func (a *Atlas) SliceEmbed(fs embed.FS, path string, cellSize image.Point) (id SliceId) {
	img, err := loadEmbedSprite(fs, path)
	panicerr(err)
	bounds := img.Bounds()
	panicif(bounds.Dx()%cellSize.X != 0 || bounds.Dy()%cellSize.Y != 0, "Texture size (%v,%v) must be multiple of cellSize (%v,%v)", bounds.Dx(), bounds.Dy(), cellSize.X, cellSize.Y)

	id = SliceId{id: a.id, atlas: a}
	a.id += uint32((bounds.Dx() / cellSize.X) * (bounds.Dy() / cellSize.Y))
	a.adding = append(a.adding, newEntry{id: id.id, path: path, bounds: bounds, frame: cellSize, embed: embedEntry{is: true, fs: fs}})
	a.clean = false
	return
}

// Pack takes all of the added textures and adds them to the atlas largest to smallest,
//
//	trying to waste as little space as possible. After this call, the textures added
//	to the atlas can be used.
func (a *Atlas) Pack(maxSheetW, maxSheetH int) (err error) {
	// If there's nothing to do, don't do anything
	if a.clean {
		return
	}

	// reset internal stuff
	a.internal = a.internal[:0]
	a.idMap = make(map[uint32]loc)

	sort.Slice(a.adding, func(i, j int) bool {
		return area(a.adding[i].bounds) >= area(a.adding[j].bounds)
	})

	sheets := make([]sheet, 1)
	for i := range sheets {
		sheets[i] = sheet{
			spaces: []image.Rectangle{image.Rect(0, 0, maxSheetW, maxSheetH)},
		}
	}

	for _, add := range a.adding {
		bw, bh := add.bounds.Dx(), add.bounds.Dy()

		if bw > maxSheetW || bh > maxSheetH {
			return fmt.Errorf("Texture for %v is larger (%v, %v) than the maximum allowed texture (%v, %v)", add.path, bw, bh, maxSheetW, maxSheetH)
		}

		found := image.Rectangle{}
		foundI := -1

	Loop:
		for i := range sheets {
			for j := range sheets[i].spaces {
				found, sheets[i].spaces = split(sheets[i].spaces, j, bw, bh)
				if found.Empty() {
					continue
				}
				sort.Slice(sheets[i].spaces, func(a, b int) bool {
					return area(sheets[i].spaces[a]) < area(sheets[i].spaces[b])
				})
				foundI = i
				break Loop
			}
		}

		if foundI == -1 {
			foundI = len(sheets)
			sheets = append(sheets, sheet{})
			found, sheets[foundI].spaces = split([]image.Rectangle{image.Rect(0, 0, maxSheetW, maxSheetH)}, 0, bw, bh)
		}

		// Increase the size of the packer so we can allocate the minimum-sized
		// 	texture later.
		if found.Min.X == 0 {
			sheets[foundI].size.Max.Y += found.Dy()
		}
		if found.Min.Y == 0 {
			sheets[foundI].size.Max.X += found.Dx()
		}

		if add.frame.Eq(image.Point{}) {
			// Found a spot, add it to the map
			a.idMap[add.id] = loc{
				index: foundI,
				rect:  found,
			}
		} else {
			// If we have a frame, that means we just added a sprite sheet to the sprite sheet
			// 	so we need to add id entries for each of the sprites
			id := add.id
			for y := 0; y < add.bounds.Dy(); y += add.frame.Y {
				for x := 0; x < add.bounds.Dx(); x += add.frame.X {
					a.idMap[id] = loc{
						index: foundI,
						rect:  rect(found.Min.X+x, found.Min.Y+y, add.frame.X, add.frame.Y),
					}
					id++
				}
			}
		}
	}

	// Create internal textures
	sprites := make([]*image.RGBA, len(sheets))
	for i := range sheets {
		if !sheets[i].size.Empty() {
			sprites[i] = image.NewRGBA(sheets[i].size)
		}
	}

	// Copy individual sprite data into internal textures
	for _, add := range a.adding {
		var (
			err    error
			sprite image.Image
			s      = a.idMap[add.id]
		)
		if add.embed.is {
			sprite, err = loadEmbedSprite(add.embed.fs, add.path)
		} else {
			sprite, err = loadSprite(add.path)
		}
		if err != nil {
			return err
		}
		draw.Draw(sprites[s.index], rect(s.rect.Min.X, s.rect.Min.Y, add.bounds.Dx(), add.bounds.Dy()), sprite, image.Point{}, draw.Src)
	}

	a.internal = make([]*pixel.PictureData, len(sprites))
	for i, sprite := range sprites {
		a.internal[i] = pixel.PictureDataFromImage(sprite)
	}

	a.adding = nil
	a.clean = true

	return
}

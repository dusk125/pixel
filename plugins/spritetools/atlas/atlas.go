package atlas

import (
	"cmp"
	"errors"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"sort"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/plugins/spritetools"
	"golang.org/x/exp/slices"
)

// This texture packer algorithm is based on this project
// https://github.com/TeamHypersomnia/rectpack2D

var (
	ErrNoEmptySpace       = errors.New("Couldn't find an empty space")
	ErrSplitFailed        = errors.New("Split failed")
	ErrGrowthFailed       = errors.New("A previously added texture failed to be added after packer growth")
	ErrUnsupportedSaveExt = errors.New("Unsupported save filename extension")
	ErrNotPacked          = errors.New("Packer must be packed")
	ErrNotFoundNoDefault  = errors.New("Id doesn't exist and a default sprite wasn't specified")
	ErrAlreadyPacked      = errors.New("Pack has already been called for this packer")
)

type PackFlags uint8
type CreateFlags uint8

type PackerCfg struct {
	Flags CreateFlags
}

type Packer[K comparable] struct {
	cfg         PackerCfg
	bounds      pixel.Rect
	emptySpaces []pixel.Rect
	queued      []queuedData[K]
	rects       map[K]pixel.Rect
	images      map[K]*pixel.PictureData
	pic         *pixel.PictureData
	packed      bool
}

// Creates a new packer instance
func NewAtlas[K comparable](cfg PackerCfg) (pack *Packer[K]) {
	bounds := rect(0, 0, 0, 0)
	pack = &Packer[K]{
		cfg:         cfg,
		bounds:      bounds,
		emptySpaces: make([]pixel.Rect, 0),
		rects:       make(map[K]pixel.Rect),
		images:      make(map[K]*pixel.PictureData),
		queued:      make([]queuedData[K], 0),
	}
	return
}

// Inserts PictureData into the packer
func (pack *Packer[K]) Insert(id K, pic *pixel.PictureData) {
	pack.queued = append(pack.queued, queuedData[K]{id: id, pic: pic})
}

// Automatically parse and insert image from file.
func (pack *Packer[K]) InsertFromFile(id K, filename string) (err error) {
	pic, err := spritetools.LoadPictureData(filename)
	if err != nil {
		return
	}

	pack.Insert(id, pic)

	return
}

// Helper to find the smallest empty space that'll fit the given bounds
func (pack Packer[K]) find(bounds pixel.Rect) (index int, found bool) {
	for i, space := range pack.emptySpaces {
		if bounds.W() <= space.W() && bounds.H() <= space.H() {
			return i, true
		}
	}
	return
}

// Helper to remove a canidate empty space and return it
func (pack *Packer[K]) remove(i int) (removed pixel.Rect) {
	removed = pack.emptySpaces[i]
	pack.emptySpaces = append(pack.emptySpaces[:i], pack.emptySpaces[i+1:]...)
	return
}

// Helper to increase the size of the internal texture and readd the queued textures to keep it defragmented
func (pack *Packer[K]) grow(growBy pixel.Vec, endex int) (err error) {
	newSize := pack.bounds.Size().Add(growBy)
	pack.bounds = rect(pack.bounds.Min.X, pack.bounds.Min.Y, newSize.X, newSize.Y)
	pack.emptySpaces = []pixel.Rect{pack.bounds}

	for _, data := range pack.queued[0:endex] {
		if err = pack.insert(data); err != nil {
			return
		}
	}

	return
}

// Helper to segment a found space so that the given data can fit in what's left
func (pack *Packer[K]) insert(data queuedData[K]) (err error) {
	var (
		s            *createdSplits
		bounds       = data.pic.Bounds()
		index, found = pack.find(bounds)
	)

	if !found {
		return ErrGrowthFailed
	}

	space := pack.remove(index)
	if s, err = split(bounds, space); err != nil {
		return
	}

	if s.hasBig {
		pack.emptySpaces = append(pack.emptySpaces, s.bigger)
	}
	if s.hasSmall {
		pack.emptySpaces = append(pack.emptySpaces, s.smaller)
	}

	slices.SortStableFunc(pack.emptySpaces, func(i, j pixel.Rect) int {
		return cmp.Compare(area(i), area(j))
	})

	pack.rects[data.id] = rect(space.Min.X, space.Min.Y, bounds.W(), bounds.H())
	pack.images[data.id] = data.pic
	return
}

// Pack takes the added textures and packs them into the packer texture, growing the texture if necessary.
func (pack *Packer[K]) Pack() (err error) {
	if pack.packed {
		return ErrAlreadyPacked
	}

	// sort queued images largest to smallest
	sort.SliceStable(pack.queued, func(i, j int) bool {
		return area(pack.queued[i].pic.Bounds()) > area(pack.queued[j].pic.Bounds())
	})

	for i, data := range pack.queued {
		var (
			bounds   = data.pic.Bounds()
			_, found = pack.find(bounds)
		)

		if !found {
			if err = pack.grow(bounds.Size(), i); err != nil {
				return
			}
		}

		if err = pack.insert(data); err != nil {
			return
		}
	}

	// img := image.NewRGBA(image.Rect(int(pack.bounds.Min.X), int(pack.bounds.Min.Y), int(pack.bounds.Max.X), int(pack.bounds.Max.Y)))
	// for id, pic := range pack.images {
	// for x := 0; x < int(pic.Bounds().W()); x++ {
	// 	for y := 0; y < int(pic.Bounds().H()); y++ {
	// 		var (
	// 			rect = pack.rects[id]
	// 		)
	// 		img.Set(x+rect.Min.X, y+rect.Min.Y, pic.At(x, y))
	// 	}
	// }
	// }

	pack.pic = pixel.MakePictureData(pack.bounds)
	for id, pic := range pack.images {
		var x, y float64
		for x = 0; x < pic.Bounds().W(); x++ {
			for y = 0; y < pic.Bounds().H(); y++ {
				var (
					rect = pack.rects[id]
					pos  = pixel.V(x+rect.Min.X, y+rect.Min.Y)
					ind  = pack.pic.Index(pos)
				)
				pack.pic.Pix[ind] = pic.Pix[pic.Index(pixel.V(x, y))]
			}
		}
	}

	pack.queued = nil
	pack.emptySpaces = nil
	pack.images = nil
	pack.packed = true

	return
}

// Saves the internal texture as a file on disk, the output type is defined by the filename extension
func (pack *Packer[K]) Export(filename string) (err error) {
	if !pack.packed {
		return ErrNotPacked
	}

	var (
		file *os.File
	)

	if err = os.Remove(filename); err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}

	if file, err = os.Create(filename); err != nil {
		return
	}
	defer file.Close()

	switch path.Ext(filename) {
	case ".png":
		err = png.Encode(file, pack.pic.Image())
	case ".jpeg", ".jpg":
		err = jpeg.Encode(file, pack.pic.Image(), nil)
	default:
		err = ErrUnsupportedSaveExt
	}
	if err != nil {
		return
	}

	return
}

// Returns the subimage bounds from the given id
func (pack *Packer[K]) Get(id K) (rect pixel.Rect) {
	if !pack.packed {
		panic(ErrNotPacked)
	}

	var has bool
	if rect, has = pack.rects[id]; !has {
		panic(ErrNotFoundNoDefault)
	}
	return
}

// Returns the subimage, as a copy, from the given id
// func (pack *Atlas[K]) SubImage(id K) (img *image.RGBA) {
// 	if !pack.packed {
// 		panic(ErrNotPacked)
// 	}

// 	r := pack.Get(id)
// 	i := pack.pic.PixOffset(r.Min.X, r.Min.Y)
// 	return &image.RGBA{
// 		Pix:    pack.pic.Pix[i:],
// 		Stride: pack.pic.Stride,
// 		Rect:   image.Rect(0, 0, r.W(), r.H()),
// 	}
// }

func (pack *Packer[K]) Bounds() pixel.Rect {
	return pack.Picture().Bounds()
}

// Returns the entire packed image
func (pack *Packer[K]) Picture() pixel.Picture {
	if !pack.packed {
		panic(ErrNotPacked)
	}

	return pack.pic
}

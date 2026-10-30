package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
)

// drawIcon draws the app icon (a camera lens in a rounded square) at size n.
func drawIcon(n int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	bg := color.NRGBA{0x2f, 0x6f, 0xed, 0xff}
	ring := color.NRGBA{0xf2, 0xf5, 0xfa, 0xff}
	lens := color.NRGBA{0x12, 0x18, 0x24, 0xff}
	glint := color.NRGBA{0x9c, 0xc0, 0xff, 0xff}
	fn := float64(n)
	rad := fn * 0.22
	c := fn / 2
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			// rounded square mask
			dx := math.Max(math.Max(rad-fx, fx-(fn-rad)), 0)
			dy := math.Max(math.Max(rad-fy, fy-(fn-rad)), 0)
			if dx*dx+dy*dy > rad*rad {
				continue
			}
			col := bg
			d := math.Hypot(fx-c, fy-c)
			switch {
			case d < fn*0.13 && math.Hypot(fx-(c-fn*0.05), fy-(c-fn*0.05)) < fn*0.045:
				col = glint
			case d < fn*0.2:
				col = lens
			case d < fn*0.31:
				col = ring
			}
			// small "recording" dot top-right
			if math.Hypot(fx-fn*0.78, fy-fn*0.22) < fn*0.07 {
				col = color.NRGBA{0xff, 0x5a, 0x5a, 0xff}
			}
			img.SetNRGBA(x, y, col)
		}
	}
	return img
}

func iconPNG(n int) []byte {
	var b bytes.Buffer
	png.Encode(&b, drawIcon(n))
	return b.Bytes()
}

// iconICO builds a .ico holding PNG images (supported since Windows Vista).
func iconICO() []byte {
	sizes := []int{16, 32, 48, 256}
	var imgs [][]byte
	for _, s := range sizes {
		imgs = append(imgs, iconPNG(s))
	}
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, []uint16{0, 1, uint16(len(sizes))})
	off := 6 + 16*len(sizes)
	for i, s := range sizes {
		w := byte(s)
		if s >= 256 {
			w = 0
		}
		b.Write([]byte{w, w, 0, 0})
		binary.Write(&b, binary.LittleEndian, []uint16{1, 32})
		binary.Write(&b, binary.LittleEndian, []uint32{uint32(len(imgs[i])), uint32(off)})
		off += len(imgs[i])
	}
	for _, im := range imgs {
		b.Write(im)
	}
	return b.Bytes()
}

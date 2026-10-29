package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
)

// Alert sounds are synthesised here and played by the app itself (not the
// window), so they work with the window minimised or closed to the tray.
//
//	error    - urgent: three falling tones, played twice
//	warn     - attention: two firm beeps
//	finished - friendly rising chime

type tone struct {
	freq, dur, gap float64 // Hz, seconds, seconds of silence after
}

var soundPatterns = map[string][]tone{
	"error": {{880, .14, .03}, {660, .14, .03}, {440, .22, .28},
		{880, .14, .03}, {660, .14, .03}, {440, .26, 0}},
	"warn":     {{740, .16, .10}, {740, .22, 0}},
	"finished": {{523.25, .14, .02}, {659.25, .14, .02}, {783.99, .34, 0}},
}

var (
	wavMu    sync.Mutex
	wavCache = map[string][]byte{}
)

func soundWAV(kind string) []byte {
	wavMu.Lock()
	defer wavMu.Unlock()
	if b, ok := wavCache[kind]; ok {
		return b
	}
	pat, ok := soundPatterns[kind]
	if !ok {
		return nil
	}
	const rate = 22050
	var pcm []int16
	for _, t := range pat {
		n := int(t.dur * rate)
		for i := 0; i < n; i++ {
			x := float64(i) / rate
			// soft attack and release so there are no clicks
			env := math.Min(1, x/0.012) * math.Min(1, (t.dur-x)/0.05)
			if kind == "finished" {
				env *= math.Exp(-x * 3) // bell-like decay
			}
			v := math.Sin(2*math.Pi*t.freq*x) + 0.25*math.Sin(4*math.Pi*t.freq*x)
			pcm = append(pcm, int16(v/1.25*env*0.55*32767))
		}
		pcm = append(pcm, make([]int16, int(t.gap*rate))...)
	}
	var b bytes.Buffer
	dataLen := uint32(len(pcm) * 2)
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, 36+dataLen)
	b.WriteString("WAVEfmt ")
	binary.Write(&b, binary.LittleEndian, []uint32{16})
	binary.Write(&b, binary.LittleEndian, []uint16{1, 1}) // PCM, mono
	binary.Write(&b, binary.LittleEndian, []uint32{rate, rate * 2})
	binary.Write(&b, binary.LittleEndian, []uint16{2, 16})
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, dataLen)
	binary.Write(&b, binary.LittleEndian, pcm)
	wavCache[kind] = b.Bytes()
	return wavCache[kind]
}

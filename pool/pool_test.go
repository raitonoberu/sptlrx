package pool

import (
	"github.com/raitonoberu/sptlrx/services/hosted"
	"math"
	"testing"
)

func TestGetIndex(t *testing.T) {
	service := hosted.New("lyricsapi.vercel.app")
	lines, err := service.Lyrics("", "Death Grips No Love")
	if err != nil {
		t.Fatal(err)
	}

	test := func(pos, curIndex, expected int) {
		if index := getIndex(pos, curIndex, lines); index != expected {
			t.Errorf("failed getting index for pos %d with curIndex %d: expected %d got %d",
				pos, curIndex, expected, index)
		}
	}

	for i := 0; i < len(lines)-1; i++ {
		line := lines[i]
		dif := lines[i+1].Time - line.Time
		pos := line.Time + (dif / 2)

		for j := 0; j < len(lines); j++ {
			test(pos, j, i)
		}
	}

	// edge cases
	test(0, 0, 0)                      // 0 if pos == 0
	test(lines[0].Time-1, 0, 0)        // 0 if pos < first.Time
	test(math.MaxInt, 0, len(lines)-1) // last if pos > last.Time
}

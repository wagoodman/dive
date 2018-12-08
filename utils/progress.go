package utils

import (
	"fmt"
	"strings"
)

type progressBar struct {
	width      int
	percent    int
	rawTotal   int64
	rawCurrent int64
}

func NewProgressBar(total int64, width int) *progressBar {
	return &progressBar{
		rawTotal: total,
		width:    width,
	}
}

func (pb *progressBar) Done() {
	pb.rawCurrent = pb.rawTotal
	pb.percent = 100
}

func (pb *progressBar) Update(currentValue int64) (hasChanged bool) {
	pb.rawCurrent = currentValue
	percent := int(100.0 * (float64(pb.rawCurrent) / float64(pb.rawTotal)))
	if percent != pb.percent {
		hasChanged = true
	}
	pb.percent = percent
	return hasChanged
}

func (pb *progressBar) String() string {
	done := int((pb.percent * pb.width) / 100.0)
	if done > pb.width {
		done = pb.width
	}
	todo := pb.width - done
	if todo < 0 {
		todo = 0
	}
	head := 1

	return "[" + strings.Repeat("=", done) + strings.Repeat(">", head) + strings.Repeat(" ", todo) + "]" + fmt.Sprintf(" %d %% (%d/%d)", pb.percent, pb.rawCurrent, pb.rawTotal)
}

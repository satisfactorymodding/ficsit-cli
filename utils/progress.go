package utils

import (
	"io"
)

type GenericProgress struct {
	Completed int64
	Total     int64
}

func (gp GenericProgress) Percentage() float64 {
	if gp.Total == 0 {
		return 0
	}
	return float64(gp.Completed) / float64(gp.Total)
}

var _ io.Writer = (*Progresser)(nil)

type Progresser struct {
	Updates chan<- GenericProgress
	Total   int64
	Running int64
}

func (pt *Progresser) Write(p []byte) (int, error) {
	pt.Running += int64(len(p))

	if pt.Updates != nil {
		select {
		case pt.Updates <- GenericProgress{Completed: pt.Running, Total: pt.Total}:
		default:
		}
	}

	return len(p), nil
}

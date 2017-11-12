package main

import (
	"fmt"
)

const (
	b = (float64)(1 << (10 * iota))
	kb
	mb
	gb
	tb
	pb
)

func convertSize(size float64) string {
	if size < 0 {
		return "0B"
	}
	if size < kb {
		return fmt.Sprintf("%fB", size/b)
	}
	if size < mb {
		return fmt.Sprintf("%fKB", size/kb)
	}
	if size < gb {
		return fmt.Sprintf("%fMB", size/mb)
	}
	if size < tb {
		return fmt.Sprintf("%fGB", size/gb)
	}
	if size < pb {
		return fmt.Sprintf("%fGB", size/tb)
	}
	return fmt.Sprintf("%fTB", size/pb)
}

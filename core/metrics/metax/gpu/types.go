package gpu

import (
	"huatuo-bamai/core/metrics/metax/device"
)

type Series string

const (
	Unknown Series = "unknown"
	SeriesN Series = "mxn"
	SeriesC Series = "mxc"
	SeriesG Series = "mxg"
)

type Mode string

const (
	ModeNative Mode = "native"
	ModePf     Mode = "pf"
	ModeVf     Mode = "vf"
)

type Info struct {
	Series      Series
	Model       string
	UUID        string
	BiosVersion string
	BDF         string
	Mode        Mode
	DieCount    uint32
}

var SeriesMap = map[device.Brand]Series{
	device.BrandUnknown: Unknown,
	device.BrandN:       SeriesN,
	device.BrandC:       SeriesC,
	device.BrandG:       SeriesG,
}

var ModeMap = map[device.VirtualizationMode]Mode{
	device.VirtualizationModeNone: ModeNative,
	device.VirtualizationModePf:   ModePf,
	device.VirtualizationModeVf:   ModeVf,
}

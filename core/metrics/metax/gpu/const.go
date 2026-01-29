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

type TemperatureSensor uint32

const (
	TemperatureSensorHotspot TemperatureSensor = iota
)

type UsageIp uint32

const (
	UsageIpDla UsageIp = iota // MetaxSmlUsageIpDla only valid for metaxSmlDeviceBrandN.
	UsageIpVpue
	UsageIpVpud
	UsageIpG2d   // MetaxSmlUsageIpG2d only valid for metaxSmlDeviceBrandN.
	UsageIpXcore // MetaxSmlUsageIpXcore only valid for metaxSmlDeviceBrandC.
)

type ClockIp uint32

const (
	ClockIpCsc ClockIp = iota
	ClockIpDla
	ClockIpMc
	ClockIpMc0
	ClockIpMc1
	ClockIpVpue
	ClockIpVpud
	ClockIpSoc
	ClockIpDnoc
	ClockIpG2d
	ClockIpCcx
	ClockIpXcore
)

type DpmIp uint32

const (
	DpmIpDla DpmIp = iota
	DpmIpXcore
	DpmIpMc
	DpmIpSoc
	DpmIpDnoc
	DpmIpVpue
	DpmIpVpud
	DpmIpHbm
	DpmIpG2d
	DpmIpHbmPower
	DpmIpCcx
	DpmIpIpGroup
	DpmIpDma
	DpmIpCsc
	DpmIpEth
	DpmIpDidt
	DpmIpReserved
)

var (
	UtilizationIpMap = map[string]UsageIp{
		"encoder": UsageIpVpue,
		"decoder": UsageIpVpud,
		"xcore":   UsageIpXcore,
	}
	ClockIpMap = map[string]ClockIp{
		"encoder": ClockIpVpue,
		"decoder": ClockIpVpud,
		"xcore":   ClockIpXcore,
		"memory":  ClockIpMc0,
	}
	DpmIpMap = map[string]DpmIp{
		"xcore": DpmIpXcore,
	}
)

var ClocksThrottleBitReasonMap = map[int]string{
	1:  "idle",
	2:  "application_limit",
	3:  "over_power",
	4:  "chip_overheated",
	5:  "vr_overheated",
	6:  "hbm_overheated",
	7:  "thermal_overheated",
	8:  "pcc",
	9:  "power_brake",
	10: "didt",
	11: "low_usage",
	12: "other",
}

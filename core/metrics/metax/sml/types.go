// Copyright 2025 The HuaTuo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sml

import (
	"huatuo-bamai/core/metrics/metax/device"
)

/*
   MetaX SML API Struct
*/

type SmlPcieInfo = device.PcieLinkInfo
type SmlMetaXLinkAer = device.MetaXLinkAerInfo
type SmlPcieThroughput = device.PcieThroughputInfo
type SmlSingleMetaXLinkInfo = device.MetaXLinkLinkInfo
type SmlBoardWayElectricInfo = device.BoardWayElectricInfo
type SmlEccErrorCount = device.DieEccMemoryInfo

type SmlMetaXLinkTrafficStat struct {
	RequestTrafficStat  int64 // requestTrafficStat in bytes.
	ResponseTrafficStat int64 // responseTrafficStat in bytes.
}
type SmlMetaXLinkBandwidth struct {
	RequestBandwidth  int32 // requestBandwidth in MB/s.
	ResponseBandwidth int32 // responseBandwidth in MB/s.
}

type SmlDeviceUnavailableReasonInfo struct {
	unavailableCode   int32
	unavailableReason [64]byte
}

type SmlTemperatureSensor uint32

const (
	SmlTemperatureSensorHotspot SmlTemperatureSensor = iota
)

type SmlUsageIp uint32

const (
	SmlUsageIpDla SmlUsageIp = iota // MetaxSmlUsageIpDla only valid for metaxSmlDeviceBrandN.
	SmlUsageIpVpue
	SmlUsageIpVpud
	SmlUsageIpG2d   // MetaxSmlUsageIpG2d only valid for metaxSmlDeviceBrandN.
	SmlUsageIpXcore // MetaxSmlUsageIpXcore only valid for metaxSmlDeviceBrandC.
)

type SmlMemoryInfo struct {
	_         int64 // visVramTotal in KB, not used yet.
	_         int64 // visVramUse in KB, not used yet.
	vramTotal int64 // vramTotal in KB.
	vramUse   int64 // vramUse in KB.
	_         int64 // xttTotal in KB, not used yet.
	_         int64 // xttUse in KB, not used yet.
}

type SmlClockIp uint32

const (
	SmlClockIpCsc SmlClockIp = iota
	SmlClockIpDla
	SmlClockIpMc
	SmlClockIpMc0
	SmlClockIpMc1
	SmlClockIpVpue
	SmlClockIpVpud
	SmlClockIpSoc
	SmlClockIpDnoc
	SmlClockIpG2d
	SmlClockIpCcx
	SmlClockIpXcore
)

type SmlDpmIp uint32

const (
	SmlDpmIpDla SmlDpmIp = iota
	SmlDpmIpXcore
	SmlDpmIpMc
	SmlDpmIpSoc
	SmlDpmIpDnoc
	SmlDpmIpVpue
	SmlDpmIpVpud
	SmlDpmIpHbm
	SmlDpmIpG2d
	SmlDpmIpHbmPower
	SmlDpmIpCcx
	SmlDpmIpIpGroup
	SmlDpmIpDma
	SmlDpmIpCsc
	SmlDpmIpEth
	SmlDpmIpDidt
	SmlDpmIpReserved
)

var (
	GpuUtilizationIpMap = map[string]SmlUsageIp{
		"encoder": SmlUsageIpVpue,
		"decoder": SmlUsageIpVpud,
		"xcore":   SmlUsageIpXcore,
	}
	GpuClockIpMap = map[string]SmlClockIp{
		"encoder": SmlClockIpVpue,
		"decoder": SmlClockIpVpud,
		"xcore":   SmlClockIpXcore,
		"memory":  SmlClockIpMc0,
	}
	GpuDpmIpMap = map[string]SmlDpmIp{
		"xcore": SmlDpmIpXcore,
	}
)

var GpuClocksThrottleBitReasonMap = map[int]string{
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

/*
   MetaX SML API RAW SYMBOLS
*/

var (
	// Error and initialization symbols
	mxSmlInit           func() Return
	mxSmlGetErrorString func(Return) string

	// MACA module symbols
	mxSmlGetMacaVersion func(*byte, *uint32) Return

	// Device information symbols
	mxSmlGetDeviceCount    func() uint32
	mxSmlGetPfDeviceCount  func() uint32
	mxSmlGetDeviceInfo     func(uint32, *device.Info) Return
	mxSmlGetDeviceDieCount func(uint32, *uint32) Return
	mxSmlGetDeviceVersion  func(uint32, device.DeviceVersionUnit, *byte, *uint32) Return

	// Board power information symbols
	mxSmlGetBoardPowerInfo func(uint32, *uint32, *SmlBoardWayElectricInfo) Return

	// PCIe information symbols
	mxSmlGetPcieInfo       func(uint32, *SmlPcieInfo) Return
	mxSmlGetPcieThroughput func(uint32, *SmlPcieThroughput) Return

	// MetaXLink symbols (similar to NVLink)
	mxSmlGetMetaXLinkInfo_v2     func(uint32, *uint32, *SmlSingleMetaXLinkInfo) Return
	mxSmlGetMetaXLinkBandwidth   func(uint32, device.MetaXLinkType, *uint32, *SmlMetaXLinkBandwidth) Return
	mxSmlGetMetaXLinkTrafficStat func(uint32, device.MetaXLinkType, *uint32, *SmlMetaXLinkTrafficStat) Return
	mxSmlGetMetaXLinkAer         func(uint32, *uint32, *SmlMetaXLinkAer) Return

	// Die information symbols
	mxSmlGetDieUnavailableReason           func(uint32, uint32, *SmlDeviceUnavailableReasonInfo) Return
	mxSmlGetDieTemperatureInfo             func(uint32, uint32, SmlTemperatureSensor, *int32) Return
	mxSmlGetDieIpUsage                     func(uint32, uint32, SmlUsageIp, *int32) Return
	mxSmlGetDieMemoryInfo                  func(uint32, uint32, *SmlMemoryInfo) Return
	mxSmlGetDieClocks                      func(uint32, uint32, SmlClockIp, *uint32, *uint32) Return
	mxSmlGetDieCurrentClocksThrottleReason func(uint32, uint32, *uint64) Return
	mxSmlGetCurrentDieDpmIpPerfLevel       func(uint32, uint32, SmlDpmIp, *uint32) Return
	mxSmlGetDieTotalEccErrors              func(uint32, uint32, *SmlEccErrorCount) Return
)

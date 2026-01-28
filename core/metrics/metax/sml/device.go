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
	"context"
	"fmt"

	"huatuo-bamai/core/metrics/metax/device"
	metaxgpu "huatuo-bamai/core/metrics/metax/gpu"
)

// getSDKVersion returns the SDK version
func (l *library) getSdkVersion(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	var (
		size uint32 = 128
		buf         = make([]byte, size)
	)
	if err := checkReturnCode("mxSmlGetMacaVersion", mxSmlGetMacaVersion(&buf[0], &size)); err != nil {
		return "", err
	}

	return cString(buf), nil
}

// getNativeAndVFGPUCount returns the number of native and VF GPUs
func (l *library) getNativeAndVfGpuCount() uint32 {
	return mxSmlGetDeviceCount()
}

// getPFGPUCount returns the number of PF GPUs
func (l *library) getPfGpuCount() uint32 {
	return mxSmlGetPfDeviceCount()
}

// getSDKVersion returns the SDK version
func (l *library) getGpuInfo(ctx context.Context, gpu uint32) (metaxgpu.Info, error) {
	select {
	case <-ctx.Done():
		return metaxgpu.Info{}, ctx.Err()
	default:
	}

	var info device.Info
	if err := checkReturnCode("mxSmlGetDeviceInfo", mxSmlGetDeviceInfo(gpu, &info)); err != nil {
		return metaxgpu.Info{}, err
	}

	series, ok := metaxgpu.SeriesMap[info.Brand]
	if !ok {
		return metaxgpu.Info{}, fmt.Errorf("invalid gpu series: %d", info.Brand)
	}

	operationGetBiosVersion := "get bios version"
	biosVersion, err := GetGPUVersion(ctx, gpu, device.DeviceVersionUnitBios)
	if IsNotSupported(err) {
		// Logging is handled in the caller
		biosVersion = ""
	} else if err != nil {
		return metaxgpu.Info{}, fmt.Errorf("failed to %s: %w", operationGetBiosVersion, err)
	}

	mode, ok := metaxgpu.ModeMap[info.Mode]
	if !ok {
		return metaxgpu.Info{}, fmt.Errorf("invalid gpu mode: %d", info.Mode)
	}

	var dieCount uint32
	if err := checkReturnCode("mxSmlGetDeviceDieCount", mxSmlGetDeviceDieCount(gpu, &dieCount)); err != nil {
		return metaxgpu.Info{}, err
	}

	return metaxgpu.Info{
		Series:      series,
		Model:       cString(info.DeviceName[:]),
		UUID:        cString(info.UUID[:]),
		BiosVersion: biosVersion,
		BDF:         cString(info.BDFId[:]),
		Mode:        mode,
		DieCount:    dieCount,
	}, nil
}

// getGPUVersion returns the BIOS or driver version for a GPU
func (l *library) getGpuVersion(ctx context.Context, gpu uint32, unit device.DeviceVersionUnit) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	const versionMaximumSize = 64

	var (
		size uint32 = versionMaximumSize
		buf         = make([]byte, size)
	)
	if err := checkReturnCode("mxSmlGetDeviceVersion", mxSmlGetDeviceVersion(gpu, unit, &buf[0], &size)); err != nil {
		return "", err
	}

	return cString(buf), nil
}

// listGpuBoardWayElectricInfos returns board power information for a GPU
func (l *library) listGpuBoardWayElectricInfos(ctx context.Context, gpu uint32) ([]device.BoardWayElectricInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	const maxBoardWays = 3

	var (
		size uint32 = maxBoardWays
		arr         = make([]SmlBoardWayElectricInfo, size)
	)
	if err := checkReturnCode("mxSmlGetBoardPowerInfo", mxSmlGetBoardPowerInfo(gpu, &size, &arr[0])); err != nil {
		return nil, err
	}

	actualSize := int(size)
	result := make([]device.BoardWayElectricInfo, actualSize)

	for i := 0; i < actualSize; i++ {
		result[i] = device.BoardWayElectricInfo{
			Voltage: arr[i].Voltage,
			Current: arr[i].Current,
			Power:   arr[i].Power,
		}
	}

	return result, nil
}

// getGpuPcieLinkInfo returns PCIe link information for a GPU
func (l *library) getGpuPcieLinkInfo(ctx context.Context, gpu uint32) (device.PcieLinkInfo, error) {
	select {
	case <-ctx.Done():
		return device.PcieLinkInfo{}, ctx.Err()
	default:
	}

	var obj SmlPcieInfo
	if err := checkReturnCode("mxSmlGetPcieInfo", mxSmlGetPcieInfo(gpu, &obj)); err != nil {
		return device.PcieLinkInfo{}, err
	}

	return device.PcieLinkInfo(obj), nil
}

// getGpuPcieThroughputInfo returns PCIe throughput information for a GPU
func (l *library) getGpuPcieThroughputInfo(ctx context.Context, gpu uint32) (device.PcieThroughputInfo, error) {
	select {
	case <-ctx.Done():
		return device.PcieThroughputInfo{}, ctx.Err()
	default:
	}

	var obj SmlPcieThroughput
	if err := checkReturnCode("mxSmlGetPcieThroughput", mxSmlGetPcieThroughput(gpu, &obj)); err != nil {
		return device.PcieThroughputInfo{}, err
	}

	return device.PcieThroughputInfo(obj), nil
}

// listGpuMetaxlinkLinkInfos returns MetaXLink link information for a GPU
func (l *library) listGpuMetaxlinkLinkInfos(ctx context.Context, gpu uint32) ([]device.MetaXLinkLinkInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var (
		size uint32 = device.MetaXLinkMaxNumber
		arr         = make([]SmlSingleMetaXLinkInfo, size)
	)
	if err := checkReturnCode("mxSmlGetMetaXLinkInfo_v2", mxSmlGetMetaXLinkInfo_v2(gpu, &size, &arr[0])); err != nil {
		return nil, err
	}

	actualSize := int(size)
	result := make([]device.MetaXLinkLinkInfo, actualSize)

	for i := range actualSize {
		result[i] = device.MetaXLinkLinkInfo{
			Speed: arr[i].Speed,
			Width: arr[i].Width,
		}
	}

	return result, nil
}

// listGpuMetaxlinkThroughputInfos returns MetaXLink throughput information for a GPU
func (l *library) listGpuMetaxlinkThroughputInfos(ctx context.Context, gpu uint32) ([]device.MetaXLinkThroughputInfo, error) {
	operationListMetaxlinkReceiveRates := "list metaxlink receive rates"
	receiveRates, err := l.listGpuMetaxlinkThroughputParts(ctx, gpu, device.MetaXLinkTypeReceive)
	if IsNotSupported(err) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkReceiveRates, err)
	}

	operationListMetaxlinkTransmitRates := "list metaxlink transmit rates"
	transmitRates, err := l.listGpuMetaxlinkThroughputParts(ctx, gpu, device.MetaXLinkTypeTransmit)
	if IsNotSupported(err) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkTransmitRates, err)
	}

	if len(receiveRates) != len(transmitRates) {
		return nil, fmt.Errorf("receive and transmit array length mismatch")
	}

	result := make([]device.MetaXLinkThroughputInfo, len(receiveRates))

	for i := 0; i < len(result); i++ {
		result[i] = device.MetaXLinkThroughputInfo{
			ReceiveRate:  receiveRates[i],
			TransmitRate: transmitRates[i],
		}
	}

	return result, nil
}

// listGpuMetaxlinkThroughputParts returns MetaXLink throughput data for a specific type
func (l *library) listGpuMetaxlinkThroughputParts(ctx context.Context, gpu uint32, typ device.MetaXLinkType) ([]int32, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var (
		size uint32 = device.MetaXLinkMaxNumber
		arr         = make([]SmlMetaXLinkBandwidth, size)
	)
	if err := checkReturnCode("mxSmlGetMetaXLinkBandwidth", mxSmlGetMetaXLinkBandwidth(gpu, typ, &size, &arr[0])); err != nil {
		return nil, err
	}

	actualSize := int(size)
	result := make([]int32, actualSize)

	for i := 0; i < actualSize; i++ {
		result[i] = arr[i].RequestBandwidth
	}

	return result, nil
}

// listGpuMetaxlinkTrafficStatInfos returns MetaXLink traffic statistics for a GPU
func (l *library) listGpuMetaxlinkTrafficStatInfos(ctx context.Context, gpu uint32) ([]device.MetaXLinkTrafficStatInfo, error) {
	operationListMetaxlinkReceives := "list metaxlink receives"
	receives, err := l.listGpuMetaxlinkTrafficStatParts(ctx, gpu, device.MetaXLinkTypeReceive)
	if IsNotSupported(err) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkReceives, err)
	}

	operationListMetaxlinkTransmits := "list metaxlink transmits"
	transmits, err := l.listGpuMetaxlinkTrafficStatParts(ctx, gpu, device.MetaXLinkTypeTransmit)
	if IsNotSupported(err) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkTransmits, err)
	}

	if len(receives) != len(transmits) {
		return nil, fmt.Errorf("receive and transmit array length mismatch")
	}

	result := make([]device.MetaXLinkTrafficStatInfo, len(receives))

	for i := 0; i < len(result); i++ {
		result[i] = device.MetaXLinkTrafficStatInfo{
			Receive:  receives[i],
			Transmit: transmits[i],
		}
	}

	return result, nil
}

// listGpuMetaxlinkTrafficStatParts returns MetaXLink traffic statistics for a specific type
func (l *library) listGpuMetaxlinkTrafficStatParts(ctx context.Context, gpu uint32, typ device.MetaXLinkType) ([]int64, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var (
		size uint32 = device.MetaXLinkMaxNumber
		arr         = make([]SmlMetaXLinkTrafficStat, size)
	)
	if err := checkReturnCode("mxSmlGetMetaXLinkTrafficStat", mxSmlGetMetaXLinkTrafficStat(gpu, typ, &size, &arr[0])); err != nil {
		return nil, err
	}

	actualSize := int(size)
	result := make([]int64, actualSize)

	for i := 0; i < actualSize; i++ {
		result[i] = arr[i].RequestTrafficStat
	}

	return result, nil
}

// listGpuMetaxlinkAerErrorsInfos returns MetaXLink AER error information for a GPU
func (l *library) listGpuMetaxlinkAerErrorsInfos(ctx context.Context, gpu uint32) ([]device.MetaXLinkAerInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var (
		size uint32 = device.MetaXLinkMaxNumber
		arr         = make([]SmlMetaXLinkAer, size)
	)
	if err := checkReturnCode("mxSmlGetMetaXLinkAer", mxSmlGetMetaXLinkAer(gpu, &size, &arr[0])); err != nil {
		return nil, err
	}

	actualSize := int(size)
	result := make([]device.MetaXLinkAerInfo, actualSize)

	for i := 0; i < actualSize; i++ {
		result[i] = device.MetaXLinkAerInfo{
			CorrectableErrorsCount:   arr[i].CorrectableErrorsCount,
			UncorrectableErrorsCount: arr[i].UncorrectableErrorsCount,
		}
	}

	return result, nil
}

// getDieStatus returns the status of a specific die
func (l *library) getDieStatus(ctx context.Context, gpu, die uint32) (int32, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	var obj SmlDeviceUnavailableReasonInfo
	if err := checkReturnCode("mxSmlGetDieUnavailableReason", mxSmlGetDieUnavailableReason(gpu, die, &obj)); err != nil {
		return 0, err
	}

	return obj.unavailableCode, nil
}

// getDieTemperature returns the temperature of a specific die
func (l *library) getDieTemperature(ctx context.Context, gpu, die uint32, sensor SmlTemperatureSensor) (float64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	var value int32
	if err := checkReturnCode("mxSmlGetDieTemperatureInfo", mxSmlGetDieTemperatureInfo(gpu, die, sensor, &value)); err != nil {
		return 0, err
	}

	return float64(value) / 100, nil
}

// getDieUtilization returns the utilization of a specific IP on a die
// 按 GPU 的 die 和硬件 IP 维度，采集各功能模块的利用率，并上报为监控指标。
func (l *library) getDieUtilization(ctx context.Context, gpu, die uint32, ip SmlUsageIp) (int32, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	var value int32
	if err := checkReturnCode("mxSmlGetDieIpUsage", mxSmlGetDieIpUsage(gpu, die, ip, &value)); err != nil {
		return 0, err
	}

	return value, nil
}

// getDieMemoryInfo returns memory information for a specific die
func (l *library) getDieMemoryInfo(ctx context.Context, gpu, die uint32) (device.DieMemoryInfo, error) {
	select {
	case <-ctx.Done():
		return device.DieMemoryInfo{}, ctx.Err()
	default:
	}

	var obj SmlMemoryInfo
	if err := checkReturnCode("mxSmlGetDieMemoryInfo", mxSmlGetDieMemoryInfo(gpu, die, &obj)); err != nil {
		return device.DieMemoryInfo{}, err
	}

	return device.DieMemoryInfo{
		Total: obj.vramTotal,
		Used:  obj.vramUse,
	}, nil
}

// listDieClocks returns clock information for a specific IP on a die
// 按 die + IP 维度采集 GPU 时钟频率（clock）指标
func (l *library) listDieClocks(ctx context.Context, gpu, die uint32, ip SmlClockIp) ([]uint32, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	const maxClocksSize = 8

	var (
		size uint32 = maxClocksSize
		arr         = make([]uint32, size)
	)
	if err := checkReturnCode("mxSmlGetDieClocks", mxSmlGetDieClocks(gpu, die, ip, &size, &arr[0])); err != nil {
		return nil, err
	}

	actualSize := int(size)
	result := make([]uint32, actualSize)

	for i := 0; i < actualSize; i++ {
		result[i] = arr[i]
	}

	return result, nil
}

// getDieClocksThrottleStatus returns the clocks throttle status for a die
func (l *library) getDieClocksThrottleStatus(ctx context.Context, gpu, die uint32) (uint64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	var value uint64
	if err := checkReturnCode("mxSmlGetDieCurrentClocksThrottleReason", mxSmlGetDieCurrentClocksThrottleReason(gpu, die, &value)); err != nil {
		return 0, err
	}

	return value, nil
}

// getDieDpmPerformanceLevel returns the DPM performance level for a specific IP on a die
// 按 GPU 的 die 和硬件 IP 维度，采集各功能模块当前的 DPM 性能等级，并导出为监控指标。
func (l *library) getDieDpmPerformanceLevel(ctx context.Context, gpu, die uint32, ip SmlDpmIp) (uint32, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	var value uint32
	if err := checkReturnCode("mxSmlGetCurrentDieDpmIpPerfLevel", mxSmlGetCurrentDieDpmIpPerfLevel(gpu, die, ip, &value)); err != nil {
		return 0, err
	}

	return value, nil
}

// getDieEccMemoryInfo returns ECC memory information for a specific die
func (l *library) getDieEccMemoryInfo(ctx context.Context, gpu, die uint32) (device.DieEccMemoryInfo, error) {
	select {
	case <-ctx.Done():
		return device.DieEccMemoryInfo{}, ctx.Err()
	default:
	}

	var obj SmlEccErrorCount
	if err := checkReturnCode("mxSmlGetDieTotalEccErrors", mxSmlGetDieTotalEccErrors(gpu, die, &obj)); err != nil {
		return device.DieEccMemoryInfo{}, err
	}

	return device.DieEccMemoryInfo(obj), nil
}

// func metaxGetGpuInfo(ctx context.Context, gpu uint32) (metaxgpu.Info, error) {
// 	select {
// 	case <-ctx.Done():
// 		return metaxgpu.Info{}, ctx.Err()
// 	default:
// 	}

// 	var info device.Info
// 	if err := checkReturnCode("mxSmlGetDeviceInfo", mxSmlGetDeviceInfo(gpu, &info)); err != nil {
// 		return metaxgpu.Info{}, err
// 	}

// 	series, ok := metaxgpu.SeriesMap[info.Brand]
// 	if !ok {
// 		return metaxgpu.Info{}, fmt.Errorf("invalid gpu series: %d", info.Brand)
// 	}

// 	operationGetBiosVersion := "get bios version"
// 	biosVersion, err := GetGPUVersion(ctx, gpu, device.DeviceVersionUnitBios)
// 	if IsNotSupported(err) {
// 		// Logging is handled in the caller
// 		biosVersion = ""
// 	} else if err != nil {
// 		return metaxgpu.Info{}, fmt.Errorf("failed to %s: %w", operationGetBiosVersion, err)
// 	}

// 	mode, ok := metaxgpu.ModeMap[info.Mode]
// 	if !ok {
// 		return metaxgpu.Info{}, fmt.Errorf("invalid gpu mode: %d", info.Mode)
// 	}

// 	var dieCount uint32
// 	if err := checkReturnCode("mxSmlGetDeviceDieCount", mxSmlGetDeviceDieCount(gpu, &dieCount)); err != nil {
// 		return metaxgpu.Info{}, err
// 	}

// 	return metaxgpu.Info{
// 		Series:      series,
// 		Model:       cString(info.DeviceName[:]),
// 		UUID:        cString(info.UUID[:]),
// 		BiosVersion: biosVersion,
// 		BDF:         cString(info.BDFId[:]),
// 		Mode:        mode,
// 		DieCount:    dieCount,
// 	}, nil
// }

// func metaxGetGpuVersion(ctx context.Context, gpu uint32, unit device.DeviceVersionUnit) (string, error) {
// 	select {
// 	case <-ctx.Done():
// 		return "", ctx.Err()
// 	default:
// 	}

// 	const versionMaximumSize = 64

// 	var (
// 		size uint32 = versionMaximumSize
// 		buf         = make([]byte, size)
// 	)
// 	if err := checkReturnCode("mxSmlGetDeviceVersion", mxSmlGetDeviceVersion(gpu, unit, &buf[0], &size)); err != nil {
// 		return "", err
// 	}

// 	return cString(buf), nil
// }

// func metaxListGpuBoardWayElectricInfos(ctx context.Context, gpu uint32) ([]device.GpuBoardWayElectricInfo, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	const maxBoardWays = 3

// 	var (
// 		size uint32 = maxBoardWays
// 		arr         = make([]SmlBoardWayElectricInfo, size)
// 	)
// 	if err := checkReturnCode("mxSmlGetBoardPowerInfo", mxSmlGetBoardPowerInfo(gpu, &size, &arr[0])); err != nil {
// 		return nil, err
// 	}

// 	actualSize := int(size)
// 	result := make([]device.GpuBoardWayElectricInfo, actualSize)

// 	for i := 0; i < actualSize; i++ {
// 		result[i] = device.GpuBoardWayElectricInfo{
// 			Voltage: arr[i].Voltage,
// 			Current: arr[i].Current,
// 			Power:   arr[i].Power,
// 		}
// 	}

// 	return result, nil
// }

// func metaxGetGpuPcieLinkInfo(ctx context.Context, gpu uint32) (device.GpuPcieLinkInfo, error) {
// 	select {
// 	case <-ctx.Done():
// 		return device.GpuPcieLinkInfo{}, ctx.Err()
// 	default:
// 	}

// 	var obj SmlPcieInfo
// 	if err := checkReturnCode("mxSmlGetPcieInfo", mxSmlGetPcieInfo(gpu, &obj)); err != nil {
// 		return device.GpuPcieLinkInfo{}, err
// 	}

// 	return device.GpuPcieLinkInfo(obj), nil
// }

// func metaxGetGpuPcieThroughputInfo(ctx context.Context, gpu uint32) (device.GpuPcieThroughputInfo, error) {
// 	select {
// 	case <-ctx.Done():
// 		return device.GpuPcieThroughputInfo{}, ctx.Err()
// 	default:
// 	}

// 	var obj SmlPcieThroughput
// 	if err := checkReturnCode("mxSmlGetPcieThroughput", mxSmlGetPcieThroughput(gpu, &obj)); err != nil {
// 		return device.GpuPcieThroughputInfo{}, err
// 	}

// 	return device.GpuPcieThroughputInfo(obj), nil
// }

// func metaxListGpuMetaxlinkLinkInfos(ctx context.Context, gpu uint32) ([]device.GpuMetaXLinkLinkInfo, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	var (
// 		size uint32 = device.MetaXLinkMaxNumber
// 		arr         = make([]SmlSingleMetaXLinkInfo, size)
// 	)
// 	if err := checkReturnCode("mxSmlGetMetaXLinkInfo_v2", mxSmlGetMetaXLinkInfo_v2(gpu, &size, &arr[0])); err != nil {
// 		return nil, err
// 	}

// 	actualSize := int(size)
// 	result := make([]device.GpuMetaXLinkLinkInfo, actualSize)

// 	for i := range actualSize {
// 		result[i] = device.GpuMetaXLinkLinkInfo{
// 			Speed: arr[i].Speed,
// 			Width: arr[i].Width,
// 		}
// 	}

// 	return result, nil
// }

// func metaxListGpuMetaxlinkThroughputInfos(ctx context.Context, gpu uint32) ([]device.GpuMetaXLinkThroughputInfo, error) {
// 	operationListMetaxlinkReceiveRates := "list metaxlink receive rates"
// 	receiveRates, err := metaxListGpuMetaxlinkThroughputParts(ctx, gpu, device.MetaXLinkTypeReceive)
// 	if IsNotSupported(err) {
// 		return nil, err
// 	} else if err != nil {
// 		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkReceiveRates, err)
// 	}

// 	operationListMetaxlinkTransmitRates := "list metaxlink transmit rates"
// 	transmitRates, err := metaxListGpuMetaxlinkThroughputParts(ctx, gpu, device.MetaXLinkTypeTransmit)
// 	if IsNotSupported(err) {
// 		return nil, err
// 	} else if err != nil {
// 		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkTransmitRates, err)
// 	}

// 	if len(receiveRates) != len(transmitRates) {
// 		return nil, fmt.Errorf("receive and transmit array length mismatch")
// 	}

// 	result := make([]device.GpuMetaXLinkThroughputInfo, len(receiveRates))

// 	for i := 0; i < len(result); i++ {
// 		result[i] = device.GpuMetaXLinkThroughputInfo{
// 			ReceiveRate:  receiveRates[i],
// 			TransmitRate: transmitRates[i],
// 		}
// 	}

// 	return result, nil
// }

// func metaxListGpuMetaxlinkThroughputParts(ctx context.Context, gpu uint32, typ device.MetaXLinkType) ([]int32, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	var (
// 		size uint32 = device.MetaXLinkMaxNumber
// 		arr         = make([]SmlMetaXLinkBandwidth, size)
// 	)
// 	if err := checkReturnCode("mxSmlGetMetaXLinkBandwidth", mxSmlGetMetaXLinkBandwidth(gpu, typ, &size, &arr[0])); err != nil {
// 		return nil, err
// 	}

// 	actualSize := int(size)
// 	result := make([]int32, actualSize)

// 	for i := 0; i < actualSize; i++ {
// 		result[i] = arr[i].RequestBandwidth
// 	}

// 	return result, nil
// }

// func metaxListGpuMetaxlinkTrafficStatInfos(ctx context.Context, gpu uint32) ([]device.GpuMetaXLinkTrafficStatInfo, error) {
// 	operationListMetaxlinkReceives := "list metaxlink receives"
// 	receives, err := metaxListGpuMetaxlinkTrafficStatParts(ctx, gpu, device.MetaXLinkTypeReceive)
// 	if IsNotSupported(err) {
// 		return nil, err
// 	} else if err != nil {
// 		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkReceives, err)
// 	}

// 	operationListMetaxlinkTransmits := "list metaxlink transmits"
// 	transmits, err := metaxListGpuMetaxlinkTrafficStatParts(ctx, gpu, device.MetaXLinkTypeTransmit)
// 	if IsNotSupported(err) {
// 		return nil, err
// 	} else if err != nil {
// 		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkTransmits, err)
// 	}

// 	if len(receives) != len(transmits) {
// 		return nil, fmt.Errorf("receive and transmit array length mismatch")
// 	}

// 	result := make([]device.GpuMetaXLinkTrafficStatInfo, len(receives))

// 	for i := 0; i < len(result); i++ {
// 		result[i] = device.GpuMetaXLinkTrafficStatInfo{
// 			Receive:  receives[i],
// 			Transmit: transmits[i],
// 		}
// 	}

// 	return result, nil
// }

// func metaxListGpuMetaxlinkTrafficStatParts(ctx context.Context, gpu uint32, typ device.MetaXLinkType) ([]int64, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	var (
// 		size uint32 = device.MetaXLinkMaxNumber
// 		arr         = make([]SmlMetaXLinkTrafficStat, size)
// 	)
// 	if err := checkReturnCode("mxSmlGetMetaXLinkTrafficStat", mxSmlGetMetaXLinkTrafficStat(gpu, typ, &size, &arr[0])); err != nil {
// 		return nil, err
// 	}

// 	actualSize := int(size)
// 	result := make([]int64, actualSize)

// 	for i := 0; i < actualSize; i++ {
// 		result[i] = arr[i].RequestTrafficStat
// 	}

// 	return result, nil
// }

// func metaxListGpuMetaxlinkAerErrorsInfos(ctx context.Context, gpu uint32) ([]device.GpuMetaXLinkAerInfo, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	var (
// 		size uint32 = device.MetaXLinkMaxNumber
// 		arr         = make([]SmlMetaXLinkAer, size)
// 	)
// 	if err := checkReturnCode("mxSmlGetMetaXLinkAer", mxSmlGetMetaXLinkAer(gpu, &size, &arr[0])); err != nil {
// 		return nil, err
// 	}

// 	actualSize := int(size)
// 	result := make([]device.GpuMetaXLinkAerInfo, actualSize)

// 	for i := 0; i < actualSize; i++ {
// 		result[i] = device.GpuMetaXLinkAerInfo{
// 			CorrectableErrorsCount:   arr[i].CorrectableErrorsCount,
// 			UncorrectableErrorsCount: arr[i].UncorrectableErrorsCount,
// 		}
// 	}

// 	return result, nil
// }

// func metaxGetDieStatus(ctx context.Context, gpu, die uint32) (int32, error) {
// 	select {
// 	case <-ctx.Done():
// 		return 0, ctx.Err()
// 	default:
// 	}

// 	var obj SmlDeviceUnavailableReasonInfo
// 	if err := checkReturnCode("mxSmlGetDieUnavailableReason", mxSmlGetDieUnavailableReason(gpu, die, &obj)); err != nil {
// 		return 0, err
// 	}

// 	return obj.unavailableCode, nil
// }

// func metaxGetDieTemperature(ctx context.Context, gpu, die uint32, sensor SmlTemperatureSensor) (float64, error) {
// 	select {
// 	case <-ctx.Done():
// 		return 0, ctx.Err()
// 	default:
// 	}

// 	var value int32
// 	if err := checkReturnCode("mxSmlGetDieTemperatureInfo", mxSmlGetDieTemperatureInfo(gpu, die, sensor, &value)); err != nil {
// 		return 0, err
// 	}

// 	return float64(value) / 100, nil
// }

// func metaxGetDieUtilization(ctx context.Context, gpu, die uint32, ip SmlUsageIp) (int32, error) {
// 	select {
// 	case <-ctx.Done():
// 		return 0, ctx.Err()
// 	default:
// 	}

// 	var value int32
// 	if err := checkReturnCode("mxSmlGetDieIpUsage", mxSmlGetDieIpUsage(gpu, die, ip, &value)); err != nil {
// 		return 0, err
// 	}

// 	return value, nil
// }

// func metaxGetDieMemoryInfo(ctx context.Context, gpu, die uint32) (device.DieMemoryInfo, error) {
// 	select {
// 	case <-ctx.Done():
// 		return device.DieMemoryInfo{}, ctx.Err()
// 	default:
// 	}

// 	var obj SmlMemoryInfo
// 	if err := checkReturnCode("mxSmlGetDieMemoryInfo", mxSmlGetDieMemoryInfo(gpu, die, &obj)); err != nil {
// 		return device.DieMemoryInfo{}, err
// 	}

// 	return device.DieMemoryInfo{
// 		Total: obj.vramTotal,
// 		Used:  obj.vramUse,
// 	}, nil
// }

// func metaxListDieClocks(ctx context.Context, gpu, die uint32, ip SmlClockIp) ([]uint32, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	const maxClocksSize = 8

// 	var (
// 		size uint32 = maxClocksSize
// 		arr         = make([]uint32, size)
// 	)
// 	if err := checkReturnCode("mxSmlGetDieClocks", mxSmlGetDieClocks(gpu, die, ip, &size, &arr[0])); err != nil {
// 		return nil, err
// 	}

// 	actualSize := int(size)
// 	result := make([]uint32, actualSize)

// 	for i := 0; i < actualSize; i++ {
// 		result[i] = arr[i]
// 	}

// 	return result, nil
// }

// func metaxGetDieClocksThrottleStatus(ctx context.Context, gpu, die uint32) (uint64, error) {
// 	select {
// 	case <-ctx.Done():
// 		return 0, ctx.Err()
// 	default:
// 	}

// 	var value uint64
// 	if err := checkReturnCode("mxSmlGetDieCurrentClocksThrottleReason", mxSmlGetDieCurrentClocksThrottleReason(gpu, die, &value)); err != nil {
// 		return 0, err
// 	}

// 	return value, nil
// }

// func metaxGetDieDpmPerformanceLevel(ctx context.Context, gpu, die uint32, ip SmlDpmIp) (uint32, error) {
// 	select {
// 	case <-ctx.Done():
// 		return 0, ctx.Err()
// 	default:
// 	}

// 	var value uint32
// 	if err := checkReturnCode("mxSmlGetCurrentDieDpmIpPerfLevel", mxSmlGetCurrentDieDpmIpPerfLevel(gpu, die, ip, &value)); err != nil {
// 		return 0, err
// 	}

// 	return value, nil
// }

// func metaxGetDieEccMemoryInfo(ctx context.Context, gpu, die uint32) (device.DieEccMemoryInfo, error) {
// 	select {
// 	case <-ctx.Done():
// 		return device.DieEccMemoryInfo{}, ctx.Err()
// 	default:
// 	}

// 	var obj SmlEccErrorCount
// 	if err := checkReturnCode("mxSmlGetDieTotalEccErrors", mxSmlGetDieTotalEccErrors(gpu, die, &obj)); err != nil {
// 		return device.DieEccMemoryInfo{}, err
// 	}

// 	return device.DieEccMemoryInfo(obj), nil
// }

func cString(bs []byte) string {
	for i, b := range bs {
		if b == 0 {
			return string(bs[:i])
		}
	}
	return string(bs)
}

func getBitsFromLsbToMsb(x uint64) []uint8 {
	size := 64
	bits := make([]uint8, size)
	for i := 0; i < size; i++ {
		bits[i] = uint8((x >> i) & 1)
	}
	return bits
}

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

package collector

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"golang.org/x/sync/errgroup"

	"huatuo-bamai/core/metrics/metax/device"
	metaxgpu "huatuo-bamai/core/metrics/metax/gpu"
	"huatuo-bamai/core/metrics/metax/sml"
	"huatuo-bamai/internal/log"
	"huatuo-bamai/pkg/metric"
	"huatuo-bamai/pkg/tracing"
)

func init() {
	tracing.RegisterEventTracing("metax_gpu", newMetaxGpuCollector)
}

type metaxGpuCollector struct{}

func newMetaxGpuCollector() (*tracing.EventTracingAttr, error) {
	// Init MetaX SML lib
	if err := sml.Init(); err != nil {
		return nil, fmt.Errorf("failed to init sml: %w", err)
	}

	return &tracing.EventTracingAttr{
		TracingData: &metaxGpuCollector{},
		Flag:        tracing.FlagMetric,
	}, nil
}

func (m *metaxGpuCollector) Update() ([]*metric.Data, error) {
	metrics, err := metaxCollectMetrics(context.Background())
	if err != nil {
		var smlError *sml.SmlError
		if errors.As(err, &smlError) {
			log.Errorf("re-initing sml and retrying because sml error: %v", err)

			if err := sml.Init(); err != nil {
				return nil, fmt.Errorf("failed to re-init sml: %w", err)
			}

			return metaxCollectMetrics(context.Background())
		}

		return nil, err
	}

	return metrics, nil
}

func metaxCollectMetrics(ctx context.Context) ([]*metric.Data, error) {
	var metrics []*metric.Data

	// SDK version
	operationGetSdkVersion := "get sdk version"
	if sdkVersion, err := sml.GetSDKVersion(ctx); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported", operationGetSdkVersion)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetSdkVersion, err)
	} else {
		metrics = append(metrics,
			metric.NewGaugeData("sdk_info", 1, "GPU SDK info.", map[string]string{
				"version": sdkVersion,
			}),
		)
	}

	var gpus []uint32

	// Native and VF GPUs
	nativeAndVfGpuCount := sml.GetNativeAndVFGPUCount()
	for i := uint32(0); i < nativeAndVfGpuCount; i++ {
		gpus = append(gpus, i)
	}

	// PF GPUs
	pfGpuCount := sml.GetPFGPUCount()
	const pfGpuIndexOffset = uint32(100)
	for i := pfGpuIndexOffset; i < pfGpuIndexOffset+pfGpuCount; i++ {
		gpus = append(gpus, i)
	}

	// Driver version
	if len(gpus) > 0 {
		operationGetDriverVersion := "get driver version"
		if driverVersion, err := sml.GetGPUVersion(ctx, gpus[0], device.DeviceVersionUnitDriver); sml.IsNotSupported(err) {
			log.Debugf("operation %s not supported on gpu 0", operationGetDriverVersion)
		} else if err != nil {
			return nil, fmt.Errorf("failed to %s: %w", operationGetDriverVersion, err)
		} else {
			metrics = append(metrics,
				metric.NewGaugeData("driver_info", 1, "GPU driver info.", map[string]string{
					"version": driverVersion,
				}),
			)
		}
	}

	// GPU
	eg, subCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	for _, gpu := range gpus {
		eg.Go(func() error {
			gpuMetrics, err := metaxCollectGpuMetrics(subCtx, gpu)
			if err != nil {
				return fmt.Errorf("failed to collect gpu %d metrics: %w", gpu, err)
			}
			mu.Lock()
			metrics = append(metrics, gpuMetrics...)
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return metrics, nil
}

/*
   GPU metrics
*/

func metaxCollectGpuMetrics(ctx context.Context, gpu uint32) ([]*metric.Data, error) {
	var metrics []*metric.Data

	// GPU info
	gpuInfo, err := sml.GetGPUInfo(ctx, gpu)
	if err != nil {
		return nil, fmt.Errorf("failed to get gpu info: %w", err)
	}
	metrics = append(metrics,
		metric.NewGaugeData("info", 1, "GPU info.", map[string]string{
			"gpu":          strconv.Itoa(int(gpu)),
			"model":        gpuInfo.Model,
			"uuid":         gpuInfo.UUID,
			"bios_version": gpuInfo.BiosVersion,
			"bdf":          gpuInfo.BDF,
			"mode":         string(gpuInfo.Mode),
			"die_count":    strconv.Itoa(int(gpuInfo.DieCount)),
		}),
	)

	// Board electric
	operationListBoardWayElectricInfos := "list board way electric infos"
	if boardWayElectricInfos, err := sml.ListGPUBoardWayElectricInfos(ctx, gpu); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d", operationListBoardWayElectricInfos, gpu)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListBoardWayElectricInfos, err)
	} else {
		var totalPower float64
		for _, info := range boardWayElectricInfos {
			totalPower += float64(info.Power)
		}

		metrics = append(metrics,
			metric.NewGaugeData("board_power_watts", totalPower/1000, "GPU board power.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
			}),
		)
	}

	// PCIe link
	operationGetPcieLinkInfo := "get pcie link info"
	if pcieLinkInfo, err := sml.GetGPUPcieLinkInfo(ctx, gpu); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d", operationGetPcieLinkInfo, gpu)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetPcieLinkInfo, err)
	} else {
		metrics = append(metrics,
			metric.NewGaugeData("pcie_link_speed_gt_per_second", float64(pcieLinkInfo.Speed), "GPU PCIe current link speed.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
			}),
			metric.NewGaugeData("pcie_link_width_lanes", float64(pcieLinkInfo.Width), "GPU PCIe current link width.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
			}),
		)
	}

	// PCIe throughput
	operationGetPcieThroughputInfo := "get pcie throughput info"
	if pcieThroughputInfo, err := sml.GetGPUPcieThroughputInfo(ctx, gpu); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d", operationGetPcieThroughputInfo, gpu)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetPcieThroughputInfo, err)
	} else {
		metrics = append(metrics,
			metric.NewGaugeData("pcie_receive_bytes_per_second", float64(pcieThroughputInfo.ReceiveRate)*1000*1000, "GPU PCIe receive throughput.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
			}),
			metric.NewGaugeData("pcie_transmit_bytes_per_second", float64(pcieThroughputInfo.TransmitRate)*1000*1000, "GPU PCIe transmit throughput.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
			}),
		)
	}

	// MetaXLink link
	operationListMetaxlinkLinkInfos := "list metaxlink link infos"
	if metaxlinkLinkInfos, err := sml.ListGPUMetaXLinkLinkInfos(ctx, gpu); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d", operationListMetaxlinkLinkInfos, gpu)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkLinkInfos, err)
	} else {
		for i, info := range metaxlinkLinkInfos {
			metrics = append(metrics,
				metric.NewGaugeData("metaxlink_link_speed_gt_per_second", float64(info.Speed), "GPU MetaXLink current link speed.", map[string]string{
					"gpu":       strconv.Itoa(int(gpu)),
					"metaxlink": strconv.Itoa(i + 1),
				}),
				metric.NewGaugeData("metaxlink_link_width_lanes", float64(info.Width), "GPU MetaXLink current link width.", map[string]string{
					"gpu":       strconv.Itoa(int(gpu)),
					"metaxlink": strconv.Itoa(i + 1),
				}),
			)
		}
	}

	// MetaXLink throughput
	operationListMetaxlinkThroughputInfos := "list metaxlink throughput infos"
	if metaxlinkThroughputInfos, err := sml.ListGPUMetaXLinkThroughputInfos(ctx, gpu); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d", operationListMetaxlinkThroughputInfos, gpu)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkThroughputInfos, err)
	} else {
		for i, info := range metaxlinkThroughputInfos {
			metrics = append(metrics,
				metric.NewGaugeData("metaxlink_receive_bytes_per_second", float64(info.ReceiveRate)*1000*1000, "GPU MetaXLink receive throughput.", map[string]string{
					"gpu":       strconv.Itoa(int(gpu)),
					"metaxlink": strconv.Itoa(i + 1),
				}),
				metric.NewGaugeData("metaxlink_transmit_bytes_per_second", float64(info.TransmitRate)*1000*1000, "GPU MetaXLink transmit throughput.", map[string]string{
					"gpu":       strconv.Itoa(int(gpu)),
					"metaxlink": strconv.Itoa(i + 1),
				}),
			)
		}
	}

	// MetaXLink traffic stat
	operationListMetaxlinkTrafficStatInfos := "list metaxlink traffic stat infos"
	if metaxlinkTrafficStatInfos, err := sml.ListGPUMetaXLinkTrafficStatInfos(ctx, gpu); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d", operationListMetaxlinkTrafficStatInfos, gpu)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkTrafficStatInfos, err)
	} else {
		for i, info := range metaxlinkTrafficStatInfos {
			metrics = append(metrics,
				metric.NewCounterData("metaxlink_receive_bytes_total", float64(info.Receive), "GPU MetaXLink receive data size.", map[string]string{
					"gpu":       strconv.Itoa(int(gpu)),
					"metaxlink": strconv.Itoa(i + 1),
				}),
				metric.NewCounterData("metaxlink_transmit_bytes_total", float64(info.Transmit), "GPU MetaXLink transmit data size.", map[string]string{
					"gpu":       strconv.Itoa(int(gpu)),
					"metaxlink": strconv.Itoa(i + 1),
				}),
			)
		}
	}

	// MetaXLink AER errors
	operationListMetaxlinkAerErrorsInfos := "list metaxlink aer errors infos"
	if metaxlinkAerErrorsInfos, err := sml.ListGPUMetaXLinkAerErrorsInfos(ctx, gpu); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d", operationListMetaxlinkAerErrorsInfos, gpu)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationListMetaxlinkAerErrorsInfos, err)
	} else {
		for i, info := range metaxlinkAerErrorsInfos {
			metrics = append(metrics,
				metric.NewCounterData("metaxlink_aer_errors_total", float64(info.CorrectableErrorsCount), "GPU MetaXLink AER errors count.", map[string]string{
					"gpu":        strconv.Itoa(int(gpu)),
					"metaxlink":  strconv.Itoa(i + 1),
					"error_type": "ce",
				}),
				metric.NewCounterData("metaxlink_aer_errors_total", float64(info.UncorrectableErrorsCount), "GPU MetaXLink AER errors count.", map[string]string{
					"gpu":        strconv.Itoa(int(gpu)),
					"metaxlink":  strconv.Itoa(i + 1),
					"error_type": "ue",
				}),
			)
		}
	}

	// Die
	eg, subCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	for die := uint32(0); die < gpuInfo.DieCount; die++ {
		eg.Go(func() error {
			dieMetrics, err := metaxCollectDieMetrics(subCtx, gpu, die, gpuInfo.Series)
			if err != nil {
				return fmt.Errorf("failed to collect die %d metrics: %w", die, err)
			}
			mu.Lock()
			metrics = append(metrics, dieMetrics...)
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return metrics, nil
}

/*
   Die metrics
*/

func metaxCollectDieMetrics(ctx context.Context, gpu, die uint32, series metaxgpu.Series) ([]*metric.Data, error) {
	var metrics []*metric.Data

	// Die status
	operationGetDieStatus := "get die status"
	if dieStatus, err := sml.GetDieStatus(ctx, gpu, die); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d die %d", operationGetDieStatus, gpu, die)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetDieStatus, err)
	} else {
		metrics = append(metrics,
			metric.NewGaugeData("status", float64(dieStatus), "GPU status, 0 means normal, other values means abnormal. Check the documentation to see the exceptions corresponding to each value.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
				"die": strconv.Itoa(int(die)),
			}),
		)
	}

	// Temperature
	operationGetTemperature := "get temperature"
	if value, err := sml.GetDieTemperature(ctx, gpu, die, sml.SmlTemperatureSensorHotspot); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d die %d", operationGetTemperature, gpu, die)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetTemperature, err)
	} else {
		metrics = append(metrics,
			metric.NewGaugeData("temperature_celsius", value, "GPU temperature.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
				"die": strconv.Itoa(int(die)),
			}),
		)
	}

	// Utilization
	for ip, ipC := range sml.GpuUtilizationIpMap {
		operationGetUtilization := fmt.Sprintf("get %s utilization", ip)
		if value, err := sml.GetDieUtilization(ctx, gpu, die, ipC); sml.IsNotSupported(err) {
			log.Debugf("operation %s not supported on gpu %d die %d", operationGetUtilization, gpu, die)
		} else if err != nil {
			return nil, fmt.Errorf("failed to %s: %w", operationGetUtilization, err)
		} else {
			metrics = append(metrics,
				metric.NewGaugeData("utilization_percent", float64(value), "GPU utilization, ranging from 0 to 100.", map[string]string{
					"gpu": strconv.Itoa(int(gpu)),
					"die": strconv.Itoa(int(die)),
					"ip":  ip,
				}),
			)
		}
	}

	// Memory
	operationGetMemoryInfo := "get memory info"
	if memoryInfo, err := sml.GetDieMemoryInfo(ctx, gpu, die); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d die %d", operationGetMemoryInfo, gpu, die)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetMemoryInfo, err)
	} else {
		metrics = append(metrics,
			metric.NewGaugeData("memory_total_bytes", float64(memoryInfo.Total)*1024, "Total vram.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
				"die": strconv.Itoa(int(die)),
			}),
			metric.NewGaugeData("memory_used_bytes", float64(memoryInfo.Used)*1024, "Used vram.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
				"die": strconv.Itoa(int(die)),
			}),
		)
	}

	// Clock
	for ip, ipC := range sml.GpuClockIpMap {
		// For metaxGpuSeriesN, use metaxSmlClockIpMc instead of metaxSmlClockIpMc0 for memory clock
		if ip == "memory" && series == metaxgpu.SeriesN {
			ipC = sml.SmlClockIpMc
		}

		operationListClocks := fmt.Sprintf("list %s clocks", ip)
		if values, err := sml.ListDieClocks(ctx, gpu, die, ipC); sml.IsNotSupported(err) {
			log.Debugf("operation %s not supported on gpu %d die %d", operationListClocks, gpu, die)
		} else if err != nil {
			return nil, fmt.Errorf("failed to %s: %w", operationListClocks, err)
		} else {
			metrics = append(metrics,
				metric.NewGaugeData("clock_mhz", float64(values[0]), "GPU clock.", map[string]string{
					"gpu": strconv.Itoa(int(gpu)),
					"die": strconv.Itoa(int(die)),
					"ip":  ip,
				}),
			)
		}
	}

	// Clocks throttle status
	operationGetClocksThrottleStatus := "get clocks throttle status"
	if clocksThrottleStatus, err := sml.GetDieClocksThrottleStatus(ctx, gpu, die); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d die %d", operationGetClocksThrottleStatus, gpu, die)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetClocksThrottleStatus, err)
	} else {
		bits := getBitsFromLsbToMsb(clocksThrottleStatus)

		for i, v := range bits {
			if v == 0 {
				// Metrics are not exported when not throttling.
				continue
			}

			bit := i + 1

			if _, ok := sml.GpuClocksThrottleBitReasonMap[bit]; !ok {
				log.Warnf("gpu %d die %d is clocks throttling for unknown reason bit %d", gpu, die, bit)
				continue
			}

			metrics = append(metrics,
				metric.NewGaugeData("clocks_throttling", float64(v), "Reason(s) for GPU clocks throttling.", map[string]string{
					"gpu":    strconv.Itoa(int(gpu)),
					"die":    strconv.Itoa(int(die)),
					"reason": sml.GpuClocksThrottleBitReasonMap[bit],
				}),
			)
		}
	}

	// DPM performance level
	for ip, ipC := range sml.GpuDpmIpMap {
		operationGetDpmPerformanceLevel := fmt.Sprintf("get %s dpm performance level", ip)
		if value, err := sml.GetDieDPMPerformanceLevel(ctx, gpu, die, ipC); sml.IsNotSupported(err) {
			log.Debugf("operation %s not supported on gpu %d die %d", operationGetDpmPerformanceLevel, gpu, die)
		} else if err != nil {
			return nil, fmt.Errorf("failed to %s: %w", operationGetDpmPerformanceLevel, err)
		} else {
			metrics = append(metrics,
				metric.NewGaugeData("dpm_performance_level", float64(value), "GPU DPM performance level.", map[string]string{
					"gpu": strconv.Itoa(int(gpu)),
					"die": strconv.Itoa(int(die)),
					"ip":  ip,
				}),
			)
		}
	}

	// Ecc memory
	operationGetEccMemoryInfo := "get ecc memory info"
	if eccMemoryInfo, err := sml.GetDieECCMemoryInfo(ctx, gpu, die); sml.IsNotSupported(err) {
		log.Debugf("operation %s not supported on gpu %d die %d", operationGetEccMemoryInfo, gpu, die)
	} else if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", operationGetEccMemoryInfo, err)
	} else {
		metrics = append(metrics,
			metric.NewCounterData("ecc_memory_errors_total", float64(eccMemoryInfo.SramCorrectableErrorsCount), "GPU ECC memory errors count.", map[string]string{
				"gpu":         strconv.Itoa(int(gpu)),
				"die":         strconv.Itoa(int(die)),
				"memory_type": "sram",
				"error_type":  "ce",
			}),
			metric.NewCounterData("ecc_memory_errors_total", float64(eccMemoryInfo.SramUncorrectableErrorsCount), "GPU ECC memory errors count.", map[string]string{
				"gpu":         strconv.Itoa(int(gpu)),
				"die":         strconv.Itoa(int(die)),
				"memory_type": "sram",
				"error_type":  "ue",
			}),
			metric.NewCounterData("ecc_memory_errors_total", float64(eccMemoryInfo.DramCorrectableErrorsCount), "GPU ECC memory errors count.", map[string]string{
				"gpu":         strconv.Itoa(int(gpu)),
				"die":         strconv.Itoa(int(die)),
				"memory_type": "dram",
				"error_type":  "ce",
			}),
			metric.NewCounterData("ecc_memory_errors_total", float64(eccMemoryInfo.DramUncorrectableErrorsCount), "GPU ECC memory errors count.", map[string]string{
				"gpu":         strconv.Itoa(int(gpu)),
				"die":         strconv.Itoa(int(die)),
				"memory_type": "dram",
				"error_type":  "ue",
			}),
			metric.NewCounterData("ecc_memory_retired_pages_total", float64(eccMemoryInfo.RetiredPagesCount), "GPU ECC memory retired pages count.", map[string]string{
				"gpu": strconv.Itoa(int(gpu)),
				"die": strconv.Itoa(int(die)),
			}),
		)
	}

	return metrics, nil
}

/*
   Util
*/

func getBitsFromLsbToMsb(x uint64) []uint8 {
	size := 64
	bits := make([]uint8, size)
	for i := 0; i < size; i++ {
		bits[i] = uint8((x >> i) & 1)
	}
	return bits
}

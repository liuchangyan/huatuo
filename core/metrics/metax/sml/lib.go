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
	"fmt"
	"runtime"
	"sync"

	"github.com/ebitengine/purego"
)

// library represents the SML shared library (purego version)
type library struct {
	sync.Mutex
	path     string
	handle   uintptr
	refcount refcount
}

// global singleton (like libnvml)
var libsml = newLibrary()

func newLibrary() *library {
	return &library{
		path: defaultSmlLibraryPath(),
	}
}

/*
 * path helpers
 */

func defaultSmlLibraryPath() string {
	switch runtime.GOOS {
	case "linux":
		return "/opt/mxdriver/lib/libmxsml.so"
	default:
		return ""
	}
}

/*
 * lifecycle
 */

func (l *library) load() (rerr error) {
	l.Lock()
	defer l.Unlock()
	defer func() { l.refcount.IncOnNoError(rerr) }()

	if l.refcount > 0 {
		return nil
	}

	if l.path == "" {
		return fmt.Errorf("GOOS=%s is not supported", runtime.GOOS)
	}

	handle, err := purego.Dlopen(
		l.path,
		purego.RTLD_NOW|purego.RTLD_GLOBAL,
	)
	if err != nil {
		return fmt.Errorf("dlopen %s failed: %w", l.path, err)
	}

	l.handle = handle

	// register all symbols
	registerSmlLibSymbols(handle)

	return nil
}

func (l *library) close() (rerr error) {
	l.Lock()
	defer l.Unlock()
	defer func() { l.refcount.DecOnNoError(rerr) }()

	if l.refcount != 1 {
		return nil
	}

	if err := purego.Dlclose(l.handle); err != nil {
		return err
	}

	l.handle = 0
	return nil
}

/*
 * symbol registration
 */

func registerSmlLibSymbols(handle uintptr) {
	purego.RegisterLibFunc(&mxSmlInit, handle, "mxSmlInit")
	purego.RegisterLibFunc(&mxSmlGetErrorString, handle, "mxSmlGetErrorString")
	purego.RegisterLibFunc(&mxSmlGetMacaVersion, handle, "mxSmlGetMacaVersion")
	purego.RegisterLibFunc(&mxSmlGetDeviceCount, handle, "mxSmlGetDeviceCount")
	purego.RegisterLibFunc(&mxSmlGetPfDeviceCount, handle, "mxSmlGetPfDeviceCount")
	purego.RegisterLibFunc(&mxSmlGetDeviceInfo, handle, "mxSmlGetDeviceInfo")
	purego.RegisterLibFunc(&mxSmlGetDeviceDieCount, handle, "mxSmlGetDeviceDieCount")
	purego.RegisterLibFunc(&mxSmlGetDeviceVersion, handle, "mxSmlGetDeviceVersion")
	purego.RegisterLibFunc(&mxSmlGetBoardPowerInfo, handle, "mxSmlGetBoardPowerInfo")
	purego.RegisterLibFunc(&mxSmlGetPcieInfo, handle, "mxSmlGetPcieInfo")
	purego.RegisterLibFunc(&mxSmlGetPcieThroughput, handle, "mxSmlGetPcieThroughput")
	purego.RegisterLibFunc(&mxSmlGetMetaXLinkInfo_v2, handle, "mxSmlGetMetaXLinkInfo_v2")
	purego.RegisterLibFunc(&mxSmlGetMetaXLinkBandwidth, handle, "mxSmlGetMetaXLinkBandwidth")
	purego.RegisterLibFunc(&mxSmlGetMetaXLinkTrafficStat, handle, "mxSmlGetMetaXLinkTrafficStat")
	purego.RegisterLibFunc(&mxSmlGetMetaXLinkAer, handle, "mxSmlGetMetaXLinkAer")
	purego.RegisterLibFunc(&mxSmlGetDieUnavailableReason, handle, "mxSmlGetDieUnavailableReason")
	purego.RegisterLibFunc(&mxSmlGetDieTemperatureInfo, handle, "mxSmlGetDieTemperatureInfo")
	purego.RegisterLibFunc(&mxSmlGetDieIpUsage, handle, "mxSmlGetDieIpUsage")
	purego.RegisterLibFunc(&mxSmlGetDieMemoryInfo, handle, "mxSmlGetDieMemoryInfo")
	purego.RegisterLibFunc(&mxSmlGetDieClocks, handle, "mxSmlGetDieClocks")
	purego.RegisterLibFunc(&mxSmlGetDieCurrentClocksThrottleReason, handle, "mxSmlGetDieCurrentClocksThrottleReason")
	purego.RegisterLibFunc(&mxSmlGetCurrentDieDpmIpPerfLevel, handle, "mxSmlGetCurrentDieDpmIpPerfLevel")
	purego.RegisterLibFunc(&mxSmlGetDieTotalEccErrors, handle, "mxSmlGetDieTotalEccErrors")
}

// func metaxGetSmlLibraryPath() (string, error) {
// 	switch runtime.GOOS {
// 	case "linux":
// 		return "/opt/mxdriver/lib/libmxsml.so", nil
// 	default:
// 		return "", fmt.Errorf("GOOS=%s is not supported", runtime.GOOS)
// 	}
// }

// func metaxLoadSmlLibrary() (uintptr, error) {
// 	path, err := metaxGetSmlLibraryPath()
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to get sml library path: %w", err)
// 	}

// 	libc, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to open sml library: %w", err)
// 	}

// 	metaxRegisterSmlLibraryFunctions(libc)

// 	return libc, nil
// }

// func metaxRegisterSmlLibraryFunctions(libc uintptr) {
// 	purego.RegisterLibFunc(&mxSmlInit, libc, "mxSmlInit")
// 	purego.RegisterLibFunc(&mxSmlGetErrorString, libc, "mxSmlGetErrorString")
// 	purego.RegisterLibFunc(&mxSmlGetMacaVersion, libc, "mxSmlGetMacaVersion")
// 	purego.RegisterLibFunc(&mxSmlGetDeviceCount, libc, "mxSmlGetDeviceCount")
// 	purego.RegisterLibFunc(&mxSmlGetPfDeviceCount, libc, "mxSmlGetPfDeviceCount")
// 	purego.RegisterLibFunc(&mxSmlGetDeviceInfo, libc, "mxSmlGetDeviceInfo")
// 	purego.RegisterLibFunc(&mxSmlGetDeviceDieCount, libc, "mxSmlGetDeviceDieCount")
// 	purego.RegisterLibFunc(&mxSmlGetDeviceVersion, libc, "mxSmlGetDeviceVersion")
// 	purego.RegisterLibFunc(&mxSmlGetBoardPowerInfo, libc, "mxSmlGetBoardPowerInfo")
// 	purego.RegisterLibFunc(&mxSmlGetPcieInfo, libc, "mxSmlGetPcieInfo")
// 	purego.RegisterLibFunc(&mxSmlGetPcieThroughput, libc, "mxSmlGetPcieThroughput")
// 	purego.RegisterLibFunc(&mxSmlGetMetaXLinkInfo_v2, libc, "mxSmlGetMetaXLinkInfo_v2")
// 	purego.RegisterLibFunc(&mxSmlGetMetaXLinkBandwidth, libc, "mxSmlGetMetaXLinkBandwidth")
// 	purego.RegisterLibFunc(&mxSmlGetMetaXLinkTrafficStat, libc, "mxSmlGetMetaXLinkTrafficStat")
// 	purego.RegisterLibFunc(&mxSmlGetMetaXLinkAer, libc, "mxSmlGetMetaXLinkAer")
// 	purego.RegisterLibFunc(&mxSmlGetDieUnavailableReason, libc, "mxSmlGetDieUnavailableReason")
// 	purego.RegisterLibFunc(&mxSmlGetDieTemperatureInfo, libc, "mxSmlGetDieTemperatureInfo")
// 	purego.RegisterLibFunc(&mxSmlGetDieIpUsage, libc, "mxSmlGetDieIpUsage")
// 	purego.RegisterLibFunc(&mxSmlGetDieMemoryInfo, libc, "mxSmlGetDieMemoryInfo")
// 	purego.RegisterLibFunc(&mxSmlGetDieClocks, libc, "mxSmlGetDieClocks")
// 	purego.RegisterLibFunc(&mxSmlGetDieCurrentClocksThrottleReason, libc, "mxSmlGetDieCurrentClocksThrottleReason")
// 	purego.RegisterLibFunc(&mxSmlGetCurrentDieDpmIpPerfLevel, libc, "mxSmlGetCurrentDieDpmIpPerfLevel")
// 	purego.RegisterLibFunc(&mxSmlGetDieTotalEccErrors, libc, "mxSmlGetDieTotalEccErrors")
// }

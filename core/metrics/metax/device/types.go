package device

type Brand uint32

const (
	BrandUnknown Brand = iota
	BrandN
	BrandC
	BrandG
)

type VirtualizationMode uint32

const (
	VirtualizationModeNone VirtualizationMode = iota
	VirtualizationModePf
	VirtualizationModeVf
)

type Info struct {
	DeviceId   uint32
	_          uint32 // DEPRECATED
	BDFId      [32]byte
	GpuId      uint32
	NodeId     uint32
	UUID       [96]byte
	Brand      Brand
	Mode       VirtualizationMode
	DeviceName [32]byte
}

type DeviceVersionUnit uint32

const (
	DeviceVersionUnitBios DeviceVersionUnit = iota
	DeviceVersionUnitDriver
)

type BoardWayElectricInfo struct {
	Voltage uint32 // voltage in mV.
	Current uint32 // current in mA.
	Power   uint32 // power in mW.
}

type PcieLinkInfo struct {
	Speed float32 // speed in GT/s.
	Width uint32  // width in lanes.
}

type PcieThroughputInfo struct {
	ReceiveRate  int32 // receiveRate in MB/s.
	TransmitRate int32 // transmitRate in MB/s.
}

/*
   MetaXLink
*/

const MetaXLinkMaxNumber = 7

type MetaXLinkType uint32

const (
	MetaXLinkTypeReceive MetaXLinkType = iota
	MetaXLinkTypeTransmit
)

type MetaXLinkLinkInfo struct {
	Speed float32 // speed in GT/s.
	Width uint32  // width in lanes.
}

type MetaXLinkThroughputInfo struct {
	ReceiveRate  int32 // receiveRate in MB/s.
	TransmitRate int32 // transmitRate in MB/s.
}

type MetaXLinkTrafficStatInfo struct {
	Receive  int64 // receive in bytes.
	Transmit int64 // transmit in bytes.
}

type MetaXLinkAerInfo struct {
	CorrectableErrorsCount   int32
	UncorrectableErrorsCount int32
}

type DieMemoryInfo struct {
	Total int64 // total in KB.
	Used  int64 // used in KB.
}

type DieEccMemoryInfo struct {
	SramCorrectableErrorsCount   uint32
	SramUncorrectableErrorsCount uint32
	DramCorrectableErrorsCount   uint32
	DramUncorrectableErrorsCount uint32
	RetiredPagesCount            uint32
}

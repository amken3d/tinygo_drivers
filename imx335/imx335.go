// Package imx335 provides a driver for the Sony IMX335 5 MP rolling-shutter
// CMOS image sensor with CSI-2 output.
//
// Native output is RAW10 Bayer at up to 2592×1944. This driver covers only
// the I²C control path (register init, stream-on/off, test pattern). The
// pixel pipeline — CSI-2 host + DCMIPP / equivalent — lives in MCU-specific
// code that consumes the sensor's CSI output.
//
// The register tables and most of the public API are direct ports of
// STMicroelectronics's stm32-mw-camera/sensors/imx335 BSP component
// (BSD-3 licensed, © 2022 STMicroelectronics).
//
// Datasheet: Sony IMX335 (proprietary; values in the register tables are
// chip-internal magic — do not edit without consulting the datasheet).
//
// Typical usage:
//
//	bus.Configure(machine.I2CConfig{Frequency: 100 * machine.KHz})
//	dev := imx335.New(bus)
//	if id, err := dev.ReadID(); err != nil || id != 0 {
//	    // sensor not responding (expected ID is 0x00)
//	}
//	dev.SetFrequency(imx335.InckMHz24)
//	dev.Init()                          // 2592×1944 RAW10, 2-lane 10-bit
//	dev.SetFramerate(30)
//	dev.SetTestPattern(imx335.TPGVerticalBars)
//	dev.Start()                         // streaming begins
package imx335

import "tinygo.org/x/drivers"

// Address is the default 7-bit I²C address. The 8-bit value commonly quoted
// (0x34) is this shifted left by one.
const Address = 0x1A

// Sensor register addresses.
const (
	regModeSelect = 0x3000 // bit 0: 0 = streaming, 1 = standby
	regHold       = 0x3001 // group-hold for atomic register updates
	regVMAX       = 0x3030 // vertical-period reg (frame rate)
	regSHUTTER    = 0x3058 // shutter (3-byte little-endian)
	regGain       = 0x30E8 // analog gain
	regTPG        = 0x329E // test-pattern selector
	regID         = 0x3912 // chip-ID register

	chipIDExpected = 0x00 // value at regID per ST's reg header

	modeStreaming = 0x00
	modeStandby   = 0x01
)

// InckMHzN selects the input clock the sensor's PLL is fed.
const (
	InckMHz6  = 0
	InckMHz18 = 1
	InckMHz24 = 2
	InckMHz27 = 3
	InckMHz74 = 4
)

// Test pattern modes (write to regTPG, then enable via the pattern-enable
// register table). TPGDisabled disables the test-pattern generator.
const (
	TPGDisabled       int8 = -1
	TPGAll000         int8 = 0  // all 000h
	TPGAllFFF         int8 = 1  // all FFFh
	TPGAll555         int8 = 2  // all 555h
	TPGAllAAA         int8 = 3  // all AAAh
	TPGToggle555AAA   int8 = 4  // toggle 555/AAA
	TPGToggleAAA555   int8 = 5  // toggle AAA/555
	TPGToggle000555   int8 = 6  // toggle 000/555
	TPGToggle555000   int8 = 7  // toggle 555/000
	TPGToggle000FFF   int8 = 8  // toggle 000/FFF
	TPGToggleFFF000   int8 = 9  // toggle FFF/000
	TPGHorizontalBars int8 = 10 // horizontal color bars
	TPGVerticalBars   int8 = 11 // vertical color bars
)

type reg struct {
	addr uint16
	val  uint8
}

// res2592x1944Regs — bring-up register sequence for 2592×1944 RAW10 mode.
// Verbatim from ST's imx335.c res_2592_1944_regs[].
var res2592x1944Regs = [...]reg{
	{0x3000, 0x01}, {0x3002, 0x00}, {0x300C, 0x3B}, {0x300D, 0x2A},
	{0x3018, 0x04}, {0x302C, 0x3C}, {0x302E, 0x20}, {0x3056, 0x98},
	{0x3074, 0xC8}, {0x3076, 0x30}, {0x304C, 0x00}, {0x314C, 0xC6},
	{0x315A, 0x02}, {0x3168, 0xA0}, {0x316A, 0x7E}, {0x31A1, 0x00},
	{0x3288, 0x21}, {0x328A, 0x02}, {0x3414, 0x05}, {0x3416, 0x18},
	{0x3648, 0x01}, {0x364A, 0x04}, {0x364C, 0x04}, {0x3678, 0x01},
	{0x367C, 0x31}, {0x367E, 0x31}, {0x3706, 0x10}, {0x3708, 0x03},
	{0x3714, 0x02}, {0x3715, 0x02}, {0x3716, 0x01}, {0x3717, 0x03},
	{0x371C, 0x3D}, {0x371D, 0x3F}, {0x372C, 0x00}, {0x372D, 0x00},
	{0x372E, 0x46}, {0x372F, 0x00}, {0x3730, 0x89}, {0x3731, 0x00},
	{0x3732, 0x08}, {0x3733, 0x01}, {0x3734, 0xFE}, {0x3735, 0x05},
	{0x3740, 0x02}, {0x375D, 0x00}, {0x375E, 0x00}, {0x375F, 0x11},
	{0x3760, 0x01}, {0x3768, 0x1B}, {0x3769, 0x1B}, {0x376A, 0x1B},
	{0x376B, 0x1B}, {0x376C, 0x1A}, {0x376D, 0x17}, {0x376E, 0x0F},
	{0x3776, 0x00}, {0x3777, 0x00}, {0x3778, 0x46}, {0x3779, 0x00},
	{0x377A, 0x89}, {0x377B, 0x00}, {0x377C, 0x08}, {0x377D, 0x01},
	{0x377E, 0x23}, {0x377F, 0x02}, {0x3780, 0xD9}, {0x3781, 0x03},
	{0x3782, 0xF5}, {0x3783, 0x06}, {0x3784, 0xA5}, {0x3788, 0x0F},
	{0x378A, 0xD9}, {0x378B, 0x03}, {0x378C, 0xEB}, {0x378D, 0x05},
	{0x378E, 0x87}, {0x378F, 0x06}, {0x3790, 0xF5}, {0x3792, 0x43},
	{0x3794, 0x7A}, {0x3796, 0xA1}, {0x37B0, 0x36}, {0x3A00, 0x01},
}

// mode2L10BRegs — 2-lane 10-bit CSI mode (after resolution table).
var mode2L10BRegs = [...]reg{
	{0x3050, 0x00},
	{0x319D, 0x00},
	{0x341C, 0xFF},
	{0x341D, 0x01},
	{0x3A01, 0x01},
}

var inck24MhzRegs = [...]reg{
	{0x300C, 0x3B}, {0x300D, 0x2A}, {0x314C, 0xC6}, {0x314D, 0x00},
	{0x315A, 0x02}, {0x3168, 0xA0}, {0x316A, 0x7E},
}

var inck27MhzRegs = [...]reg{
	{0x300C, 0x42}, {0x300D, 0x2E}, {0x314C, 0xB0}, {0x314D, 0x00},
	{0x315A, 0x02}, {0x3168, 0x8F}, {0x316A, 0x7E},
}

var inck18MhzRegs = [...]reg{
	{0x300C, 0x2D}, {0x300D, 0x1F}, {0x314C, 0x84}, {0x314D, 0x00},
	{0x315A, 0x01}, {0x3168, 0x6B}, {0x316A, 0x7D},
}

var inck6MhzRegs = [...]reg{
	{0x300C, 0x0F}, {0x300D, 0x0B}, {0x314C, 0xC6}, {0x314D, 0x00},
	{0x315A, 0x00}, {0x3168, 0xA0}, {0x316A, 0x7C},
}

var inck74MhzRegs = [...]reg{
	{0x300C, 0xB6}, {0x300D, 0x7F}, {0x314C, 0x80}, {0x314D, 0x00},
	{0x315A, 0x03}, {0x3168, 0x68}, {0x316A, 0x7F},
}

var (
	framerate10Regs = [...]reg{{0x3030, 0xC0}, {0x3031, 0x34}}
	framerate15Regs = [...]reg{{0x3030, 0x2A}, {0x3031, 0x23}}
	framerate20Regs = [...]reg{{0x3030, 0x60}, {0x3031, 0x1A}}
	framerate25Regs = [...]reg{{0x3030, 0x1A}, {0x3031, 0x15}}
	framerate30Regs = [...]reg{{0x3030, 0x94}, {0x3031, 0x11}}
)

var tpgEnableRegs = [...]reg{
	{0x3148, 0x10}, {0x3280, 0x00}, {0x329C, 0x01}, {0x32A0, 0x11},
	{0x3302, 0x00}, {0x3303, 0x00}, {0x336C, 0x00},
}

var tpgDisableRegs = [...]reg{
	{0x3148, 0x00}, {0x3280, 0x01}, {0x329C, 0x00}, {0x32A0, 0x10},
	{0x3302, 0x32}, {0x3303, 0x00}, {0x336C, 0x01},
}

// Device wraps an I²C connection to an IMX335.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// New creates a new IMX335 driver bound to the given I²C bus, using the
// default address. The bus must already be configured. This function only
// creates the Device value; it does not touch the chip.
func New(bus drivers.I2C) Device {
	return Device{bus: bus, Address: Address}
}

// ReadReg reads one byte from a 16-bit-addressed register.
func (d *Device) ReadReg(r uint16) (uint8, error) {
	var data [1]byte
	err := d.bus.Tx(d.Address, []byte{byte(r >> 8), byte(r)}, data[:])
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

// WriteReg writes one byte to a 16-bit-addressed register.
func (d *Device) WriteReg(r uint16, val uint8) error {
	return d.bus.Tx(d.Address, []byte{byte(r >> 8), byte(r), val}, nil)
}

// ReadID reads the chip-ID register. Per ST's reg header the IMX335 returns
// 0x00 in this register.
func (d *Device) ReadID() (uint8, error) {
	return d.ReadReg(regID)
}

// SetFrequency programs the sensor's PLL for the given INCK frequency
// (use the InckMHzN constants). Defaults to 6 MHz on unknown values.
func (d *Device) SetFrequency(freq int) error {
	var regs []reg
	switch freq {
	case InckMHz74:
		regs = inck74MhzRegs[:]
	case InckMHz27:
		regs = inck27MhzRegs[:]
	case InckMHz24:
		regs = inck24MhzRegs[:]
	case InckMHz18:
		regs = inck18MhzRegs[:]
	default:
		regs = inck6MhzRegs[:]
	}
	return d.writeTable(regs)
}

// Init writes the 2592×1944 RAW10 resolution table followed by the 2-lane
// 10-bit CSI mode table. Call SetFrequency first if your INCK isn't 24 MHz.
func (d *Device) Init() error {
	if err := d.writeTable(res2592x1944Regs[:]); err != nil {
		return err
	}
	return d.writeTable(mode2L10BRegs[:])
}

// SetFramerate selects 10/15/20/25/30 fps. Defaults to 30 on unknown values.
func (d *Device) SetFramerate(fps int) error {
	var regs []reg
	switch fps {
	case 10:
		regs = framerate10Regs[:]
	case 15:
		regs = framerate15Regs[:]
	case 20:
		regs = framerate20Regs[:]
	case 25:
		regs = framerate25Regs[:]
	default:
		regs = framerate30Regs[:]
	}
	return d.writeTable(regs)
}

// Start releases the sensor from standby — streaming begins.
func (d *Device) Start() error {
	return d.WriteReg(regModeSelect, modeStreaming)
}

// Stop puts the sensor back into standby (streaming halted, power reduced).
func (d *Device) Stop() error {
	return d.WriteReg(regModeSelect, modeStandby)
}

// SetTestPattern enables the sensor's internal test-pattern generator
// (mode 0..11; see TPG* constants) or disables it (mode TPGDisabled).
// Useful for verifying the CSI/DCMIPP path with no optics or scene.
func (d *Device) SetTestPattern(mode int8) error {
	if mode < 0 {
		return d.writeTable(tpgDisableRegs[:])
	}
	if err := d.WriteReg(regTPG, uint8(mode)); err != nil {
		return err
	}
	return d.writeTable(tpgEnableRegs[:])
}

func (d *Device) writeTable(t []reg) error {
	for _, e := range t {
		if err := d.WriteReg(e.addr, e.val); err != nil {
			return err
		}
	}
	return nil
}

# TMC5160 Driver for Go (TinyGo)

This repository provides a Go-based driver for the **TMC5160** stepper motor driver, implemented for both **SPI** and **UART** communication modes. The driver allows you to easily interface with the TMC5160 to configure and control stepper motors.

## Table of Contents

- [Installation](#installation)
- [Communication Modes](#communication-modes)
    - [SPI Mode](#spi-mode)
    - [UART Mode](#uart-mode)
- [Usage Example](#usage-example)
    - [Setting and Getting Modes](#setting-and-getting-modes)
    - [Reading and Writing Registers](#reading-and-writing-registers)
- [API Reference](#api-reference)
- [License](#license)

## Installation

To use the TMC5160 driver, you'll need to have **TinyGo** installed. You can install TinyGo by following the [official installation guide](https://tinygo.org/getting-started/).

### Dependencies

- **machine**: To interface with hardware on platforms like Raspberry Pi, STM32, etc.
- **TinyGo**: A Go compiler for embedded systems.

Add the module

```bash
go get github.com/amken3d/tinygo_tmc5160/tmc5160
```
Alternate method is to use the tinygo official drivers repo

```aiignore
import "tinygo.org/x/drivers/tmc5160"
```
### Communication Modes

The TMC5160 supports two communication modes for controlling the motor:

**SPI Mode**

To communicate with the TMC5160 in SPI mode, you'll need to configure the SPI bus and the chip-select (CS) pin. This allows full-speed communication between your microcontroller and the TMC5160.
SPI Setup

In SPI mode, you must configure the SPI interface on your microcontroller. Here's how to set up SPI communication for the TMC5160.

```go
spi := machine.SPI1
csPin := machine.GPIO13
spi.Configure(machine.SPIConfig{
SCK:       machine.GPIO10,
SDI:       machine.GPIO11,
SDO:       machine.GPIO12,
Frequency: 5000000,
Mode:      3,
LSBFirst:  false,
})

csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
```
**Sending Commands via SPI**

The driver supports reading and writing registers using the SPIComm interface, which is initialized with the configured SPI bus and CS pins

```go
comm := tmc5160.NewSPIComm(*spi, csPins)
driver := tmc5160.NewTMC5160(comm, driverIndex)
driver.WriteRegister(tmc5160.GCONF, value)

```

**UART Mode**

Alternatively, you can use UART mode to communicate with the TMC5160. UART mode is useful for cases where SPI is not available or when the TMC5160 is used in multi-driver configurations with limited SPI pins.
UART Setup

In UART mode, you will need to configure the UART interface with the appropriate baud rate and settings:

```go
uart := machine.UART0
uart.Configure(machine.UARTConfig{
    BaudRate: 115200,
})
```
#### Sending Commands via UART

The UART communication is handled through the UARTComm struct, which wraps the UART interface.

```go
comm := tmc5160.NewUARTComm(uart, 0x01)
driver := tmc5160.NewTMC5160(comm, 0)
driver.WriteRegister(tmc5160.GCONF, 0x01)
```

## Usage Example

Here’s a simple example of how to use the TMC5160 driver with SPI and UART modes:

```aiignore
package main

import (
    "fmt"
    "machine"
    "time"
    "tmc5160"
)

func main() {
    // SPI setup
    spi := machine.SPI1
    csPin := machine.GPIO13
    spi.Configure(machine.SPIConfig{
        SCK:       machine.GPIO10,
        SDI:       machine.GPIO11,
        SDO:       machine.GPIO12,
        Frequency: 5000000,
        Mode:      3,
        LSBFirst:  false,
    })

    csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
    csPins := map[uint8]machine.Pin{0: csPin}

    comm := tmc5160.NewSPIComm(*spi, csPins)
    driver := tmc5160.NewTMC5160(comm, 0)

    // Setting and getting mode
    rampMode := tmc5160.NewRAMPMODE(comm)
    rampMode.SetMode(tmc5160.PositioningMode)
    mode, err := rampMode.GetMode()
    if err != nil {
        fmt.Println("Error getting mode:", err)
    } else {
        fmt.Println("Current Mode:", mode)
    }

    // Read GCONF register
    GCONF := tmc5160.NewGCONF()
    gconfVal, err := driver.ReadRegister(tmc5160.GCONF)
    GCONF.Unpack(gconfVal)
    fmt.Println("GCONF:", GCONF)
}

```
## Reading and Writing Registers

You can easily read and write registers using the WriteRegister and ReadRegister methods:

```aiignore
// Write a value to a register
err := driver.WriteRegister(tmc5160.GCONF, 0x01)
if err != nil {
    fmt.Println("Error writing register:", err)
}

// Read a register
value, err := driver.ReadRegister(tmc5160.GCONF)
if err != nil {
    fmt.Println("Error reading register:", err)
} else {
    fmt.Println("Read value from GCONF:", value)
}

```

## API Reference

    NewSPIComm(spi machine.SPI, csPins map[uint8]machine.Pin) *SPIComm

Creates a new SPI communication interface for the TMC5160.

    NewUARTComm(uart machine.UART, address uint8) *UARTComm

Creates a new UART communication interface for the TMC5160.

    NewTMC5160(comm RegisterComm, address uint8) *TMC5160

Creates a new instance of the TMC5160 driver.

    WriteRegister(register uint8, value uint32) error

Writes a value to the specified register.

    ReadRegister(register uint8) (uint32, error)

Reads a value from the specified register.

## License

This project is licensed under the MIT License - see the LICENSE file for details.


### Key Sections:

1. **Installation**: Explains how to install TinyGo and clone the repository.
2. **Communication Modes**: Describes how to set up SPI and UART communication with the TMC5160.
3. **Usage Example**: Provides an example of how to use the driver, including how to set modes and read/write registers.
4. **API Reference**: Lists functions available in the TMC5160 driver, such as `WriteRegister`, `ReadRegister`, and constructors like `NewSPIComm`.

This README provides a comprehensive overview and guide on how to use the TMC5160 driver in both SPI and UART communication modes.

package tmc2209

import "log"

func CalculateCRC(data []byte) uint8 {
	crc := uint8(0)

	// MSB-first CRC-8 with polynomial 0x07, initial value 0x00
	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if (crc & 0x80) != 0 {
				crc = (crc << 1) ^ 0x07
			} else {
				crc <<= 1
			}
		}
	}

	// Bit-reverse final CRC (matches reference implementation)
	crc = ((crc >> 1) & 0x55) | ((crc & 0x55) << 1) // swap odd/even bits
	crc = ((crc >> 2) & 0x33) | ((crc & 0x33) << 2) // swap consecutive pairs
	crc = ((crc >> 4) & 0x0F) | ((crc & 0x0F) << 4) // swap nibbles

	return crc
}

// VerifyCommunication checks the communication with the TMC2209 by reading the version register (IOIN).
// It returns true if the communication is successful (i.e., the version matches the expected version).
// VerifyCommunication verifies the communication with the TMC2209 by reading the version register (IOIN).
// It explicitly resets the struct and de-references it after the check to ensure memory is managed manually.
func VerifyCommunication(comm RegisterComm, driverIndex uint8) bool {
	var io *Ioin
	if io == nil {
		io = NewIoin() // Initialize the struct if not already initialized
	} else {
		*io = Ioin{}
	}
	_, err := ReadRegister(comm, driverIndex, io.GetAddress())
	if err != nil {
		return false
	}
	if io.Version == expectedVersion {
		io = nil
		return true
	}
	io = nil
	return false
}

// CheckErrorStatus verifies the communication and checks for error flags in the TMC2209 driver status.
// It explicitly resets the struct and de-references it when done to ensure memory is managed manually.
func CheckErrorStatus(comm RegisterComm, driverIndex uint8) bool {
	var d *DrvStatus
	if d == nil {
		d = NewDrvStatus()
	} else {
		*d = DrvStatus{}
	}
	_, err := d.Read(comm, driverIndex)
	if err != nil {
		return false
	}
	errorFlags := d.Ola | d.S2vsa | d.S2vsb | d.Ot | d.S2ga | d.S2gb | d.Olb
	if errorFlags != 0 {
		log.Printf("TMC2209 Error Detected: %X", errorFlags)
		return false
	}
	d = nil
	return true
}

// GetInterfaceTransmissionCount reads the IFCNT register to check for UART transmission status
func GetInterfaceTransmissionCount(comm RegisterComm, driverIndex uint8) (uint32, error) {
	ifcnt := NewIfcnt()
	_, err := ReadRegister(comm, driverIndex, ifcnt.GetAddress())
	if err != nil {
		return 0, err
	}
	return ifcnt.Bytes, nil
}

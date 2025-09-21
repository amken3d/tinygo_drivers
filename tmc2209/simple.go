package tmc2209

// InitMotor
// Configures the registers with the right settings that are needed for rotating the motor.
// E.g Enabling driver, setting IRUN current etc.
// Note: Step pulses (S/D) are generated externally by the MCU.
func (driver *TMC2209) InitMotor() error {
	// GCONF = 0x00000040
	if err := driver.WriteRegister(GCONF, 0x00000040); err != nil {
		return err
	}
	// IHOLD_IRUN = 0x00071703
	if err := driver.WriteRegister(IHOLD_IRUN, 0x00071703); err != nil {
		return err
	}
	// TPOWERDOWN = 0x00000014
	if err := driver.WriteRegister(TPOWERDOWN, 0x00000014); err != nil {
		return err
	}
	// CHOPCONF = 0x10000053
	if err := driver.WriteRegister(CHOPCONF, 0x10000053); err != nil {
		return err
	}
	// PWMCONF = 0xC10D0024
	if err := driver.WriteRegister(PWMCONF, 0xC10D0024); err != nil {
		return err
	}

	return nil
}

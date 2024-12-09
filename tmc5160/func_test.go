//go:build test

package tmc5160

import (
	"log"
	"testing"
)

func TestCurrentVelocityToVMAX(t *testing.T) {
	stepper := NewDefaultStepper()
	log.Printf("Current Velocity = %f", stepper.VelocitySPS)
	// Expected output based on the formula
	expectedVMAX := 1398

	result := stepper.CurrentVelocityToVMAX()

	if result != uint32(expectedVMAX) {
		t.Errorf("CurrentVelocityToVMAX() = %d; expected %v", result, expectedVMAX)
	}
}

func TestDesiredVelocityToVMAX(t *testing.T) {
	stepper := NewDefaultStepper()

	// Test with a desired velocity of 1500 steps per second
	desiredVelocity := float32(51200.0)
	expectedVMAX := 71583

	result := stepper.DesiredVelocityToVMAX(desiredVelocity)

	if result != uint32(expectedVMAX) {
		t.Errorf("DesiredVelocityToVMAX() = %d; expected %d", result, expectedVMAX)
	}
}

func TestDesiredAccelToAMAX(t *testing.T) {
	stepper := NewDefaultStepper()

	// Test desired velocity 1000 steps per second, and desired acceleration 1.5 sec from 0 to VMAX
	dVel := float32(51200.0)
	dAcc := float32(1.50)
	expectedAMAX := 521
	result := stepper.DesiredAccelToAMAX(dAcc, dVel)
	if result != uint32(expectedAMAX) {
		t.Errorf("DesiredAccelToAMAX() = %d; expected %d", result, expectedAMAX)
	}
}

func TestVMAXToTSTEP(t *testing.T) {
	stepper := NewDefaultStepper()
	// Test with a threshold speed of 1000 Hz
	thrsSpeed := uint32(5145)
	expectedTSTEP := uint32(204)

	result := stepper.VMAXToTSTEP(thrsSpeed)

	if result != (expectedTSTEP) {
		t.Errorf("VMAXToTSTEP() = %d; expected %d", result, expectedTSTEP)
	}
}

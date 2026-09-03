package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartSpinner_DisabledByNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	started := StartSpinner("working")
	assert.False(t, started)

	// Safe when spinner never started
	StopSpinner()
	StopSpinnerWithError()
	UpdateSpinnerMessage("ignored")
}

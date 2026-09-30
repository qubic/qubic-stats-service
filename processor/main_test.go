package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ParseEpochFromFileName(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		expected uint32
	}{
		{name: "plain", fileName: "spectrum.119", expected: 119},
		{name: "relative path", fileName: "./latest.118", expected: 118},
		{name: "nested path", fileName: "/var/qubic/spectrum.231", expected: 231},
		{name: "path with dots", fileName: "../data/spectrum.v2.150", expected: 150},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			epoch, err := parseEpochFromFileName(test.fileName)
			require.NoError(t, err)
			assert.Equal(t, test.expected, epoch)
		})
	}
}

func Test_ParseEpochFromFileName_givenNoEpoch_thenError(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
	}{
		{name: "no extension", fileName: "spectrum"},
		{name: "empty extension", fileName: "spectrum."},
		{name: "not a number", fileName: "spectrum.latest"},
		{name: "negative", fileName: "spectrum.-1"},
		{name: "beyond uint32", fileName: "spectrum.4294967296"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseEpochFromFileName(test.fileName)
			assert.Error(t, err)
		})
	}
}

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConvert is an entrypoint for debugging `pulumi convert“. To use this with an editor such as
// VS Code, drop a Pulumi.yaml in the convert_testdata folder and with the VS Code Go extension, the
// code lens (grayed out text above TestConvert) should display an option to "debug test".
//
// This is ideal for debugging panics in the convert command, as the debugger will break on the
// panic.
//
// See: https://github.com/golang/vscode-go/wiki/debugging
//
// Your mileage may vary with other tooling.
func TestConvert(t *testing.T) {
	t.Parallel()

	if info, err := os.Stat("convert_testdata/Pulumi.yaml"); err != nil && os.IsNotExist(err) {
		t.Skip("skipping test, no Pulumi.yaml found")
	} else if err != nil {
		t.Fatalf("failed to stat Pulumi.yaml: %v", err)
	} else if info.IsDir() {
		t.Fatalf("Pulumi.yaml is a directory, not a file")
	}

	result := runConvert("convert_testdata", "yaml", "go", "convert_testdata/go", true)
	require.Nil(t, result, "convert failed: %v", result)
}

//nolint:paralleltest // sets env var, must be run in isolation
func TestPclConvert(t *testing.T) {
	t.Setenv("PULUMI_DEV", "TRUE")

	// Check that we can run convert from PCL to PCL
	tmp, err := os.MkdirTemp("", "pulumi-convert-test")
	assert.NoError(t, err)

	result := runConvert("pcl_convert_testdata", "pcl", "pcl", tmp, true)
	assert.Nil(t, result)

	// Check that we made one file
	pclBytes, err := os.ReadFile(filepath.Join(tmp, "main.pp"))
	assert.NoError(t, err)
	// On Windows, we need to replace \r\n with \n to match the expected string below
	pclCode := string(pclBytes)
	if runtime.GOOS == "windows" {
		pclCode = strings.Replace(pclCode, "\r\n", "\n", -1)
	}
	expectedPclCode := `key = readFile("key.pub")

output result {
    __logicalName = "result"
    value = key
}`
	assert.Equal(t, expectedPclCode, pclCode)
}

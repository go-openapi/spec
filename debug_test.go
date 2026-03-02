// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"os"
	"sync"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

var logMutex = &sync.Mutex{} //nolint:gochecknoglobals // test fixture

func TestDebug(t *testing.T) {
	// usetesting linter disabled until https://github.com/golang/go/issues/71544 is fixed for windows
	tmpFile, _ := os.CreateTemp("", "debug-test") //nolint:usetesting
	tmpName := tmpFile.Name()
	defer func() {
		Debug = false
		// mutex for -race
		logMutex.Unlock()
		_ = os.Remove(tmpName)
	}()

	// mutex for -race
	logMutex.Lock()
	Debug = true
	debugOptions()
	defer func() {
		specLogger.SetOutput(os.Stdout)
	}()

	specLogger.SetOutput(tmpFile)

	debugLog("A debug")
	Debug = false
	_ = tmpFile.Close()

	flushed, _ := os.Open(tmpName) //nolint:gosec // test file, path is from os.CreateTemp
	buf := make([]byte, 500)
	_, _ = flushed.Read(buf)
	specLogger.SetOutput(os.Stdout)
	assert.StringContainsT(t, string(buf), "A debug")
}

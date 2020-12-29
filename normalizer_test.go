package spec

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// tests that paths are normalized correctly
func TestNormalizePaths(t *testing.T) {
	type testNormalizePathsTestCases []struct {
		refPath   string
		base      string
		expOutput string
	}

	testCases := func() testNormalizePathsTestCases {
		testCases := testNormalizePathsTestCases{
			{
				// http basePath, absolute refPath
				refPath:   "http://www.anotherexample.com/another/base/path/swagger.json#/definitions/Pet",
				base:      "http://www.example.com/base/path/swagger.json",
				expOutput: "http://www.anotherexample.com/another/base/path/swagger.json#/definitions/Pet",
			},
			{
				// http basePath, relative refPath
				refPath:   "another/base/path/swagger.json#/definitions/Pet",
				base:      "http://www.example.com/base/path/swagger.json",
				expOutput: "http://www.example.com/base/path/another/base/path/swagger.json#/definitions/Pet",
			},
		}
		if runtime.GOOS == "windows" {
			testCases = append(testCases, testNormalizePathsTestCases{
				{
					// file basePath, absolute refPath, no fragment
					refPath:   `C:\another\base\path.json`,
					base:      `C:\base\path.json`,
					expOutput: `c:\another\base\path.json`,
				},
				{
					// file basePath, absolute refPath
					refPath:   `C:\another\base\path.json#/definitions/Pet`,
					base:      `C:\base\path.json`,
					expOutput: `c:\another\base\path.json#/definitions/Pet`,
				},
				{
					// file basePath, relative refPath
					refPath:   `another\base\path.json#/definitions/Pet`,
					base:      `C:\base\path.json`,
					expOutput: `c:\base\another\base\path.json#/definitions/Pet`,
				},
			}...)
			return testCases
		}
		// linux case
		testCases = append(testCases, testNormalizePathsTestCases{
			{
				// file basePath, absolute refPath, no fragment
				refPath:   "/another/base/path.json",
				base:      "/base/path.json",
				expOutput: "/another/base/path.json",
			},
			{
				// file basePath, absolute refPath
				refPath:   "/another/base/path.json#/definitions/Pet",
				base:      "/base/path.json",
				expOutput: "/another/base/path.json#/definitions/Pet",
			},
			{
				// file basePath, relative refPath
				refPath:   "another/base/path.json#/definitions/Pet",
				base:      "/base/path.json",
				expOutput: "/base/another/base/path.json#/definitions/Pet",
			},
		}...)
		return testCases
	}()

	for _, tcase := range testCases {
		out := normalizePaths(tcase.refPath, tcase.base)
		assert.Equal(t, tcase.expOutput, out)
	}
}

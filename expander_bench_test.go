// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// expandBenchmarks are the documents the expansion benchmarks run on, ordered by how much cycle
// detection they exercise.
//
// fixture-957.json is the one PR #121 had to disable for slowness, so it is the case any change
// to how cycles are remembered has to answer for. bitbucket.json is the largest cyclic document
// here, 300 $ref; clickmeter.json is the largest acyclic one, and says what expansion costs when
// cycles are not the subject.
func expandBenchmarks() []struct{ name, path string } {
	return []struct{ name, path string }{
		{"petstore-acyclic", filepath.Join("testdata", "expansion", "petstore2.0.json")},
		{"circular-spec", filepath.Join("testdata", "expansion", "circularSpec.json")},
		{"shared-node-cycles", filepath.Join("testdata", "expansion", "shared-node-cycles.json")},
		{"issue-957", filepath.Join("testdata", "bugs", "957", "fixture-957.json")},
		{"bitbucket", filepath.Join("testdata", "more_circulars", "bitbucket.json")},
		{"clickmeter", filepath.Join("testdata", "expansion", "clickmeter.json")},
	}
}

// BenchmarkExpandSpec measures expansion alone: the document is unmarshalled again for every
// iteration, since expansion rewrites it, and that unmarshalling is not timed.
func BenchmarkExpandSpec(b *testing.B) {
	for _, bench := range expandBenchmarks() {
		data, err := os.ReadFile(bench.path)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(bench.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				b.StopTimer()
				doc := new(Swagger)
				if err := json.Unmarshal(data, doc); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()

				if err := ExpandSpec(doc, &ExpandOptions{RelativeBase: bench.path}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkExpandSpecSkipSchemas measures the path flatten takes for its minimal and full modes,
// where schemata are not inlined and only $ref are rebased.
func BenchmarkExpandSpecSkipSchemas(b *testing.B) {
	for _, bench := range expandBenchmarks() {
		data, err := os.ReadFile(bench.path)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(bench.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				b.StopTimer()
				doc := new(Swagger)
				if err := json.Unmarshal(data, doc); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()

				if err := ExpandSpec(doc, &ExpandOptions{RelativeBase: bench.path, SkipSchemas: true}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

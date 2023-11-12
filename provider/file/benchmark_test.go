// Copyright (c) 2023 The konf authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package file_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ktong/konf/provider/file"
)

func BenchmarkNew(b *testing.B) {
	var loader file.File
	for i := 0; i < b.N; i++ {
		loader = file.New("testdata/config.json")
	}
	b.StopTimer()

	values, err := loader.Load()
	require.NoError(b, err)
	require.Equal(b, "v", values["k"])
}

func BenchmarkLoad(b *testing.B) {
	loader := file.New("testdata/config.json")
	b.ResetTimer()

	var (
		values map[string]any
		err    error
	)
	for i := 0; i < b.N; i++ {
		values, err = loader.Load()
	}
	b.StopTimer()

	require.NoError(b, err)
	require.Equal(b, "v", values["k"])
}

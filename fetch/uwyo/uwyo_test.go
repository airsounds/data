package uwyo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "embed"
)

//go:embed testdata/uwyo.html
var webpage []byte

func TestParseBody(t *testing.T) {
	tables, err := parseBody(bytes.NewReader(webpage))
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))

	table := tables[0]
	want := 59
	wantWind := 8
	assert.Equal(t, want, len(table.Height))
	assert.Equal(t, want, len(table.Pressure))
	assert.Equal(t, want, len(table.Temp))
	assert.Equal(t, want, len(table.Dew))
	assert.Equal(t, wantWind, len(table.WindDir))
	assert.Equal(t, wantWind, len(table.WindSpeed))

	for i := 0; i < want; i++ {
		assert.NotEqual(t, 0, table.Height[i], "table.Height[%d]", i)
		assert.NotEqual(t, 0, table.Pressure[i], "table.Pressure[%d]", i)
		assert.NotEqual(t, 0, table.Temp[i], "table.Temp[%d]", i)
		assert.NotEqual(t, 0, table.Dew[i], "table.Dew[%d]", i)
		if i < wantWind {
			assert.NotEqual(t, 0, table.WindSpeed[i], "table.WindSpeed[%d]", i)
			assert.NotEqual(t, 0, table.WindDir[i], "table.WindDir[%d]", i)
		}
	}
}

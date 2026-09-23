package uwyo

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSV(t *testing.T) {
	data := `time,longitude,latitude,pressure_hPa,geopotential height_m,temperature_C,dew point temperature_C,ice point temperature_C,relative humidity_%,humidity wrt ice_%,mixing ratio_g/kg,wind direction_degree,wind speed_m/s
2026-09-22 22:38:44,34.8149,32.0070,1010.3,31,22.5,17.9,17.9,75,75,12.84,109,1.6
2026-09-22 22:39:10,34.8134,32.0079,995.4,161,24.8,18.5,18.5,68,68,13.54,166,2.1
2026-09-22 22:39:40,34.8134,32.0085,970.4,384,22.9,18.9,18.9,78,78,14.33,232,1.5
`
	requested := time.Date(2026, time.September, 23, 0, 0, 0, 0, timezone)
	table, err := parseCSV(strings.NewReader(data), 40179, requested)
	require.NoError(t, err)
	assert.Equal(t, requested, table.Time)
	assert.Equal(t, 40179, table.Station)
	assert.Equal(t, []int{1010, 995, 970}, table.Pressure)
	assert.Equal(t, []int{101, 528, 1259}, table.Height)
	assert.Equal(t, []float32{22.5, 24.8, 22.9}, table.Temp)
	assert.Equal(t, []float32{17.9, 18.5, 18.9}, table.Dew)
	assert.Equal(t, []int{109, 166, 232}, table.WindDir)
	assert.Equal(t, []int{3, 4, 2}, table.WindSpeed)
}

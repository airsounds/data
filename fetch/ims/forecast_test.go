package ims

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "embed"
)

//go:embed testdata/forecast.xml
var data []byte

func TestForecast(t *testing.T) {
	t.Parallel()

	fs, err := predict(bytes.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, 10, len(fs))
	require.Equal(t, 79, len(fs[0].Forecast))

	// Test only the first forecast and its first hour.
	got := Forecast{
		Name:      fs[0].Name,
		Lat:       fs[0].Lat,
		Long:      fs[0].Long,
		Elevation: fs[0].Elevation,
		Forecast: []HourlyForecast{
			{
				Temp:      fs[0].Forecast[0].Temp,
				RelHum:    fs[0].Forecast[0].RelHum,
				WindSpeed: fs[0].Forecast[0].WindSpeed,
				WindDir:   fs[0].Forecast[0].WindDir,
			},
		},
	}

	want := Forecast{
		Name:      "AFULA NIR HAEMEQ",
		Lat:       32.596001,
		Long:      35.2769,
		Elevation: 59,
		Forecast: []HourlyForecast{
			{
				Temp:      12,
				RelHum:    85,
				WindSpeed: 1.3,
				WindDir:   294,
			},
		},
	}

	assert.Equal(t, got, want)

	gotTime := fs[0].Forecast[0].Time.Time
	assert.Equal(t, 2022, gotTime.Year())
	assert.Equal(t, time.February, gotTime.Month())
	assert.Equal(t, 20, gotTime.Day())
	assert.Equal(t, 15, gotTime.Hour()) // 18 IDT converted to UTC.
	assert.Equal(t, 0, gotTime.Minute())
	assert.Equal(t, 0, gotTime.Second())
	assert.Equal(t, 0, gotTime.Nanosecond())
	assert.Equal(t, time.UTC, gotTime.Location())
}

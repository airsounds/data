package noaa

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"
)

// NOAA is the sounding format consumed by the Airsounds web application.
// Heights are feet, temperatures are degrees Celsius, and wind speeds are knots.
type NOAA struct {
	Time      time.Time
	Pressure  []int
	Height    []int
	Temp      []int
	Dew       []int
	WindDir   []int
	WindSpeed []int
}

const forecastURL = "https://api.open-meteo.com/v1/forecast"

var pressureLevels = []int{1000, 925, 850, 700, 600, 500, 400, 300, 250, 200}

// GetDate returns all forecast soundings for the UTC day containing date.
func GetDate(date time.Time, lat, long float32) ([]*NOAA, error) {
	start := date.UTC().Truncate(24 * time.Hour)
	return Get(start, start.Add(24*time.Hour), lat, long)
}

// Get fetches pressure-level GFS data from Open-Meteo. NOAA's former RUC
// sounding endpoint was retired; this API exposes the same forecast fields
// globally as JSON.
func Get(start, end time.Time, lat, long float32) ([]*NOAA, error) {
	if !end.After(start) {
		return nil, fmt.Errorf("end must be after start")
	}
	query := url.Values{
		"latitude":        {fmt.Sprint(lat)},
		"longitude":       {fmt.Sprint(long)},
		"start_date":      {start.UTC().Format("2006-01-02")},
		"end_date":        {end.Add(-time.Nanosecond).UTC().Format("2006-01-02")},
		"timezone":        {"UTC"},
		"wind_speed_unit": {"kn"},
		"models":          {"gfs_seamless"},
	}
	for _, level := range pressureLevels {
		for _, variable := range []string{"temperature", "dew_point", "wind_speed", "wind_direction", "geopotential_height"} {
			query.Add("hourly", fmt.Sprintf("%s_%dhPa", variable, level))
		}
	}

	resp, err := http.Get(forecastURL + "?" + query.Encode())
	if err != nil {
		return nil, fmt.Errorf("fetching forecast: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var data forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding forecast: %w", err)
	}
	return data.soundings()
}

type forecastResponse struct {
	Hourly map[string]json.RawMessage `json:"hourly"`
}

func (r forecastResponse) values(name string, want int) ([]float64, error) {
	raw, ok := r.Hourly[name]
	if !ok {
		return nil, fmt.Errorf("missing %s in forecast response", name)
	}
	var values []float64
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", name, err)
	}
	if len(values) != want {
		return nil, fmt.Errorf("expected %d %s values, got %d", want, name, len(values))
	}
	return values, nil
}

func (r forecastResponse) soundings() ([]*NOAA, error) {
	var timestamps []string
	rawTimes, ok := r.Hourly["time"]
	if !ok {
		return nil, fmt.Errorf("missing times in forecast response")
	}
	if err := json.Unmarshal(rawTimes, &timestamps); err != nil {
		return nil, fmt.Errorf("decoding forecast times: %w", err)
	}

	type levelData struct{ temp, dew, wind, direction, height []float64 }
	levels := make([]levelData, len(pressureLevels))
	for i, pressure := range pressureLevels {
		var err error
		for _, field := range []struct {
			name string
			target *[]float64
		}{
			{fmt.Sprintf("temperature_%dhPa", pressure), &levels[i].temp},
			{fmt.Sprintf("dew_point_%dhPa", pressure), &levels[i].dew},
			{fmt.Sprintf("wind_speed_%dhPa", pressure), &levels[i].wind},
			{fmt.Sprintf("wind_direction_%dhPa", pressure), &levels[i].direction},
			{fmt.Sprintf("geopotential_height_%dhPa", pressure), &levels[i].height},
		} {
			*field.target, err = r.values(field.name, len(timestamps))
			if err != nil {
				return nil, err
			}
		}
	}

	result := make([]*NOAA, 0, len(timestamps))
	for i, timestamp := range timestamps {
		t, err := time.Parse("2006-01-02T15:04", timestamp)
		if err != nil {
			return nil, fmt.Errorf("parsing forecast time %q: %w", timestamp, err)
		}
		n := &NOAA{Time: t}
		for j, pressure := range pressureLevels {
			level := levels[j]
			n.Pressure = append(n.Pressure, pressure)
			n.Height = append(n.Height, rounded(level.height[i]*3.28084))
			n.Temp = append(n.Temp, rounded(level.temp[i]))
			n.Dew = append(n.Dew, rounded(level.dew[i]))
			n.WindSpeed = append(n.WindSpeed, rounded(level.wind[i]))
			n.WindDir = append(n.WindDir, rounded(level.direction[i]))
		}
		result = append(result, n)
	}
	return result, nil
}

func rounded(v float64) int { return int(math.Round(v)) }

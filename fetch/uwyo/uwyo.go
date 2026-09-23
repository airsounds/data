package uwyo

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const soundingURL = "https://weather.uwyo.edu/wsgi/sounding"

var timezone, _ = time.LoadLocation("Asia/Jerusalem")

// UWYO contains an observed radiosonde profile in the format consumed by the
// Airsounds web application. Heights are feet and wind speeds are knots.
type UWYO struct {
	Time      time.Time
	Station   int
	Pressure  []int
	Height    []int
	Temp      []float32
	Dew       []float32
	WindDir   []int
	WindSpeed []int
}

// Fetch retrieves the most recent scheduled midnight or noon sounding. UWYO
// retired its legacy HTML endpoint in 2024; the replacement WSGI endpoint
// returns the profile as CSV.
func Fetch(station int, t time.Time) ([]*UWYO, error) {
	hour := 0
	if t.In(timezone).Hour() > 12 {
		hour = 12
	}
	date := t.In(timezone)
	requested := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, timezone)

	q := url.Values{}
	q.Set("datetime", requested.Format("2006-01-02 15:04:05"))
	q.Set("id", strconv.Itoa(station))
	q.Set("type", "TEXT:CSV")
	requestURL := soundingURL + "?" + q.Encode()

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("fetching UWYO sounding: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("UWYO returned status %d", resp.StatusCode)
	}

	table, err := parseCSV(resp.Body, station, requested)
	if err != nil {
		return nil, err
	}
	return []*UWYO{table}, nil
}

// parseCSV converts the documented WSGI CSV output. Retaining roughly one
// sample per 100 feet keeps the published daily JSON compact while preserving
// substantially more detail than the old fixed-width endpoint.
func parseCSV(r io.Reader, station int, requested time.Time) (*UWYO, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading UWYO header: %w", err)
	}
	if len(header) < 13 || header[0] != "time" || header[3] != "pressure_hPa" {
		return nil, fmt.Errorf("unexpected UWYO CSV header")
	}

	table := &UWYO{Time: requested, Station: station}
	lastHeight := -1000
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading UWYO CSV: %w", err)
		}
		if len(record) < 13 {
			continue
		}
		pressure, ok := parseFloat(record[3])
		if !ok {
			continue
		}
		heightM, ok := parseFloat(record[4])
		if !ok {
			continue
		}
		temp, tempOK := parseFloat(record[5])
		dew, dewOK := parseFloat(record[6])
		if !tempOK || !dewOK {
			continue
		}
		heightFt := int(heightM * 3.28084)
		if heightFt < lastHeight+100 {
			continue
		}
		windDir, _ := parseFloat(record[11])
		windMS, _ := parseFloat(record[12])

		table.Pressure = append(table.Pressure, int(pressure))
		table.Height = append(table.Height, heightFt)
		table.Temp = append(table.Temp, float32(temp))
		table.Dew = append(table.Dew, float32(dew))
		table.WindDir = append(table.WindDir, int(windDir))
		table.WindSpeed = append(table.WindSpeed, int(windMS*1.94384))
		lastHeight = heightFt
	}
	if len(table.Pressure) < 2 {
		return nil, fmt.Errorf("UWYO returned no usable sounding data")
	}
	return table, nil
}

func parseFloat(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "-9999" || value == "-9999.0" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil
}

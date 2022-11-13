package ims

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/html/charset"
)

const forecastPath = "https://ims.gov.il/sites/default/files/ims_data/xml_files/IMS_001.xml"

type ForecastTime struct {
	time.Time
}

func (c *ForecastTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	const format = "2/1/2006 15:04 MST"
	var v string
	d.DecodeElement(&v, &start)
	// The IMS timestamps are given in the format "2/1/2006 15:04 MST". The timezone used is
	// always set to "UTC" where it is actually should be in IDT.
	v = v[:len(v)-3] + "IDT"
	parse, err := time.Parse(format, v)
	if err != nil {
		return err
	}
	parse = parse.UTC()
	*c = ForecastTime{parse}
	return nil
}

type Forecast struct {
	Name      string           `xml:"LocationMetaData>LocationName"`
	Lat       float32          `xml:"LocationMetaData>LocationLatitude"`
	Long      float32          `xml:"LocationMetaData>LocationLongitude"`
	Elevation float32          `xml:"LocationMetaData>LocationHeight"`
	Forecast  []HourlyForecast `xml:"LocationData>Forecast"`
}

type HourlyForecast struct {
	Time      ForecastTime `xml:"ForecastTime"`
	Temp      float32      `xml:"Temperature"`
	RelHum    float32      `xml:"RelativeHumidity"`
	WindSpeed float32      `xml:"WindSpeed"`
	WindDir   float32      `xml:"WindDirection"`
}

type forecastResponse struct {
	XMLName   xml.Name   `xml:"HourlyLocationsForecast"`
	Forecasts []Forecast `xml:"Location"`
}

func Predict() ([]Forecast, error) {
	resp, err := http.Get(forecastPath)
	if err != nil {
		return nil, fmt.Errorf("fetching forecast: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return predict(resp.Body)
}

func predict(r io.Reader) ([]Forecast, error) {
	var data forecastResponse
	d := xml.NewDecoder(r)
	d.CharsetReader = charset.NewReaderLabel
	err := d.Decode(&data)
	return data.Forecasts, err
}

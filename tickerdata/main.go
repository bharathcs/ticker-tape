package tickerdata

import (
	"encoding/csv"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"strconv"
)

type TickerData struct {
	Name        string
	Ticker      string
	Data        []opts.LineData
	XAxisSeries []string
	Period      string
	Points      int
}

func ReadData(name, ticker, period string, points int, csvData io.Reader) (*TickerData, error) {
	r := csv.NewReader(csvData)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse the csv data: %w", err)
	}

	data := make([]opts.LineData, 0)
	xAxisLabels := make([]string, 0)
	for _, row := range rows[1:] {
		xAxisLabels = append(xAxisLabels, row[0])
		high, _ := strconv.ParseFloat(row[2], 10)
		low, _ := strconv.ParseFloat(row[3], 10)
		data = append(data, opts.LineData{
			Value:  (high + low) / 2.0,
			Symbol: ticker,
		})
	}

	return &TickerData{
		Name:        name,
		Ticker:      ticker,
		Data:        data[0:points],
		Period:      period,
		Points:      points,
		XAxisSeries: xAxisLabels,
	}, nil
}

func (t TickerData) CreateLineChart(writer io.Writer) {
	line := charts.NewLine()
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: fmt.Sprintf("%s (%s) - %s (%d)", t.Name, t.Ticker, t.Period, t.Points),
	}))

	line.SetXAxis(t.XAxisSeries)
	line.AddSeries(t.Name, t.Data)
	line.Render(writer)
}

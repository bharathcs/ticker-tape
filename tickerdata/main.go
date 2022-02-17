package tickerdata

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
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
		xAxisLabels = append([]string{row[0]}, xAxisLabels...)
		high, _ := strconv.ParseFloat(row[2], 10)
		low, _ := strconv.ParseFloat(row[3], 10)
		data = append([]opts.LineData{{
			Value:  math.Round((high+low)/2.0*100) / 100, // avg of high + low, rounded to 2dp
			Symbol: ticker,
		}}, data...)
	}

	return &TickerData{
		Name:        name,
		Ticker:      ticker,
		Data:        data[len(data)-points:],
		Period:      period,
		Points:      points,
		XAxisSeries: xAxisLabels[len(xAxisLabels)-points:],
	}, nil
}

func (t TickerData) CreateLineChart(writer io.Writer) error {

	minVal, maxVal := math.Inf(1), math.Inf(-1)
	for _, val := range t.Data {
		value, _ := strconv.ParseFloat(fmt.Sprint(val.Value), 32)
		minVal = math.Min(minVal, value)
		maxVal = math.Max(maxVal, value)
	}

	buffer := 0.1 * (maxVal - minVal)
	minVal, maxVal = minVal-buffer, maxVal+buffer
	minVal, maxVal = math.Floor(minVal*10)/10, math.Ceil(maxVal*10)/10

	line := charts.NewLine()
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: fmt.Sprintf("%s (%s) - %s (%d)", t.Name, t.Ticker, t.Period, t.Points),
	}), charts.WithYAxisOpts(opts.YAxis{
		Name: "USD",
		Type: "value",
		Show: true,
		Min:  minVal,
		Max:  maxVal,
	}))

	line.SetXAxis(t.XAxisSeries).
		AddSeries(t.Name, t.Data).
		SetGlobalOptions(charts.WithTooltipOpts(opts.Tooltip{
			Show:      true,
			Trigger:   "axis",
			TriggerOn: "mousemove|click",
		})).
		SetSeriesOptions(charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 0.2}))
	return line.Render(writer)
}

package alphavantage_api

import (
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"net/http"
)

const BaseUrl = "https://www.alphavantage.co/query?"
const BaseUrlWithFormatStrings = "https://www.alphavantage.co/query?function=%s&symbol=%s&apikey=%s&datatype=%s"

type TickerQueryConfig struct {
	ApiKey   string `url:"apikey"`
	Ticker   string `url:"symbol"`
	Function string `url:"function"`
	Output   string `url:"outputsize,omitempty""`
	DataType string `url:"datatype"`
}

type apiQueryUrl = string

func (q TickerQueryConfig) SaveCsvData(writer io.Writer) error {
	responseBody, err := get(q.getUrlHacky())
	if q.DataType != "csv" {
		panic("Not coded yet :V")
	}

	data, err := ioutil.ReadAll(responseBody)
	if err != nil {
		return fmt.Errorf("failed to read data %w", err)
	}

	_, err = writer.Write(data)
	return err
}

func (q TickerQueryConfig) getUrl() apiQueryUrl {
	values, _ := query.Values(q)
	return BaseUrl + values.Encode()
}

// Required as Alpha Vantage API does NOT follow standards and order of query params is strict
func (q TickerQueryConfig) getUrlHacky() apiQueryUrl {
	return fmt.Sprintf(BaseUrlWithFormatStrings, q.Function, q.Ticker, q.ApiKey, q.DataType)
}

func get(url apiQueryUrl) (io.Reader, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get response from url: %w", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("returned non-200 status code '%d' and status '%s'", response.StatusCode, response.Status)
	}

	return response.Body, nil
}

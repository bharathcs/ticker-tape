package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"ticker-tape/alphavantage-api"
)

type RawTickerConfig struct {
	Ticker string `json:"ticker"`
	Tabs   []struct {
		Period string `json:"period"`
		Points int    `json:"points"`
	} `json:"tabs"`
}

type TickerConfig struct {
	QueryConfig alphavantage_api.TickerQueryConfig
	Period      string
	Points      int
}

type RawConfig = map[string]RawTickerConfig
type Config = map[string][]TickerConfig

func ReadConfig(configData []byte, apikey string) (Config, error) {
	if len(apikey) == 0 {
		return nil, fmt.Errorf("api key is not set")
	}

	var rawConfig map[string]RawTickerConfig
	err := json.Unmarshal(configData, &rawConfig)
	if err != nil {
		return nil, err
	}

	config := make(Config)
	for tickerFriendlyName, tickerRawConfig := range rawConfig {
		newTickerConfigs, err := parseRawTickerConfig(tickerRawConfig, apikey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s': %w", tickerFriendlyName, err)
		}

		config[tickerFriendlyName] = newTickerConfigs
	}

	return config, nil
}

func parseRawTickerConfig(tickerRawConfig RawTickerConfig, apikey string) ([]TickerConfig, error) {
	if len(tickerRawConfig.Tabs) == 0 {
		return nil, fmt.Errorf("zero tabs in config")
	}

	var result []TickerConfig

	for i, tab := range tickerRawConfig.Tabs {
		var function string
		switch strings.ToLower(tab.Period) {
		case "daily":
			function = "TIME_SERIES_DAILY"
		case "weekly":
			function = "TIME_SERIES_WEEKLY"
		case "monthly":
			function = "TIME_SERIES_MONTHLY"
		default:
			return nil, fmt.Errorf(
				"invalid period '%s' in config (index %d)", tab.Period, i)
		}

		var output string
		switch {
		case tab.Points <= 3:
			return nil, fmt.Errorf(
				"invalid num of points '%d' in config of (index %d)", tab.Points, i)
		case strings.ToLower(tab.Period) != "daily":
			output = ""
		case tab.Points <= 100:
			output = "compact"
		default:
			output = "full"
		}
		result = append(result, TickerConfig{
			QueryConfig: alphavantage_api.TickerQueryConfig{
				ApiKey:   apikey,
				Ticker:   tickerRawConfig.Ticker,
				Function: function,
				Output:   output,
				DataType: "csv",
			},
			Period: strings.Title(tab.Period),
			Points: tab.Points,
		})

	}
	return result, nil
}

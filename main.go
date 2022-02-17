package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"ticker-tape/config"
	"ticker-tape/tickerdata"
)

func main() {
	configFilePathFlag := flag.String("config", "config.json", "Path to json configuration file")
	apikeyFlag := flag.String("apikey", "-----", "Private AlphaVantage API key")
	flag.Parse()

	userConfig := parseConfig(readConfig(*configFilePathFlag), *apikeyFlag)
	saveToLocal(userConfig)
	renderCharts(userConfig)
}

func readConfig(filepath string) []byte {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0755)
	failIfError(fmt.Sprintf("Failed to open config file in path '%s'", filepath), err)
	bytes, err := ioutil.ReadAll(f)
	failIfError(fmt.Sprintf("Failed to read config file in path '%s'", filepath), err)
	return bytes
}

func parseConfig(input []byte, apikey string) config.Config {
	userConfig, err := config.ReadConfig(input, apikey)
	failIfError("Failed to parse the config:", err)
	return userConfig
}

func saveToLocal(userConfig config.Config) {
	err := os.MkdirAll("data", os.ModePerm)
	failIfError("The folder 'data' could not be created:", err)

	var apiCallsWg sync.WaitGroup

	for name, tickerConfigs := range userConfig {
		for i, tickerConfig := range tickerConfigs {
			filepath := "data/" + getFilePath(name, i, ".csv")
			f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
			failIfError(fmt.Sprintf("Failed to create file '%s':", filepath), err)

			apiCallsWg.Add(1)

			ticker := tickerConfig // intermediate variable for go routine (avoid loop problems)
			go func() {
				defer apiCallsWg.Done()
				_ = f.Truncate(0)
				err := ticker.QueryConfig.SaveCsvData(f)
				if err != nil {
					failIfError("Failed to execute query and save results", err)
				}
				_ = f.Close()
			}()
		}
	}

	apiCallsWg.Wait()
}

func renderCharts(userConfig config.Config) {
	var renderCallsWg sync.WaitGroup

	err := os.MkdirAll("out", os.ModePerm)
	failIfError("The folder 'out' could not be created:", err)

	for name, tickerConfigs := range userConfig {
		for i, tickerConfig := range tickerConfigs {
			renderCallsWg.Add(1)

			src, err := os.Open("data/" + getFilePath(name, i, ".csv"))
			failIfError("Csv data could not be read:", err)

			filepath := "out/" + getFilePath(name, i, ".html")
			f, err := os.Create(filepath)
			failIfError("Html file could not be created:", err)

			ticker := tickerConfig // intermediate variable for go routine (avoid loop problems)
			go func() {
				defer renderCallsWg.Done()
				t, err := tickerdata.ReadData(name, ticker.QueryConfig.Ticker, ticker.Period, ticker.Points, src)
				failIfError("Ticker data could not be created:", err)

				_ = f.Truncate(0)
				err = t.CreateLineChart(f)
				failIfError("Chart could not be rendered in html", err)

				_ = f.Close()
				fmt.Println("Saved to ", filepath)
			}()
		}
	}

	renderCallsWg.Wait()

	var htmlLinks []struct {
		Link string
		Text string
	}

	for name, x := range userConfig {
		for i, tab := range x {
			filepath := "./" + getFilePath(name, i, ".html")
			htmlLinks = append(htmlLinks, struct {
				Link string
				Text string
			}{Link: filepath, Text: fmt.Sprintf("%s (%s)", name, tab.Period)})
		}
	}

	t := template.Must(template.New("mainpage").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Centerpiece</title>
</head>
<body>
{{range .}}
	<a href="{{.Link}}">{{.Text}}</a></br>
{{end}}
</body>
</html>
`))
	filepath := "out/" + "index.html"
	mainFile, err := os.Create(filepath)
	_ = mainFile.Truncate(0)
	failIfError("Html file could not be created:", err)

	err = t.Execute(mainFile, htmlLinks)
	if err != nil {
		failIfError("Html file could not be created:", err)
	}
	fmt.Println("Saved to ", filepath)
}

func getFilePath(name string, index int, extension string) string {
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, "\\", "-", -1)
	return strings.Replace(name, ".", "-", -1) + "-" + fmt.Sprint(index) + extension
}

func failIfError(message string, err error) {
	if err != nil {
		log.Fatalln(message, err)
	}
}

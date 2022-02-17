package main

import (
	"flag"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/opts"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"ticker-tape/config"
	"ticker-tape/tickerdata"
)

// generate random data for bar chart
func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(300)})
	}
	return items
}

func main() {
	configFilePathFlag := flag.String("config", "config.json", "Path to json configuration file")
	apikeyFlag := flag.String("apikey", "-----", "Private AlphaVantage API key")
	flag.Parse()

	userConfig := readConfig(*configFilePathFlag)
	saveToLocal(parseConfig(userConfig, *apikeyFlag))
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
			go func() {
				defer apiCallsWg.Done()
				f.Truncate(0)
				tickerConfig.QueryConfig.SaveCsvData(f)
				f.Close()
			}()
		}
	}

	apiCallsWg.Wait()

	var renderCallsWg sync.WaitGroup

	err = os.MkdirAll("out", os.ModePerm)
	failIfError("The folder 'out' could not be created:", err)

	for name, tickerConfigs := range userConfig {
		for i, tickerConfig := range tickerConfigs {
			renderCallsWg.Add(1)

			src, err := os.Open("data/" + getFilePath(name, i, ".csv"))
			failIfError("Csv data could not be created:", err)

			filepath := "out/" + getFilePath(name, i, ".html")
			f, err := os.Create(filepath)
			failIfError("Html file could not be created:", err)

			go func() {
				defer renderCallsWg.Done()
				t, err := tickerdata.ReadData(name, tickerConfig.QueryConfig.Ticker, tickerConfig.Period, tickerConfig.Points, src)
				failIfError("Ticker data could not be created:", err)

				f.Truncate(0)
				t.CreateLineChart(f)
				f.Close()
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
	failIfError("Html file could not be created:", err)
	mainFile.Truncate(0)
	t.Execute(mainFile, htmlLinks)
	fmt.Println("Savedto ", filepath)
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

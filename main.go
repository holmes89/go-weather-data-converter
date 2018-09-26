package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

const BASE_URL = "https://www.metaweather.com/api/location/%s/"
func main() {

	cities := []string{"44418", "2358820", "2471217", "2459115", "4118", "2372071", "615702", "968019", "727232", "650272"}

	fmt.Printf("Gathering weather information for %d cities\n", len(cities))
	fmt.Printf("\n#######################\n\n")
	data := convertWeatherData(getWeatherData(cities...))
	for data := range data{

		if data.Error != nil {
			fmt.Printf("Error fetching weather data for city id: %s\n", data.ID)
			continue
		}

		fmt.Printf("Weather Forcast for %s, %s\n", data.City, data.Country.Name)

		for _, day := range data.Forecast {
			fmt.Printf("\tDate: %s\n", day.Date)
			fmt.Printf("\t\t%s, High of %.2f℉, Low of %.2f℉\n\n", day.Type, day.MaxTemp, day.MinTemp)
		}
		fmt.Printf("\n#######################\n\n")
	}

	fmt.Println("Data fetch complete!")
}

func getWeatherData(ids ...string) <-chan WeatherData {

	var wg sync.WaitGroup

	out := make(chan WeatherData)
	wg.Add(len(ids))

	for _, id := range ids {
		go func(cityId string){
			var data WeatherData
			data.ID = cityId

			url := fmt.Sprintf(BASE_URL, cityId)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				data.Error = errors.New("Error making request")
				out <- data
				wg.Done()
				return
			}

			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				data.Error = errors.New("Error calling client")
				out <- data
				wg.Done()
				return
			}

			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				data.Error = errors.New("Error decoding message")
				out <- data
				wg.Done()
				return
			}
			out <- data
			wg.Done()
		}(id)
	}


	go func(){
		wg.Wait()
		close(out)
	}()

	return out
}

func convertWeatherData(weatherData <-chan WeatherData)  <-chan WeatherData {
	out := make(chan WeatherData)

	go func(){
		for data := range weatherData{

			//This could be broken out even more
			for _, day := range data.Forecast {
				day.MaxTemp = conversionCtoF(day.MaxTemp)
				day.MinTemp = conversionCtoF(day.MinTemp)
			}
			out <- data
		}
		close(out)
	}()

	return out
}

func conversionCtoF(temp float64) float64{
	return temp * 1.8 + 32
}

type WeatherData struct {
	ID string
	Country CountryData `json:"parent"`
	City string `json:"title"`
	Forecast []*WeatherDay `json:"consolidated_weather"`
	Error error
}

type WeatherDay struct {
	Type string `json:"weather_state_name"`
	Date string `json:"applicable_date"`
	MinTemp float64 `json:"min_temp"`
	MaxTemp float64 `json:"max_temp"`
}

type CountryData struct {
	Name string `json:"title"`
}

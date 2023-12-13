package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/wcharczuk/go-chart"
)


type CalculationRequest struct {
	Num1   float64 `json:"num1"`
	Num2   float64 `json:"num2"`
	System *string `json:"system"`
}

type CalculationResponse struct {
	Result string `json:"result"`
}

type RequestInfo struct {
	Num1   float64
	Num2   float64
	System string
}

var requests []RequestInfo

func add(num1, num2 float64) float64 {
	return num1 + num2
}

func subtract(num1, num2 float64) float64 {
	return num1 - num2
}

func multiply(num1, num2 float64) float64 {
	return num1 * num2
}

func divide(num1, num2 float64) (float64, error) {
	if num2 == 0 {
		return 0, errors.New("division by zero")
	}
	return num1 / num2, nil
}

func modulus(num1, num2 float64) (float64, error) {
	if num2 == 0 {
		return 0, errors.New("division by zero")
	}
	return math.Mod(num1, num2), nil
}

func convertToDecimal(num string, system string) (float64, error) {
	var base int
	switch system {
	case "binary":
		base = 2
	case "octal":
		base = 8
	case "hexadecimal":
		base = 16
	default:
		return 0, errors.New("invalid system")
	}

	decimal, err := strconv.ParseInt(num, base, 64)
	if err != nil {
		return 0, err
	}

	return float64(decimal), nil
}

func convertFromDecimal(num float64, system string) string {
	var base int
	switch system {
	case "binary":
		base = 2
	case "octal":
		base = 8
	case "hexadecimal":
		base = 16
	}

	var result string
	for num > 0 {
		remainder := math.Mod(num, float64(base))
		num = math.Floor(num / float64(base))
		switch int(remainder) {
		case 0:
			result = "0" + result
		case 1:
			result = "1" + result
		case 2:
			result = "2" + result
		case 3:
			result = "3" + result
		case 4:
			result = "4" + result
		case 5:
			result = "5" + result
		case 6:
			result = "6" + result
		case 7:
			result = "7" + result
		case 8:
			result = "8" + result
		case 9:
			result = "9" + result
		case 10:
			result = "A" + result
		case 11:
			result = "B" + result
		case 12:
			result = "C" + result
		case 13:
			result = "D" + result
		case 14:
			result = "E" + result
		case 15:
			result = "F" + result
		}
	}

	if len(result) == 0 {
		return "0"
	}

	return result
}

func calculate(w http.ResponseWriter, r *http.Request) {
	var req CalculationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var system string
	if req.System == nil {
		system = "decimal"
	} else {
		system = *req.System
	}

	var res CalculationResponse
	switch system {
	case "decimal":
		switch r.URL.Path {
		case "/add":
			res.Result = strconv.FormatFloat(add(req.Num1, req.Num2), 'f', -1, 64)
		case "/subtract":
			res.Result = strconv.FormatFloat(subtract(req.Num1, req.Num2), 'f', -1, 64)
		case "/multiply":
			res.Result = strconv.FormatFloat(multiply(req.Num1, req.Num2), 'f', -1, 64)
		case "/divide":
			result, err := divide(req.Num1, req.Num2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			res.Result = strconv.FormatFloat(result, 'f', -1, 64)
		case "/modulus": // Добавлен новый случай для подсчета остатка от деления
			result, err := modulus(req.Num1, req.Num2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			res.Result = strconv.FormatFloat(result, 'f', -1, 64)	
		default:
			http.NotFound(w, r)
			return
		}
	case "binary", "octal", "hexadecimal":
		num1, err := convertToDecimal(strconv.FormatFloat(req.Num1, 'f', -1, 64), system)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		num2, err := convertToDecimal(strconv.FormatFloat(req.Num2, 'f', -1, 64), system)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var result float64
		switch r.URL.Path {
		case "/add":
			result = add(num1, num2)
		case "/subtract":
			result = subtract(num1, num2)
		case "/multiply":
			result = multiply(num1, num2)
		case "/divide":
			result, err = divide(num1, num2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case "/modulus": // Добавлен новый случай для подсчета остатка от деления
			result, err = modulus(num1, num2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		default:
			http.NotFound(w, r)
			return
		}
		res.Result = convertFromDecimal(result, system)
	default:
		http.Error(w, "invalid system", http.StatusBadRequest)
		return
	}

	requests = append(requests, RequestInfo{Num1: req.Num1, Num2: req.Num2, System: system})

	json.NewEncoder(w).Encode(res)
}

func drawChart(w http.ResponseWriter, r *http.Request) {
	var data []chart.Value

	freq := make(map[string]int)
	for _, req := range requests {
		freq[req.System]++
	}

	for system, count := range freq {
		data = append(data, chart.Value{Label: system, Value: float64(count)})
	}

	pieChart := chart.PieChart{
		Title:  "Frequency of Number Systems",
		Values: data,
	}

	w.Header().Set("Content-Type", "image/png")
	err := pieChart.Render(chart.PNG, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//Если хочется нарисовать столбчатую диаграмму
/*func drawChart(w http.ResponseWriter, r *http.Request) {
	var data []chart.Value

	freq := make(map[string]int)
	for _, req := range requests {
	  freq[req.System]++
	}

	for system, count := range freq {
	  data = append(data, chart.Value{Label: system, Value: float64(count)})
	}

	barChart := chart.BarChart{
	  Title: "Frequency of Number Systems",
	  Bars: data,
	  BarWidth: 50,
	  XAxis: chart.Style{
		Show: true,
	  },
	  YAxis: chart.YAxis{
		Name: "Count",
		Style: chart.Style{
		  Show: true,
		},
	  },
	}

	w.Header().Set("Content-Type", "image/png")
	err := barChart.Render(chart.PNG, w)
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  }*/

func main() {
	http.HandleFunc("/add", calculate)
	http.HandleFunc("/subtract", calculate)
	http.HandleFunc("/multiply", calculate)
	http.HandleFunc("/divide", calculate)
	
	http.HandleFunc("/modulus", calculate)
	http.HandleFunc("/chart", drawChart)

	fmt.Println("Listening on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

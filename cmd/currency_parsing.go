package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"reflect"
)

var list CurrencyList

func UpdateCurrencyJson() {
	log.Printf("UpdateCurrencyJson invoke")
	resp, err := http.Get("https://www.cbr-xml-daily.ru/daily_json.js")
	if err != nil {
		log.Print(err)
	}

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		log.Print(err)
	}
}

func ConvertFromRubTo(currency string, sum float32) (float64, error) {
	fmt.Printf("ConvertFromRubTo invoke, currency = %s, sum = %f", currency, sum)
	nominal, ok1 := list.Valute[currency]["Nominal"]
	value, ok2 := list.Valute[currency]["Value"]
	if !ok1 || !ok2 {
		return 0, echo.NewHTTPError(412, "wrong currency param")
	}

	v, err := getFloat(value)
	if err != nil {
		return 0, err
	}
	n, err := getFloat(nominal)
	if err != nil {
		return 0, err
	}
	fmt.Printf("nominal = %f, value = %f", nominal, value)

	return float64(sum) / v * n, nil
}

func getFloat(unk interface{}) (float64, error) {
	floatType := reflect.TypeOf(float64(0))
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, echo.NewHTTPError(412, "wrong currency param")
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const ratesUrl = "https://www.cbr-xml-daily.ru/latest.js"
const ratesCacheKey = "rates"

var myClient = &http.Client{Timeout: 10 * time.Second}

func (app *Application) getCbrRate(currency string) (float64, error) {
	cache := ""
	if app.Redis != nil {
		cache = app.Redis.GetObj(ratesCacheKey)
	}
	if cache == "" {
		r, err := myClient.Get(ratesUrl)
		if err != nil {
			return 0, err
		}
		defer r.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		cache = buf.String()
		if app.Redis != nil {
			app.Redis.SetObj(ratesCacheKey, cache)
		}
	}
	var result map[string]interface{}
	err := json.Unmarshal([]byte(cache), &result)
	if err != nil {
		return 0, err
	}
	if rates, ok := result["rates"]; ok {
		if rate, ok := rates.(map[string]interface{})[currency]; ok {
			return rate.(float64), nil
		}
	}
	return 0, errors.New("unable to get rates")
}

package utils

import (
	"encoding/json"
	"fmt"
	"g-fresh/internal/model"
	"math"
	"net/http"
)

func CalculateDistance(userPincode string) uint {
	api := "6b57bfc6a50448ce9ba966b0a86c3532"
	pincode := userPincode
	pincodeAdmin := "682304"

	lat, lng, err := getCoordinatesFromPincode(pincode, api)
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}
	lata, lnga, err := getCoordinatesFromPincode(pincodeAdmin, api)
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}
	const R = 6371
	dLat := (lata - lat) * (math.Pi / 180.0)
	dLon := (lnga - lng) * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat*(math.Pi/180.0))*math.Cos(lata*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c
	fmt.Println(distance, pincode)
	switch {
	case distance <= 10:
		return 0
	case distance <= 20:
		return 50
	case distance > 20:
		return 100
	default:
		return 0
	}
}
func getCoordinatesFromPincode(pincode string, apiKey string) (float64, float64, error) {
	url := fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s", pincode, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var geoResp model.GeoCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return 0, 0, err
	}

	if len(geoResp.Results) > 0 {
		lat := geoResp.Results[0].Geometry.Lat
		lng := geoResp.Results[0].Geometry.Lng
		return lat, lng, nil
	}

	return 0, 0, fmt.Errorf("no results found for pincode %s", pincode)
}

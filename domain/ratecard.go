package domain

import (
	"time"
)

// Meter - Contains the detail for each rate
type Meter struct {
	EffectiveDate    time.Time          `json:"EffecitveDate"`
	IncludedQuantity float64            `json:"IncludedQuanity"`
	MeterCategory    string             `json:"MeterCategory"`
	MeterID          string             `json:"MeterId"`
	MeterName        string             `json:"MeterName"`
	MeterRates       map[string]float64 `json:"MeterRates"`
	MeterRegion      string             `json:"MeterRegion"`
	MeterStatus      string             `json:"MeterStatsu"`
	MeterSubCategory string             `json:"MeterSubCategory"`
	UnitOfMeasure    string             `json:"UnitOfMeasure"`
	Unit             string             `json:"Unit"`
}

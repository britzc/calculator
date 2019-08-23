package domain

import "time"

type Point struct {
	SubscriptionID   string
	Subscription     string
	MeterID          string
	MeterCategory    string
	MeterSubCategory string
	ResourceGroup    string
	Resource         string
	BillPeriod       string
	Quantity         float64
	Cost             float64
	Tags             map[string]string
	Timestamp        time.Time
}

package domain

import "time"

// InstanceData - Contains all the additional data of the entry
type InstanceData struct {
	Resources struct {
		ResourceURI string                 `json:"resourceUri"`
		Location    string                 `json:"location"`
		Tags        map[string]interface{} `json:"tags"`
		// Tags        struct {
		// 	Organization string `json:"Organization"`
		// 	ServerRole   string `json:"ServerRole"`
		// 	Function     string `json:"Function"`
		// 	Environment  string `json:"Environment"`
		// 	DisplayName  string `json:"DisplayName"`
		// } `json:"tags"`
	} `json:"Microsoft.Resources"`
}

// Properties - Contains details for each usage record
type Properties struct {
	SubscriptionID   string        `json:"subscriptionId"`
	UsageStartTime   time.Time     `json:"usageStartTime"`
	UsageEndTime     time.Time     `json:"usageEndTime"`
	MeterName        string        `json:"meterName"`
	MeterRegion      string        `json:"meterRegion"`
	MeterCategory    string        `json:"meterCategory"`
	MeterSubCategory string        `json:"meterSubCategory"`
	MeterRate        float64       `json:"meterRate"`
	Unit             string        `json:"unit"`
	InstanceDataText string        `json:"instanceData"`
	InstanceData     *InstanceData `json:"-"`
	ResourceGroup    string        `json:"-"`
	Resource         string        `json:"-"`
	MeterID          string        `json:"meterId"`
	Quantity         float64       `json:"quantity"`
}

// UsageRecords - Contains the individual records data
type UsageRecord struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
}

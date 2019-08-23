package domain

// Config - Application utilisation parameters
type Config struct {
	TenantID          string            `json:"tenantId"`
	Subscription      string            `json:"subscription"`
	SubscriptionID    string            `json:"subscriptionId"`
	ClientID          string            `json:"clientId"`
	ClientSecret      string            `json:"clientSecret"`
	OfferDurableID    string            `json:"offerDurableId"`
	Currency          string            `json:"currency"`
	Locale            string            `json:"locale"`
	RegionInfo        string            `json:"regionInfo"`
	TimeOffset        int               `json:"timeOffset"`
	InfluxHost        string            `json:"influxHost"`
	InfluxDB          string            `json:"influxDB"`
	InfluxMeasurement string            `json:"influxMeasurement"`
	RateMultiply      float64           `json:"rateMultiply"`
	TagDefaults       map[string]string `json:"tagDefaults"`
	MissingDefault    string            `json:"missingDefault"`
}

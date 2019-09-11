package aggregate

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/corneilebritz/cloudcostcalculator/domain"
)

func AggregateData(records []*domain.UsageRecord, config *domain.Config) (data map[string]*domain.Point) {
	data = make(map[string]*domain.Point)

	for _, record := range records {
		resourceGroup := config.MissingDefault
		resource := config.MissingDefault
		if record.Properties.InstanceData != nil {
			parts := strings.Split(record.Properties.InstanceData.Resources.ResourceURI, "/")
			resourceGroup = parts[4]
			resource = parts[8]
		}

		timestamp := record.Properties.UsageStartTime.Add(time.Duration(config.TimeOffset) + time.Hour)
		key := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s/%s", record.Properties.SubscriptionID, record.Properties.MeterID, record.Properties.MeterCategory, record.Properties.MeterSubCategory, resourceGroup, resource, record.Name, timestamp.Format(time.RFC3339))
		pd, found := data[key]
		if !found {
			tags := CreateTags(record, config)

			pd = &domain.Point{
				SubscriptionID:   record.Properties.SubscriptionID,
				Subscription:     config.Subscription,
				MeterID:          record.Properties.MeterID,
				MeterCategory:    record.Properties.MeterCategory,
				MeterSubCategory: record.Properties.MeterSubCategory,
				ResourceGroup:    resourceGroup,
				Resource:         resource,
				BillPeriod:       record.Name,
				Tags:             tags,
				Quantity:         0,
				Cost:             0,
				Timestamp:        timestamp,
			}
		}

		pd.Quantity += record.Properties.Quantity
		pd.Cost += record.Properties.MeterRate * record.Properties.Quantity

		data[key] = pd
	}

	return data
}

func CreateTags(record *domain.UsageRecord, config *domain.Config) (tags map[string]string) {
	tags = map[string]string{
		"SubscriptionID":   record.Properties.SubscriptionID,
		"Subscription":     config.Subscription,
		"MeterID":          record.Properties.MeterID,
		"MeterCategory":    record.Properties.MeterCategory,
		"MeterSubCategory": record.Properties.MeterSubCategory,
		"BillPeriod":       record.Name,
		"ResourceGroup":    record.Properties.ResourceGroup,
		"Resource":         record.Properties.Resource,
	}

	for key, value := range config.TagDefaults {
		tags[fmt.Sprintf("_%s", key)] = value
	}

	if record.Properties.InstanceData != nil {
		for key, value := range record.Properties.InstanceData.Resources.Tags {
			if _, ok := tags[fmt.Sprintf("_%s", key)]; !ok {
				continue
			}

			switch v := value.(type) {
			case string:
				tags[fmt.Sprintf("_%s", key)] = v
			}
		}
	}

	return tags
}

package usagedata

import (
	"encoding/json"

	"bitbucket.org/corneilebritz/cloudcostcalculator/domain"
)

// Parse - Read the json file and parse instance data
func Parse(ua *domain.UsageAggregates, tagDefaults map[string]string) (err error) {
	for i := 0; i < len(ua.UsageRecords); i++ {
		record := ua.UsageRecords[i]

		if err := json.Unmarshal([]byte(record.Properties.InstanceDataText), &record.Properties.InstanceData); err != nil {
			continue
		}

		if record.Properties.InstanceData.Resources.Tags == nil {
			record.Properties.InstanceData.Resources.Tags = make(map[string]interface{})
		}

		for key, value := range tagDefaults {
			current, ok := record.Properties.InstanceData.Resources.Tags[key]
			if !ok {
				current = ""
			}

			switch v := current.(type) {
			case string:
				if len(v) == 0 {
					record.Properties.InstanceData.Resources.Tags[key] = value
				}
			}

		}

		ua.UsageRecords[i] = record
	}

	return nil
}

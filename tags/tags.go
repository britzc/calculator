package tags

import (
	"log"

	"bitbucket.org/corneilebritz/cloudcostcalculator/domain"
)

// ApplyDefaults - Apply default values to the record
func ApplyDefaults(records []*domain.UsageRecord, groupMap map[string]*domain.Group, tagDefaults map[string]string) (err error) {
	for _, record := range records {
		if record.Properties.InstanceData == nil {
			continue
		}

		for key, value := range tagDefaults {
			current, ok := record.Properties.InstanceData.Resources.Tags[key]
			if !ok {
				current = ""
			}

			switch v := current.(type) {
			case string:
				if len(v) > 0 {
					continue
				}
			}

			if group, groupOK := groupMap[record.Properties.ResourceGroup]; groupOK {
				if groupTag, groupTagOK := group.Tags[key]; groupTagOK {
					log.Println("set group tag")
					record.Properties.InstanceData.Resources.Tags[key] = groupTag
					continue
				}
			}

			record.Properties.InstanceData.Resources.Tags[key] = value
		}

	}

	return nil
}

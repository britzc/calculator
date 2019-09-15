package cloud

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"bitbucket.org/corneilebritz/cloudcostcalculator/domain"
)

type AzureClient struct {
	config *domain.Config
	token  *domain.Token
}

func NewAzureClient(config *domain.Config) (client *AzureClient, err error) {
	client = &AzureClient{
		config: config,
	}

	if err := client.login(); err != nil {
		return nil, err
	}

	return client, nil
}

func (z *AzureClient) GetMeters() (meterMap map[string]*domain.Meter, err error) {
	baseURL, _ := url.ParseRequestURI("https://management.azure.com")
	baseURL.Path = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Commerce/RateCard", z.config.SubscriptionID)

	params := &url.Values{}
	params.Add("api-version", "2016-08-31-preview")
	params.Add("$filter", fmt.Sprintf("OfferDurableId eq '%s' and Currency eq '%s' and Locale eq '%s' and RegionInfo eq '%s'", z.config.OfferDurableID, z.config.Currency, z.config.Locale, z.config.RegionInfo))
	baseURL.RawQuery = params.Encode()

	type jsonBody struct {
		Meters        []*domain.Meter `json:"Meters"`
		Currency      string          `json:"Currency"`
		Local         string          `json:"Local"`
		IsTaxIncluded bool            `json:"IsTaxIncluded"`
	}

	var jb jsonBody
	if err := httpGetJson(baseURL.String(), z.token.AccessToken, &jb); err != nil {
		return nil, err
	}

	meterMap = make(map[string]*domain.Meter)
	for _, meter := range jb.Meters {
		meterMap[meter.MeterID] = meter
	}

	return meterMap, nil
}

func (z *AzureClient) GetGroups() (groupMap map[string]*domain.Group, err error) {
	baseURL, _ := url.ParseRequestURI("https://management.azure.com")
	baseURL.Path = fmt.Sprintf("/subscriptions/%s/resourcegroups", z.config.SubscriptionID)

	params := &url.Values{}
	params.Add("api-version", "2019-05-10")
	baseURL.RawQuery = params.Encode()

	type jsonBody struct {
		Groups   []*domain.Group `json:"value"`
		NextLink string          `json:"nextLink"`
	}

	var jb jsonBody
	if err := httpGetJson(baseURL.String(), z.token.AccessToken, &jb); err != nil {
		return nil, err
	}

	groupMap = make(map[string]*domain.Group)
	for _, group := range jb.Groups {
		groupMap[group.Name] = group
	}

	return groupMap, nil
}

func (z *AzureClient) GetReadings(startDate, endDate time.Time) (ur []*domain.UsageRecord, err error) {
	baseURL, _ := url.ParseRequestURI("https://management.azure.com")
	baseURL.Path = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Commerce/UsageAggregates", z.config.SubscriptionID)

	params := &url.Values{}
	params.Add("api-version", "2015-06-01-preview")
	params.Add("reportedStartTime", startDate.Format(time.RFC3339))
	params.Add("reportedEndTime", endDate.Format(time.RFC3339))
	params.Add("aggregationGranulatiry", "daily")
	params.Add("showDetails", "true")
	baseURL.RawQuery = params.Encode()

	type jsonBody struct {
		UsageRecords []*domain.UsageRecord `json:"value"`
		NextLink     string                `json:"nextLink"`
	}

	jb := &jsonBody{}
	if err := httpGetJson(baseURL.String(), z.token.AccessToken, &jb); err != nil {
		return nil, err
	}

	ur = append(ur, jb.UsageRecords...)

	for len(jb.NextLink) > 0 {
		log.Println("Retrieving Next Batch of Records")

		jb = &jsonBody{}
		if err = httpGetJson(jb.NextLink, z.token.AccessToken, &jb); err != nil {
			return nil, err
		}

		ur = append(ur, jb.UsageRecords...)
	}

	if err := populateInstanceData(ur); err != nil {
		return nil, err
	}

	return ur, nil
}

func (z *AzureClient) login() (err error) {
	baseURL, _ := url.ParseRequestURI("https://login.microsoftonline.com")
	baseURL.Path = fmt.Sprintf("%s/oauth2/token", z.config.TenantID)

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", z.config.ClientID)
	values.Set("client_secret", z.config.ClientSecret)
	values.Set("resource", "https://management.azure.com/")

	var token *domain.Token
	if err := httpPostJson(baseURL.String(), values, &token); err != nil {
		return err
	}

	z.token = token

	return nil
}

func populateInstanceData(records []*domain.UsageRecord) (err error) {
	for _, record := range records {
		if err := json.Unmarshal([]byte(record.Properties.InstanceDataText), &record.Properties.InstanceData); err != nil {
			continue
		}

		if record.Properties.InstanceData != nil {
			record.Properties.InstanceData.Resources.ResourceURI = strings.TrimLeft(record.Properties.InstanceData.Resources.ResourceURI, "/")
			parts := strings.Split(record.Properties.InstanceData.Resources.ResourceURI, "/")
			record.Properties.ResourceGroup = parts[3]
			record.Properties.Resource = parts[7]
		}

		if record.Properties.InstanceData.Resources.Tags == nil {
			record.Properties.InstanceData.Resources.Tags = make(map[string]interface{})
		}

	}

	return nil
}

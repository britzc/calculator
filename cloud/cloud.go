package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/corneilebritz/cloudcostcalculator/domain"
	"bitbucket.org/corneilebritz/cloudcostcalculator/usagedata"
)

func GetToken(config *domain.Config) (token *domain.Token, err error) {
	u, _ := url.ParseRequestURI("https://login.microsoftonline.com")
	u.Path = fmt.Sprintf("%s/oauth2/token", config.TenantID)

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", config.ClientID)
	values.Set("client_secret", config.ClientSecret)
	values.Set("resource", "https://management.azure.com/")

	r, _ := http.NewRequest("POST", u.String(), strings.NewReader(values.Encode())) // URL-encoded payload
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(values.Encode())))

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return token, nil
}

func GetRateCard(token *domain.Token, config *domain.Config) (rc *domain.RateCard, err error) {
	baseURL, _ := url.ParseRequestURI("https://management.azure.com")
	baseURL.Path = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Commerce/RateCard", config.SubscriptionID)

	params := &url.Values{}
	params.Add("api-version", "2016-08-31-preview")
	params.Add("$filter", fmt.Sprintf("OfferDurableId eq '%s' and Currency eq '%s' and Locale eq '%s' and RegionInfo eq '%s'", config.OfferDurableID, config.Currency, config.Locale, config.RegionInfo))
	baseURL.RawQuery = params.Encode()

	if err := getData(baseURL.String(), token.AccessToken, &rc); err != nil {
		return nil, err
	}

	return rc, nil
}

func GetUsageRecords(token *domain.Token, config *domain.Config, startDate, endDate time.Time) (ur []*domain.UsageRecord, err error) {
	baseURL, _ := url.ParseRequestURI("https://management.azure.com")
	baseURL.Path = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Commerce/UsageAggregates", config.SubscriptionID)

	params := &url.Values{}
	params.Add("api-version", "2015-06-01-preview")
	params.Add("reportedStartTime", startDate.Format(time.RFC3339))
	params.Add("reportedEndTime", endDate.Format(time.RFC3339))
	params.Add("aggregationGranulatiry", "daily")
	params.Add("showDetails", "true")
	baseURL.RawQuery = params.Encode()

	ua := &domain.UsageAggregates{}
	if err := getData(baseURL.String(), token.AccessToken, &ua); err != nil {
		return nil, err
	}

	if err := usagedata.Parse(ua, config.TagDefaults); err != nil {
		return nil, err
	}

	ur = append(ur, ua.UsageRecords...)

	for len(ua.NextLink) > 0 {
		log.Println("Retrieving Next Batch of Records")

		ua = &domain.UsageAggregates{}
		if err = getData(ua.NextLink, token.AccessToken, &ua); err != nil {
			return nil, err
		}

		if err := usagedata.Parse(ua, config.TagDefaults); err != nil {
			return nil, err
		}

		ur = append(ur, ua.UsageRecords...)
	}

	return ur, nil
}

func getData(url, token string, v interface{}) (err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{
		Timeout: time.Duration(5 * time.Minute),
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&v)
}

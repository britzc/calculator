package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"bitbucket.org/corneilebritz/cloudcostcalculator/aggregate"
	"bitbucket.org/corneilebritz/cloudcostcalculator/cloud"
	"bitbucket.org/corneilebritz/cloudcostcalculator/csv"
	"bitbucket.org/corneilebritz/cloudcostcalculator/domain"
	"bitbucket.org/corneilebritz/cloudcostcalculator/tags"

	client "github.com/influxdata/influxdb1-client/v2"
)

var (
	configPath   = flag.String("cp", "", "configuration path")
	fromDateText = flag.String("fd", "", "from date")
	toDateText   = flag.String("td", "", "to date")
	daysBack     = flag.Int("db", 5, "days back")
)

func LoadConfig(path string) (config *domain.Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func calcDates() (fromDate, toDate time.Time) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -1**daysBack)
	endDate := now.AddDate(0, 0, 2)

	if (len(*fromDateText) != 0) && (len(*toDateText) != 0) {
		startDate, _ = time.Parse("2006-01-02", *fromDateText)
		endDate, _ = time.Parse("2006-01-02", *toDateText)
	}

	fromDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	toDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	return fromDate, toDate
}

func main() {
	flag.Parse()

	fromDate, toDate := calcDates()

	log.Printf("FromDate: %s, ToDate: %s", fromDate, toDate)

	log.Printf("Loading configuration from %s\n", *configPath)
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating Azure Client")
	azureClient, err := cloud.NewAzureClient(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Loading Groups")
	groupMap, err := azureClient.GetGroups()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connecting to InfluxDB")
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: config.InfluxHost,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Printf("Extracting Azure Costs: %s\n", config.Subscription)
	if err := ExtractData(azureClient, groupMap, config, fromDate, toDate, c); err != nil {
		log.Fatal(err)
	}
}

func ExtractData(cloudClient *cloud.AzureClient, groupMap map[string]*domain.Group, config *domain.Config, fromDate, toDate time.Time, c client.Client) (err error) {
	outPath := fmt.Sprintf("data/_%s.csv", config.Subscription)
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	outWriter := bufio.NewWriter(outFile)

	csv.WriteHeaders(outWriter)

	log.Println("Loading Meters")
	meters, err := cloudClient.GetMeters()
	if err != nil {
		return err
	}

	for fromDate.Before(toDate) {
		var usageRecords []*domain.UsageRecord

		retryCount := 1
		for retryCount >= 0 {
			log.Printf("Retrieving Readings for %s: Attempt %d\n", fromDate, retryCount)
			usageRecords, err = cloudClient.GetReadings(fromDate, fromDate.Add(24*time.Hour))
			if err != nil {
				return err
			}
			log.Printf("Readin Count: %d\n", len(usageRecords))

			if len(usageRecords) > 0 {
				retryCount = 0
			}

			retryCount -= 1
		}

		tags.ApplyDefaults(usageRecords, groupMap, config.TagDefaults)

		log.Println("Calculating Costs")
		CalculateCosts(usageRecords, config, meters)

		log.Println("Aggregating Records")
		points := aggregate.AggregateData(usageRecords, config)

		log.Println("Writing Records")
		csv.WriteLines(outWriter, points)

		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  config.InfluxDB,
			Precision: "h",
		})
		if err != nil {
			return err
		}

		log.Println("Creating Metrics")
		for _, point := range points {
			if err := CreatePoint(bp, config.InfluxMeasurement, point); err != nil {
				return err
			}
		}

		log.Println("Writing Metrics")
		if err := c.Write(bp); err != nil {
			return err
		}

		fromDate = fromDate.Add(24 * time.Hour)
	}

	return nil
}

func CalculateCosts(records []*domain.UsageRecord, config *domain.Config, meters map[string]*domain.Meter) {
	for _, record := range records {
		meter := meters[record.Properties.MeterID]
		if meter == nil {
			continue
		}

		rate := meter.MeterRates["0"]
		record.Properties.MeterRate = rate * config.RateMultiply
	}
}

func CreatePoint(batchPoint client.BatchPoints, measurement string, point *domain.Point) (err error) {
	fields := map[string]interface{}{
		"Quantity": point.Quantity,
		"Cost":     point.Cost,
	}

	pt, err := client.NewPoint(measurement, point.Tags, fields, point.Timestamp)
	if err != nil {
		return err
	}

	batchPoint.AddPoint(pt)

	return nil
}

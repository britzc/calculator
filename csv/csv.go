package csv

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/corneilebritz/cloudcostcalculator/domain"
)

func WriteHeaders(w *bufio.Writer) (err error) {
	parts := []string{
		"SubscriptionID",
		"Subscription",
		"MeterID",
		"MeterCategory",
		"MeterSubCategory",
		"ResourceGroup",
		"Resource",
		"BillPeriod",
		"Quantity",
		"Cost",
		"Month",
		"Day",
	}

	line := strings.Join(parts, ",")

	if _, err := w.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
		return err
	}

	return nil
}

func WriteLines(w *bufio.Writer, points map[string]*domain.Point) (err error) {
	for _, point := range points {
		parts := []string{
			point.SubscriptionID,
			point.Subscription,
			point.MeterID,
			point.MeterCategory,
			point.MeterSubCategory,
			point.ResourceGroup,
			point.Resource,
			point.BillPeriod,
			fmt.Sprintf("%f", point.Quantity),
			fmt.Sprintf("%f", point.Cost),
			strconv.Itoa(int(point.Timestamp.Month())),
			strconv.Itoa(point.Timestamp.Day()),
		}

		line := strings.Join(parts, ",")

		if _, err := w.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			return err
		}
	}

	w.Flush()

	return nil
}

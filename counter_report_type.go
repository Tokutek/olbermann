package olbermann

import (
	"errors"
	"reflect"
	"strings"
	"time"
)

type iterCounterReportType struct {
	value             float64
	lastReportedValue float64
}

func (t *iterCounterReportType) name() string {
	return "iter"
}

func toFloat(fval reflect.Value) float64 {
	switch fval.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(fval.Int())
	case reflect.Float32, reflect.Float64:
		return fval.Float()
	default:
		panic("wrong kind for float " + fval.Kind().String())
	}
}

func (t *iterCounterReportType) add(fval reflect.Value) {
	t.value += toFloat(fval)
}

func (t *iterCounterReportType) get(iterDuration time.Duration, cumDuration time.Duration) (res float64) {
	res = (t.value - t.lastReportedValue) / iterDuration.Seconds()
	t.lastReportedValue = t.value
	return
}

type cumulativeCounterReportType struct {
	value float64
}

func (t *cumulativeCounterReportType) name() string {
	return "cum"
}

func (t *cumulativeCounterReportType) add(fval reflect.Value) {
	t.value += toFloat(fval)
}

func (t *cumulativeCounterReportType) get(iterDuration time.Duration, cumDuration time.Duration) (res float64) {
	res = t.value / cumDuration.Seconds()
	return
}

type totalCounterReportType struct {
	value float64
}

func (t *totalCounterReportType) name() string {
	return "total"
}

func (t *totalCounterReportType) add(fval reflect.Value) {
	t.value += toFloat(fval)
}

func (t *totalCounterReportType) get(iterDuration time.Duration, cumDuration time.Duration) (res float64) {
	res = t.value
	return
}

func newCounterMetric(field reflect.StructField) (metric metricType, err error) {
	reportNames := strings.Split(field.Tag.Get("report"), ",")
	if len(reportNames) < 1 {
		err = errors.New("counter metric " + field.Name + " must define reports")
		return
	}
	reports := make([]reportType, len(reportNames))
	for j := range reportNames {
		switch reportNames[j] {
		case "iter":
			reports[j] = new(iterCounterReportType)
		case "cum":
			reports[j] = new(cumulativeCounterReportType)
		case "total":
			reports[j] = new(totalCounterReportType)
		}
	}
	metric.name = field.Name
	metric.reports = reports
	return
}

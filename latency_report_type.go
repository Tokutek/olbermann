package olbermann

import (
	"errors"
	"fmt"
	"github.com/bmizerany/perks/quantile"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type windowLatencyReportType struct {
	nameString string
	strm *quantile.Stream
	quant float64
}

func newWindowLatencyReportType(name string, quant float64) (res *windowLatencyReportType) {
	res = &windowLatencyReportType{nameString: name, strm: quantile.NewTargeted(quant), quant: quant}
	return
}

func (t *windowLatencyReportType) name() string {
	return t.nameString
}

func (t *windowLatencyReportType) add(fval reflect.Value) {
	t.strm.Insert(toFloat(fval))
}

func (t *windowLatencyReportType) get(iterDuration time.Duration, cumDuration time.Duration) (res float64) {
	res = t.strm.Query(t.quant)
	t.strm.Reset()
	return
}

func (t *windowLatencyReportType) string(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func (t *windowLatencyReportType) close() {}

type cumulativeLatencyReportType struct {
	nameString string
	strm *quantile.Stream
	quant float64
}

func newCumulativeLatencyReportType(name string, quant float64) (res *cumulativeLatencyReportType) {
	res = &cumulativeLatencyReportType{nameString: name, strm: quantile.NewTargeted(quant), quant: quant}
	return
}

func (t *cumulativeLatencyReportType) name() string {
	return t.nameString
}

func (t *cumulativeLatencyReportType) add(fval reflect.Value) {
	t.strm.Insert(toFloat(fval))
}

func (t *cumulativeLatencyReportType) get(iterDuration time.Duration, cumDuration time.Duration) (res float64) {
	res = t.strm.Query(t.quant)
	return
}

func (t *cumulativeLatencyReportType) string(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func (t *cumulativeLatencyReportType) close() {}

func newLatencyMetric(field reflect.StructField) (metric metricType, err error) {
	reportNames := strings.Split(field.Tag.Get("report"), ",")
	if len(reportNames) < 1 {
		err = errors.New("latency metric " + field.Name + " must define reports")
		return
	}
	reports := make([]reportType, len(reportNames))
	for i := range reportNames {
		var percentile float64
		if percentile, err = strconv.ParseFloat(reportNames[i][1:], 64); err != nil {
			return
		}
		switch reportNames[i][:1] {
		case "w":
			reports[i] = newWindowLatencyReportType(reportNames[i], percentile*0.01)
		case "c":
			reports[i] = newCumulativeLatencyReportType(reportNames[i], percentile*0.01)
		}
	}
	metric.name = field.Name
	metric.reports = reports
	return
}

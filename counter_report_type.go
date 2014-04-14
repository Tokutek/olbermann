package olbermann

import (
	"errors"
	"fmt"
	"github.com/VividCortex/ewma"
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

func (t *iterCounterReportType) add(fval reflect.Value) {
	t.value += toFloat(fval)
}

func (t *iterCounterReportType) get(iterDuration time.Duration, cumDuration time.Duration) (res float64) {
	res = (t.value - t.lastReportedValue) / iterDuration.Seconds()
	t.lastReportedValue = t.value
	return
}

func (t *iterCounterReportType) string(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func (t *iterCounterReportType) close() {}

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

func (t *cumulativeCounterReportType) string(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func (t *cumulativeCounterReportType) close() {}

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

func (t *totalCounterReportType) string(val float64) string {
	return fmt.Sprintf("%d", int64(val))
}

func (t *totalCounterReportType) close() {}

type ewmaCounterReportType struct {
	nameString        string
	value             float64
	avg               ewma.MovingAverage
	killer            chan<- bool
}

func newEwmaCounterReportType(decaySamples int) (res *ewmaCounterReportType) {
	kill := make(chan bool)
	res = &ewmaCounterReportType{nameString: fmt.Sprintf("ewma%d", decaySamples/60), avg: ewma.NewMovingAverage(float64(decaySamples)), killer: kill}
	go func() {
		ticks := time.Tick(time.Second)
		lastTick := time.Now()
		lastReportedValue := res.value
		for {
			select {
			case <-kill:
				close(kill)
				return
			case tick := <-ticks:
				res.avg.Add((res.value - lastReportedValue) / tick.Sub(lastTick).Seconds())
				lastReportedValue = res.value
				lastTick = tick
			}
		}
	}()
	return
}

func (t *ewmaCounterReportType) name() string {
	return t.nameString
}

func (t *ewmaCounterReportType) add(fval reflect.Value) {
	t.value += toFloat(fval)
}

func (t *ewmaCounterReportType) get(iterDuration time.Duration, cumDuration time.Duration) (res float64) {
	res = t.avg.Value()
	return
}

func (t *ewmaCounterReportType) string(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func (t *ewmaCounterReportType) close() {
	t.killer <- true
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
		case "ewma1":
			reports[j] = newEwmaCounterReportType(60)
		case "ewma5":
			reports[j] = newEwmaCounterReportType(300)
		case "ewma15":
			reports[j] = newEwmaCounterReportType(900)
		case "ewma60":
			reports[j] = newEwmaCounterReportType(3600)
		case "total":
			reports[j] = new(totalCounterReportType)
		}
	}
	metric.name = field.Name
	metric.reports = reports
	return
}

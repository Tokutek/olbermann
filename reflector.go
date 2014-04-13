package olbermann

import (
	"errors"
	"reflect"
	"time"
)

type reportType interface {
	name() string
	add(fval reflect.Value)
	get(iterDuration time.Duration, cumDuration time.Duration) float64
	string(val float64) string
	close()
}

type metricType struct {
	name    string
	reports []reportType
}

type metricSetType struct {
	metrics []metricType
}

func newMetricSetTypeOf(val interface{}) (mst *metricSetType, err error) {
	rtype := reflect.TypeOf(val)
	if rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}
	if rtype.Kind() != reflect.Struct {
		err = errors.New("invalid kind of metric " + rtype.Kind().String())
		return
	}
	return newMetricSetType(rtype)
}

func newMetricSetType(rtype reflect.Type) (mst *metricSetType, err error) {
	newMst := new(metricSetType)
	newMst.metrics = make([]metricType, rtype.NumField())
	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)
		switch field.Tag.Get("type") {
		case "counter":
			if newMst.metrics[i], err = newCounterMetric(field); err != nil {
				return
			}
		default:
			err = errors.New("metric " + field.Name + " is of unknown type " + field.Tag.Get("type"))
			return
		}
	}
	mst = newMst
	return
}

func (mst *metricSetType) update(val interface{}) (err error) {
	rval := reflect.Indirect(reflect.ValueOf(val))
	if rval.Kind() != reflect.Struct {
		err = errors.New("invalid kind of metric " + rval.Kind().String())
		return
	}
	for i := range mst.metrics {
		rt := mst.metrics[i]
		val := rval.Field(i)
		for j := range rt.reports {
			rt.reports[j].add(val)
		}
	}
	return
}

func (mst *metricSetType) close() {
	for i := range mst.metrics {
		rt := mst.metrics[i]
		for j := range rt.reports {
			rt.reports[j].close()
		}
	}
}

type reportValue struct {
	name  string
	value float64
}

type metricValue struct {
	name    string
	reports []reportValue
}

type metricSetValue struct {
	metrics []metricValue
}

func (mst *metricSetType) getValues(iterDuration time.Duration, cumDuration time.Duration) (msv *metricSetValue) {
	msv = &metricSetValue{metrics: make([]metricValue, len(mst.metrics))}
	for i := range mst.metrics {
		metric := mst.metrics[i]
		msv.metrics[i].name = metric.name
		msv.metrics[i].reports = make([]reportValue, len(metric.reports))
		for j := range metric.reports {
			report := metric.reports[j]
			msv.metrics[i].reports[j] = reportValue{name: report.name(), value: report.get(iterDuration, cumDuration)}
		}
	}
	return
}

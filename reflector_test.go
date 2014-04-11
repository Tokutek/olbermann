package olbermann

import (
	"testing"
	"time"
)

type SampleMetric struct {
	IntVal   int     `type:"counter" report:"iter,total"`
	FloatVal float64 `type:"counter" report:"cum"`
}

func TestSample(t *testing.T) {
	mst, err := newMetricSetTypeOf(SampleMetric{})
	if err != nil {
		t.Error(err)
	}
	var floatTotal float64
	for i := 0; i < 100; i++ {
		mst.update(&SampleMetric{5, float64(i) * 0.1})
		floatTotal += float64(i) * 0.1
		if i%5 == 4 {
			msv := mst.getValues(time.Millisecond, time.Duration(i)*time.Millisecond)
			if msv.metrics[0].name != "IntVal" || msv.metrics[1].name != "FloatVal" {
				t.Error("invalid metric names")
			}
			if msv.metrics[0].reports[0].name != "iter" || msv.metrics[0].reports[1].name != "total" || msv.metrics[1].reports[0].name != "cum" {
				t.Error("invalid report names")
			}
			if msv.metrics[0].reports[0].value != float64(25000) {
				t.Error("expected 25000/s for IntVal iter, got", msv.metrics[0].reports[0].value)
			}
			if msv.metrics[0].reports[1].value != float64((i+1)*5) {
				t.Error("expected", (i+1)*5, " for IntVal total, got", msv.metrics[0].reports[1].value)
			}
			if msv.metrics[1].reports[0].value != floatTotal/(time.Duration(i)*time.Millisecond).Seconds() {
				t.Error("expected", floatTotal/(time.Duration(i)*time.Millisecond).Seconds(), "for FloatVal cum, got", msv.metrics[1].reports[0].value)
			}
		}
	}
}

func BenchmarkUpdateMetricsPtr(b *testing.B) {
	mst, err := newMetricSetTypeOf(SampleMetric{})
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mst.update(&SampleMetric{5, float64(i) * 0.1})
	}
}

func BenchmarkUpdateMetricsStruct(b *testing.B) {
	mst, err := newMetricSetTypeOf(SampleMetric{})
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mst.update(SampleMetric{5, float64(i) * 0.1})
	}
}

func BenchmarkIncrements(b *testing.B) {
	var metric SampleMetric
	for i := 0; i < b.N; i++ {
		metric.IntVal += 5
		metric.FloatVal += float64(i) * 0.1
	}
}

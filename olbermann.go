// Package olbermann is a reporter.
// Give it a metrics struct with tags, and it will report those metrics for you.
//
// Define a struct with tags "type" indicating what kind of data is there, and "report" requesting that a set of quantities be reported.
//
// Example:
//
// 	type ReportableMetric struct {
// 		Transactions   int64   `type:"counter" report:"iter,cum"`
// 		Faults         int64   `type:"counter" report:"cum,total"`
// 		ProcessingTime float64 `type:"counter" report:"total"`
// 	}
//
// Then just send ReportableMetric objects or pointers down a channel, olbermann will take care of the rest.
package olbermann

import (
	"reflect"
	"sync"
	"time"
)

// A Reporter represents a central point to collect a metric stream.
//
// Must create with C a channel of structs or pointers to structs defined with tags explaining the metrics to track and how to report them.
//
// Can Start() multiple reporting goroutines off the same Reporter.
//
// Must invoke Feed() on a goroutine to pull metrics off the stream.
//
// Usage:
// 	type ReportableMetric struct {
// 		Ips int64   `type:"counter" report:"iter,cum"`
// 		Ups int64   `type:"counter" report:"iter,cum"`
// 		X   float64 `type:"counter" report:"total"`
// 	}
//
// 	{
// 		metricChannel := make(chan interface{}, 100)
// 		r := olbermann.Reporter{C: metricChannel}
// 		go r.Feed()
// 		dstatKiller, err := r.Start(ReportableMetric{}, &olbermann.BasicDstatStyler)
// 		if err != nil {
// 			return
// 		}
// 		for i := 0; i < 100; i++ {
// 			metricChannel <- &ReportableMetric{1000, 20, 0.5}
// 		}
// 		dstatKiller <- true
// 	}
type Reporter struct {
	C    <-chan interface{}
	msts []*metricSetType
	lock sync.RWMutex
}

// Feed is a long-running function that consumes input to the reporter's channel until the channel is closed.
//
// Should be done on a goroutine.
func (r *Reporter) Feed() {
	for val := range r.C {
		r.lock.RLock()
		for i := range r.msts {
			if r.msts[i] != nil {
				r.msts[i].update(val)
			}
		}
		r.lock.RUnlock()
	}
}

// A Styler describes how, when, and where to display results.
//
// Current implementations:
//
//  - CsvStyler
//  - DstatStyler
type Styler interface {
	period() time.Duration
	linesBetweenHeaders() int
	printHeader(mst *metricSetType)
	printValues(curTime time.Time, mst *metricSetType, msv *metricSetValue)
}

// Start creates a goroutine printing the Reporter's metrics according to the provided Styler, and returns a channel that can be used to kill the goroutine.
//
// Needs a sample object to initialize some state, the zero value for the metric will do.
//
// Usage:
// 	dstatKiller, err := r.Start(ReportableMetric{}, &BasicDstatStyler)
// 	if err != nil {
// 		return
// 	}
// 	...
// 	dstatKiller <- true
func (r *Reporter) Start(sample interface{}, styler Styler) (killerChannel chan<- bool, err error) {
	sampleType := reflect.TypeOf(sample)
	if sampleType.Kind() == reflect.Ptr {
		sampleType = sampleType.Elem()
	}
	mst, err := newMetricSetType(sampleType)
	if err != nil {
		return
	}
	killer := make(chan bool)
	go func() {
		r.lock.Lock()
		idx := len(r.msts)
		r.msts = append(r.msts, mst)
		r.lock.Unlock()
		defer func() {
			r.lock.Lock()
			r.msts[idx].close()
			r.msts[idx] = nil
			r.lock.Unlock()
		}()

		var linesSinceHeader int
		if styler.linesBetweenHeaders() >= 0 {
			styler.printHeader(mst)
		}

		startTime := time.Now()
		lastTime := startTime
		ticker := time.Tick(time.Second)
		for {
			select {
			case <-killer:
				close(killer)
				return
			case curTime := <-ticker:
				tDiff := curTime.Sub(lastTime)
				tTotal := curTime.Sub(startTime)
				msv := mst.getValues(tDiff, tTotal)
				if styler.linesBetweenHeaders() > 0 && linesSinceHeader > styler.linesBetweenHeaders() {
					linesSinceHeader = 0
					styler.printHeader(mst)
				}
				styler.printValues(curTime, mst, msv)
				linesSinceHeader++
			}
		}
	}()
	killerChannel = killer
	return
}

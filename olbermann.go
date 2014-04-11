package olbermann

import (
	"reflect"
	"sync"
	"time"
)

// Represents a central point to collect a metric stream.
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
// 		metricChannel := make(chan *ReportableMetric, 100)
// 		r := olbermann.Reporter{C: metricChannel}
// 		go r.Feed()
// 		dstatKiller := r.Start(&olbermann.BasicDstatStyler)
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

/*
Consumes input to the reporter's channel.  Should be done on a goroutine.
*/
func (r *Reporter) Feed() {
	for val := range r.C {
		r.lock.RLock()
		for i := range r.msts {
			r.msts[i].update(val)
		}
		r.lock.RUnlock()
	}
}

// Interface describing how, when, and where to display results.
//
// Current implementations:
//
// 	- DstatStyler
type Styler interface {
	period() time.Duration
	linesBetweenHeaders() int
	printHeader(mst *metricSetType)
	printValues(mst *metricSetType, msv *metricSetValue)
}

// Starts a goroutine printing the Reporter's metrics according to the provided Styler.
//
// Returns a channel used to kill the goroutine.
//
// Usage:
// 	dstatKiller := r.Start(&BasicDstatStyler)
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
			copy(r.msts[idx:], r.msts[idx+1:])
			r.msts[len(r.msts)-1] = nil
			r.msts = r.msts[:len(r.msts)-1]
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
				styler.printValues(mst, msv)
				linesSinceHeader++
			}
		}
	}()
	killerChannel = killer
	return
}

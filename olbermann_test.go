package olbermann

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"
)

type exampleValueSet struct {
	A int `type:"counter" report:"iter,total"`
	B int `type:"counter" report:"ewma1,cum,total"`
	//Tps int `type:"counter" report:"iter,cum" name:"tps"`
}

func gen(c chan<- interface{}) {
	for i := 0; i < 10; i++ {
		//c <- &exampleValueSet{A: 1, B: 1, Tps: 198273}
		c <- &exampleValueSet{A: 1, B: 1}
		time.Sleep(400 * time.Millisecond)
	}
}

// This is not a perfect example, because we can't rely on clocks to
// generate the right output.  So we use a log.Logger that doesn't print
// times, and we don't use the iter or cum report types on large values in
// this example.  Also, EWMA values take 10 samples to start reporting.
func Example() {
	c := make(chan interface{}, 10)
	r := &Reporter{C: c}
	go r.Feed()
	if err := r.Start(exampleValueSet{}, &DstatStyler{Period: time.Second, LinesBetweenHeaders: 0, Logger: log.New(os.Stdout, "example: ", 0)}); err != nil {
		return
	}
	defer r.Close()
	gen(c)
	close(c)
	// Output:
	// example: ----------- a ------------ ------------------ b ------------------
	// example:         iter        total |        ewma1          cum        total
	// example:         2.00            2 |         0.00         2.00            2
	// example:         1.00            4 |         0.00         2.00            4
	// example:         1.00            7 |         0.00         2.33            7
	// example:         0.50            9 |         0.00         2.25            9
}

// This output is too high precision to be an accurate test, but this is about what it would produce:
// Output:
// time,"A iter","A total","B ewma1","B cum","B total"
// "2014-04-12 00:41:06.921153316 -0400 EDT",1.999723,2.000000,1.999723,1.999723,2.000000
// "2014-04-12 00:41:07.921158351 -0400 EDT",0.999928,4.000000,1.935220,1.999856,4.000000
// "2014-04-12 00:41:08.921147586 -0400 EDT",0.999956,7.000000,1.874880,2.333230,7.000000
// "2014-04-12 00:41:09.921124252 -0400 EDT",0.499986,9.000000,1.786177,2.249938,9.000000
func ExampleCsv() {
	c := make(chan interface{}, 10)
	r := &Reporter{C: c}
	go r.Feed()
	if err := r.Start(exampleValueSet{}, &CsvStyler{Period: time.Duration(10)*time.Millisecond, Writer: bufio.NewWriter(os.Stdout)}); err != nil {
		return
	}
	defer r.Close()
	gen(c)
	close(c)
}

type latencyValueSet struct {
	Latency float64 `type:"latency" report:"w50,w90,w99,c50,c90,c99,c99.9"`
}

func genLats(c chan<- interface{}) {
	for i := 0; i < 400; i++ {
		c <- latencyValueSet{Latency: rand.NormFloat64() * 20.0 + 100.0}
		time.Sleep(20 * time.Millisecond)
	}
}

// Also too high precision, but here is an idea:
// Output:
// time,"Latency w50","Latency w90","Latency w99","Latency c50","Latency c90","Latency c99","Latency c99.9"
// "2014-04-14 01:33:44.830270033 -0400 EDT",105.594874,122.014584,137.789284,105.594874,122.014584,137.789284,137.789284
// "2014-04-14 01:33:45.830272023 -0400 EDT",98.361251,124.074632,139.718386,101.896382,124.074632,145.714382,145.714382
// "2014-04-14 01:33:46.830270576 -0400 EDT",92.861055,126.280476,147.657207,98.719318,124.079168,147.657207,150.604769
// "2014-04-14 01:33:47.830261516 -0400 EDT",101.501077,120.582739,131.460395,99.969272,124.074632,147.657207,150.604769
// "2014-04-14 01:33:48.830269214 -0400 EDT",98.410273,120.395267,135.708998,99.166213,124.074632,147.657207,154.622437
// "2014-04-14 01:33:49.830263433 -0400 EDT",101.317478,125.973768,135.040215,100.043213,124.830474,147.657207,154.622437
// "2014-04-14 01:33:50.830216128 -0400 EDT",98.955441,122.682114,135.073727,99.867549,124.481247,145.714382,154.622437
// "2014-04-14 01:33:51.830258193 -0400 EDT",100.241188,125.728247,134.042818,99.969272,124.830474,145.714382,154.622437
func ExampleLatency() {
	c := make(chan interface{}, 10)
	r := &Reporter{C: c}
	go r.Feed()
	if err := r.Start(latencyValueSet{}, &CsvStyler{Period: time.Duration(10)*time.Millisecond, Writer: bufio.NewWriter(os.Stdout)}); err != nil {
		return
	}
	defer r.Close()
	genLats(c)
	close(c)
}

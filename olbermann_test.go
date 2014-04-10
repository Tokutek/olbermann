package olbermann

import (
	"log"
	"os"
	//"testing"
	"time"
)

type exampleValueSet struct {
	A int `type:"counter" report:"iter,total" name:"as"`
	B int `type:"counter" report:"total" name:"bs"`
	//Tps int `type:"counter" report:"iter,cum" name:"tps"`
}

// This is not a perfect example, because we can't rely on clocks to
// generate the right output.  So we use a log.Logger that doesn't print
// times, and we don't use the iter or cum report types on large values in
// this example.
func Example() {
	var acc exampleValueSet
	c := make(chan ValueSet, 10)
	if err := Feed(&acc, c); err != nil {
		log.Fatal(err)
	}
	//killer := BasicStart("example ", &acc)
	killer := Start(&LogPrinter{log.New(os.Stdout, "example: ", 0)}, HumanReporter, &acc)
	for i := 0; i < 10; i++ {
		//c <- &exampleValueSet{A: 1, B: 1, Tps: 198273}
		c <- &exampleValueSet{A: 1, B: 1}
		time.Sleep(400 * time.Millisecond)
	}
	close(c)
	killer <- true
	// Output:
	// example: ------------ as ---------- ------ bs ---
	// example:         iter        total |        total
	// example:         2.00            3 |            3
	// example:         2.00            5 |            5
	// example:         3.00            8 |            8
	// example:         2.00           10 |           10
}

package broker

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"
)

func printSum(
	hdr string,
	apiTime, totalTime time.Duration,
	reqSize, count int,
) {

	fmt.Printf("%s: records in: %d\n", hdr, count)
	fmt.Printf("%s: api time (time spent on api calls): %v\n", hdr, apiTime.Round(time.Millisecond))
	fmt.Printf("%s: total time (api + async processing): %v\n", hdr, totalTime.Round(time.Millisecond))

	rpb := float64(count) / apiTime.Seconds()
	mbs := rpb * float64(reqSize) / 1e6

	fmt.Printf("%s: api capacity (api requests / second): %v req/sec %v mb/sec\n",
		hdr, math.Round(rpb*100)/100, math.Round(mbs*1e3)/1e3)

	rpb = float64(count) / totalTime.Seconds()
	mbs = rpb * float64(reqSize) / 1e6

	fmt.Printf("%s: total capacity (total requests / second): %v req/sec %v mb/sec\n",
		hdr, math.Round(rpb*100)/100, math.Round(mbs*1e3)/1e3)
}

func runCapacityTest(t *testing.T, name string, count, reqSize int) {

	fmt.Printf("%s: starting capacity test\n", name)

	start := time.Now()

	numParallel := int64(runtime.NumCPU())
	fmt.Printf("%s: running using %d threads\n", name, numParallel)

	sem := semaphore.NewWeighted(numParallel)
	ctx := context.Background()

	for i := 0; i < count; i++ {
		sem.Acquire(ctx, 1)
		go func() {
			AssertTestCase(t, newBrokerTc(name, prandString(reqSize)))
			sem.Release(1)
		}()
	}

	sem.Acquire(ctx, numParallel)

	apiTime := time.Since(start)

	bg.RestartAllBrokers()

	totalTime := time.Since(start)

	printSum(name, apiTime, totalTime, reqSize, count)
}

func verifyCassBroker(
	t *testing.T, name string, expectedCount int, broker *Cass) {

	fmt.Printf("%s: ", name)
	actCount, err := broker.GetCount()
	if err != nil {
		t.Fatal(err)
	}
	if actCount != int64(expectedCount) {
		t.Fatalf("invalid count, expected: %d, got: %d\n", expectedCount, actCount)
	}
	fmt.Println("ok")
}

func newTestCassBroker(name string) *Cass {
	ks := name + "_broker_keyspace"
	ResetCassDb(ks)
	broker := NewCassBroker(ks)
	return broker
}

func TestCass(t *testing.T) {

	var count int = 1e5
	var reqSize int = 100

	syncBrokerID := "cass_sync"
	syncBroker := newTestCassBroker(syncBrokerID)
	bg.MustAddBroker(Sync(syncBrokerID, syncBroker.GetSyncSendFunc()))

	runCapacityTest(t, syncBrokerID, count, reqSize)

	asyncBrokerID := "cass_async"
	asyncBroker := newTestCassBroker(asyncBrokerID)
	bg.MustAddBroker(Async(asyncBrokerID, 1e6, asyncBroker.GetAsyncSendFunc(), os.Stderr))
	bg.StartAllBrokers()

	runCapacityTest(t, asyncBrokerID, count, reqSize)

	bg.StopAllBrokers()

	bg.MustRmBroker(syncBrokerID)
	bg.MustRmBroker(asyncBrokerID)

	fmt.Println("test done, running verification...")

	verifyCassBroker(t, asyncBrokerID, count, asyncBroker)
	verifyCassBroker(t, syncBrokerID, count, syncBroker)
}

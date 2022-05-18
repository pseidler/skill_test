package broker

import (
	"os"
	"testing"
)

func noopAsyncSendFunc([]Payload) ([]Payload, error) {
	return nil, nil
}

func TestBrokerEnqueue(t *testing.T) {

	broker := Async("test", 10, noopAsyncSendFunc, os.Stderr)

	msgs := [][]byte{
		[]byte("foo"),
		[]byte("bar"),
		[]byte("foobar"),
	}

	for i := range msgs {
		if err := broker.Enqueue(msgs[i]); err != nil {
			t.Fatal(err)
		}
	}

	if len(broker.queue) != len(msgs) {
		t.Fatalf("exepected %d elements in queue", len(msgs))
	}

	for i := 0; i < len(msgs); i++ {
		expected := msgs[i]
		got := <-broker.queue
		if string(expected) != string(got.Bytes) {
			t.Fatalf("expected: %s != got: %s\n", expected, got.Bytes)
		}
	}

}

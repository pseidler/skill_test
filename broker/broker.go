package broker

import (
	"fmt"
	"io"
	"time"
)

type AsyncSendFunc func([]Payload) ([]Payload, error)
type SendFunc func(*Payload) error

type Payload struct {
	Timestamp time.Time
	Bytes     []byte
}

type Broker struct {

	// broker will log errors into this writer
	logger io.Writer

	// broker name
	ID string

	// broker can be either:
	// async: data wont be written on arrival - instead will be queued and written in one batch later
	// sync: data will be written immediately after arrival
	SendAsync AsyncSendFunc
	Send      SendFunc

	// fields used when SendAsync is defined

	// api handler writes to queue, and broker loop is reading from it
	queue chan Payload
	// once written into sigquit broker will try to dry remaining data
	// 		- once drained will terminate loop and write into onquit channel
	sigquit chan struct{}
	onquit  chan struct{}

	// internal broker cache used to store pending messages
	cache []Payload
}

// sync broker will process data on the spot
func Sync(id string, sender SendFunc) *Broker {
	return &Broker{
		ID:   id,
		Send: sender,
	}
}

// async broker will process data asynchronously
// queueSize is maximum size for in memory buffer (in records, not bytes)
// logger will be used to report errors occurring during async processing
func Async(
	id string,
	queueSize int,
	sender AsyncSendFunc,
	logger io.Writer,
) *Broker {
	return &Broker{
		logger:    logger,
		queue:     make(chan Payload, queueSize),
		sigquit:   make(chan struct{}),
		onquit:    make(chan struct{}),
		ID:        id,
		SendAsync: sender,
	}
}

// process payload with specified broker
// this will either enqueue it so DequeueLoop can read and process it later (aync broker)
// or will instantly process payload (sync broker)
func (b *Broker) Enqueue(payload []byte) error {
	req := Payload{
		Bytes:     payload,
		Timestamp: time.Now(),
	}

	if !b.IsAsync() {
		return b.Send(&req)
	}

	select {
	case b.queue <- req:
		return nil
	default:
		return fmt.Errorf("broker '%s' can't handle more requests", b.ID)
	}
}

// try to process everything in the broker cache
func (b *Broker) tryToDrainCache() {
	if len(b.cache) == 0 {
		return
	}
	if rem, err := b.SendAsync(b.cache); err != nil {
		b.cache = rem
		fmt.Fprintf(b.logger, "failed to send batch, err was: %s\n", err)
		return
	}
	// everything was written successfully - empty the cache
	b.cache = b.cache[0:0]
}

// keep processing data until sigquit is received
func (b *Broker) RunDequeue() {

	if b.cache == nil {
		b.cache = make([]Payload, 0, 80)
	}

	defaultInterval := time.Millisecond * 500
	deadline := time.After(defaultInterval)

	reducedDeadline := false
	quitRequested := false

	for {
		select {
		case val := <-b.queue:
			b.cache = append(b.cache, val)
			if !reducedDeadline {
				// after receiving first user request reduce wait time
				deadline = time.After(time.Millisecond * 50)
				reducedDeadline = true
			}
		case <-deadline:
			if quitRequested && len(b.cache) == 0 {
				b.onquit <- struct{}{}
				return
			}
			if len(b.cache) > 0 {
				b.tryToDrainCache()
			}
			reducedDeadline = false
			deadline = time.After(defaultInterval)
		case <-b.sigquit:
			quitRequested = true
		}
	}
}

func (b Broker) IsAsync() bool {
	return b.SendAsync != nil
}

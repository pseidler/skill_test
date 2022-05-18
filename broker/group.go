package broker

import (
	"fmt"
	"os"
)

// group is collection of brokers
type Group map[string]*Broker

// find broker with specified id
func (bg Group) GetBroker(id string) *Broker {
	return bg[id]
}

// add new broker into group, return error on duplicate
func (bg Group) AddBroker(broker *Broker) error {
	// this monstrosity is basically XoR
	if (broker.Send == nil) == (broker.SendAsync == nil) {
		return fmt.Errorf("must provide either Send or SendAsync")
	}
	if broker.ID == "" {
		return fmt.Errorf("invalid broker.ID")
	}
	if bg[broker.ID] != nil {
		return fmt.Errorf("tried to add duplicate broker with key: " + broker.ID)
	}
	if broker.SendAsync != nil {
		if broker.logger == nil {
			broker.logger = os.Stderr
		}
		if broker.sigquit == nil {
			broker.sigquit = make(chan struct{})
		}
		if broker.onquit == nil {
			broker.onquit = make(chan struct{})
		}
		if broker.queue == nil {
			broker.queue = make(chan Payload, 1000)
		}
	}
	bg[broker.ID] = broker
	return nil
}

// same as AddBroker but will panic on error
func (bg Group) MustAddBroker(broker *Broker) {
	if err := bg.AddBroker(broker); err != nil {
		panic(err)
	}
}

// remove broker with specified id from the group
func (bg Group) RmBroker(id string) error {
	if bg[id] == nil {
		return fmt.Errorf("tried to remove broker %s which doesnt exist", id)
	}
	delete(bg, id)
	return nil
}

// remove broker with specified id from the group
func (bg Group) MustRmBroker(id string) {
	if err := bg.RmBroker(id); err != nil {
		panic(err)
	}
}

// start all async brokers in connection
// note that this function _is not_ idempotent.
// before calling it again you must call StopAllBrokers
func (bg Group) StartAllBrokers() {
	for i := range bg {
		if !bg[i].IsAsync() {
			continue
		}
		go bg[i].RunDequeue()
	}
}

// terminates brokers processing but not before draining all of their cache
// in case of broker error this method will block indefinitely
func (bg Group) StopAllBrokers() {
	for i := range bg {
		broker := bg[i]
		if !broker.IsAsync() {
			continue
		}

		fmt.Println(broker.ID + ": sent sigquit")
		broker.sigquit <- struct{}{}
		<-broker.onquit
		fmt.Println(broker.ID + ": broker exited")
	}
}

// equivalent to calling StopAllBrokers and then StartAllBrokers
func (bg Group) RestartAllBrokers() {
	bg.StopAllBrokers()
	bg.StartAllBrokers()
}

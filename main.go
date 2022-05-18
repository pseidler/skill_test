package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pseidler/skill_test/broker"
)

func main() {
	engine := gin.New()

	// disable trusted proxies, since this service is not supposed to be run with any
	engine.SetTrustedProxies(nil)

	brokerGroup := make(broker.Group)

	// here you may add another brokers, note that names must be unique

	brokerGroup.MustAddBroker(broker.Sync("fs-sync", broker.GetFsSendFunc("fs-sync")))
	brokerGroup.MustAddBroker(
		broker.Async("fs-async", 1e6, broker.GetAsyncFsSendFunc("fs-async"), os.Stderr),
	)

	cassKs := "skill_test_cass_db"
	broker.ResetCassDb(cassKs)
	cassBroker := broker.NewCassBroker(cassKs)

	brokerGroup.MustAddBroker(broker.Sync("cass-sync", cassBroker.GetSyncSendFunc()))
	brokerGroup.MustAddBroker(
		broker.Async("cass-async", 1e6, cassBroker.GetAsyncSendFunc(), os.Stderr),
	)

	brokerGroup.StartAllBrokers()

	brokerGroup.RegisterBrokerHandler(engine)

	var addr string
	flag.StringVar(&addr, "addr", "0.0.0.0:1500", "address")
	flag.Parse()

	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}

//
// This api is supposed to act like gateway / middleman for various brokers
// it operates using HTTP protocol
//
// example api request:
// 	 curl -X POST -i http://127.0.0.1:1500/broker?id=fs-sync -d hello
//
// this will log 'hello' into a file using synchronous logger: 'fs-sync'
//
package broker

# API

## Get started:

`go run main.go` to run api natively <br/>
`make inline-container` to run api and cassandra in podman container <br/>
`make inline-container engine="sudo docker"` if you want to run it with default docker installation <br/>
`make test` to run basic test using filesystem broker <br/>
`make test-cass` to run more advanced stress test utilizing single cassandra node <br/>
`make doc` to see html documentation <br/>

## How does it work:

This api is supposed to act like gateway / middleman for various brokers
it operates using HTTP protocol; example request:

curl -X POST -i http://127.0.0.1:1500/broker?id=fs-async -d $'hello\n'

during planning i decided to separate policy from mechanism, thus i do not enforce payload format.
Instead it's up to specific broker implementation to validate and process payload.
In previous example module fs-async is just writing everything into a file.

## implemented brokers:

1. fs-async 
2. fs-sync
3. cass-async
4. cass-sync

'sync' broker will process user data in realtime and then return <br/>
'async' broker will enqueue user data in memory and then process it at later time <br/>

cass broker is something i did for fun, to test whether async / batch processing
with cassandra can yield better results than synchronous processing.
If you are curious, you may verify it by yourself by executing `make test-cass`

## Contributing

to expand this api with other brokers, you may create your own Go module 
and implement broker.AsyncSendFunc (for async support) and/or broker.SendFunc (for sync support)

then add broker into the main broker.Group with MustAddBroker providing your implementations.

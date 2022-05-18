test:
	go test ./broker -v -run "(Fs|Group|Broker)"

test-cass:
	go test ./broker -v -run Cass

engine=podman

inline-container:
	$(engine) container rm sk-cass -f || :
	$(engine) container rm sk-api -f || :
	$(engine) run -d --rm --name sk-cass --net host docker.io/library/cassandra:4.0.3
	$(engine) run -it --rm --name sk-api --net host	\
		 -v `pwd`:/api -w /api docker.io/golang:1.18.0 go run main.go

doc:
	godoc -http=:6543
	firefox http://127.0.0.1:6543/pkg/github.com/pseidler/skill_test/broker/

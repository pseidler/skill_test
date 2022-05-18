package broker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
)

// cass is more advanced broker which saves request into cassandra table

// represents cass connection to the single keyspace
type Cass struct {
	MaxBatchSize int
	Sess         *gocql.Session
}

func NewCassBroker(keyspace string) *Cass {
	ret := new(Cass)
	ret.Sess = CreateCassSession(keyspace)
	ret.MaxBatchSize = 10 * 1024
	return ret
}

const insertQuery = "insert into data (id, ts, content) values (?,?,?)"

// function returned from this method will be processing data in cassandra batches
// launched asynchronously using semaphore
func (cb *Cass) GetAsyncSendFunc() AsyncSendFunc {

	type payloadBatch struct {
		start, end int
		batch      *gocql.Batch
		err        error
	}

	return func(p []Payload) ([]Payload, error) {
		var n int64 = 100 // int64(runtime.NumCPU())
		sem := semaphore.NewWeighted(n)
		ctx := context.Background()
		batch := cb.Sess.NewBatch(gocql.LoggedBatch)
		start := 0
		currentBatchSize := 0

		failedBatches := make([]payloadBatch, 0, 10)
		var fblock sync.Mutex

		for i := range p {
			id := uuid.New()
			// sizeof(uuid) + sizeof(timestamp) + sizeof(bytes)
			currentBatchSize += 16 + 16 + len(p[i].Bytes)
			batch.Query(insertQuery, id[:], p[i].Timestamp, p[i].Bytes)
			if currentBatchSize >= cb.MaxBatchSize || i == len(p)-1 {
				sem.Acquire(ctx, 1)
				pb := payloadBatch{
					start: start,
					end:   i,
					batch: batch,
				}
				go func(pb payloadBatch) {
					if err := cb.Sess.ExecuteBatch(pb.batch); err != nil {
						fblock.Lock()
						defer fblock.Unlock()
						pb.err = err
						failedBatches = append(failedBatches, pb)
					}
					sem.Release(1)
				}(pb)
				batch = cb.Sess.NewBatch(gocql.LoggedBatch)
				start = i
				currentBatchSize = 0
			}
		}

		sem.Acquire(ctx, n)

		// if some of the batches failed, combine all of the failed data
		// and return it so it'll be requeued
		if len(failedBatches) != 0 {
			ret := make([]Payload, 0, 100)
			for i := range failedBatches {
				s := failedBatches[i].start
				e := failedBatches[i].end
				ret = append(ret, p[s:e]...)
			}
			return ret, failedBatches[0].err
		}

		return nil, nil
	}
}

// this is effectively simple insert
func (cb *Cass) GetSyncSendFunc() SendFunc {
	return func(p *Payload) error {
		id := uuid.New()
		return cb.Sess.Query(insertQuery, id[:], p.Timestamp, p.Bytes).Exec()
	}
}

// get count from data table - note that this will be very slow, use in testing only
func (cb *Cass) GetCount() (int64, error) {
	var c int64
	return c, cb.Sess.Query("select count(1) from data").Scan(&c)
}

// create new cassandra session from default username and password
func CreateCassSession(ks string) *gocql.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 5
	cluster.Timeout = time.Second * 5
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.Keyspace = ks

	retries := 10
	var lerr error

	// giving cassandra some time to wake up
	for {
		sess, err := gocql.NewSession(*cluster)
		if err == nil {
			return sess
		}
		fmt.Printf("failed to connect to cass node, retries left: %d\n", retries)
		time.Sleep(time.Second * 6)
		retries--
		if retries == 0 {
			lerr = err
			break
		}
	}

	log.Fatalf("failed to connect to cass node, last err was: %v", lerr)
	return nil
}

// drop existing keyspace and then recreate it
func ResetCassDb(ks string) {
	sess := CreateCassSession("")
	ddl := []string{
		`drop keyspace if exists %s`,
		`create keyspace %s with replication = {'class':'SimpleStrategy', 'replication_factor' : 1}`,
		`create table %s.data (id uuid, ts timestamp, content blob, primary key ( ( id ), ts ))`,
	}

	for i := range ddl {
		if err := sess.Query(fmt.Sprintf(ddl[i], ks)).Exec(); err != nil {
			log.Fatal(err)
		}
	}

	sess.Close()
}

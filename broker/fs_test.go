package broker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

var apiEngine *gin.Engine
var bg Group

func init() {
	apiEngine = gin.New()
	bg = make(Group)
	bg.RegisterBrokerHandler(apiEngine)
}

type TestCase struct {
	RequestMethod      string
	RequestUrl         string
	RequestReader      io.Reader
	ExpectedStatusCode int
	BodyCallback       func([]byte) error `json:"-"`
}

func JsonMustMarshalIndent(v interface{}) string {
	if b, err := json.MarshalIndent(v, "", "  "); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

func AssertTestCase(t *testing.T, tc TestCase) {
	req, err := http.NewRequest(tc.RequestMethod, tc.RequestUrl, tc.RequestReader)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	apiEngine.ServeHTTP(recorder, req)
	if recorder.Code != tc.ExpectedStatusCode {
		t.Fatalf("invalid status code for tc:\n%v\n: expected %d, got %d",
			JsonMustMarshalIndent(tc),
			tc.ExpectedStatusCode,
			recorder.Code)
	}
}

// adding some more weight into LF
var prandCharset = []rune("\n\n\n\n\n\n\n1234567890asdfghjklqwertyuiopzxc")

func prandString(size int) string {
	ret := make([]rune, size)
	for i := range ret {
		cindex := rand.Intn(len(prandCharset))
		ret[i] = prandCharset[cindex]
	}
	return string(ret)
}

var brokerMethod = "POST"

func newBrokerTc(brokerID string, req string) TestCase {
	brokerUrl := fmt.Sprintf("%s?id=%s", BrokerPath, brokerID)
	return TestCase{
		RequestMethod:      brokerMethod,
		RequestUrl:         brokerUrl,
		RequestReader:      strings.NewReader(req),
		ExpectedStatusCode: 204,
	}
}

func runTestOnFsBroker(t *testing.T, id string) {

	// ensure we have clean start
	_ = os.Remove(id)

	// cleanup
	defer os.Remove(id)

	// make {count} requests each containing {size} bytes of payload.
	// all of them must succeed

	size := 100
	count := 100
	var sb strings.Builder
	for i := 0; i < count; i++ {
		randString := prandString(size)
		sb.WriteString(randString)
		AssertTestCase(t, newBrokerTc(id, randString))
	}

	bg.RestartAllBrokers()

	// verify that file was created and its content conforms to request
	c, err := ioutil.ReadFile(id)
	if err != nil {
		t.Fatal(err)
	}

	if string(c) != sb.String() {
		t.Fatalf("%s: unexpected file content - does not conform to request", id)
	}

}

func TestFsEndpoint(t *testing.T) {

	bname := "fs-sync"

	// add mockup brokers
	bg.MustAddBroker(Sync(bname, GetFsSendFunc(bname)))
	bg.StartAllBrokers()

	// assert that invalid path won't work
	AssertTestCase(t, TestCase{
		RequestMethod:      brokerMethod,
		RequestUrl:         "/",
		ExpectedStatusCode: 404,
	})

	// assert that invalid method won't work
	AssertTestCase(t, TestCase{
		RequestMethod:      "GET",
		RequestUrl:         BrokerPath,
		ExpectedStatusCode: 404,
	})

	// assert thay broker_id is required
	AssertTestCase(t, TestCase{
		RequestMethod:      brokerMethod,
		RequestUrl:         BrokerPath,
		ExpectedStatusCode: 400,
		RequestReader:      strings.NewReader("will fail"),
	})

	syncFsBrokerUrl := fmt.Sprintf("%s?id=%s", BrokerPath, bname)

	// assert thay payload is required
	AssertTestCase(t, TestCase{
		RequestMethod:      brokerMethod,
		RequestUrl:         syncFsBrokerUrl,
		ExpectedStatusCode: 400,
	})

	bg.StopAllBrokers()
	bg.MustRmBroker(bname)
}

func TestFsAsync(t *testing.T) {
	bname := "test-fs-async"
	bg.MustAddBroker(Async(bname, 1e6, GetAsyncFsSendFunc(bname), os.Stderr))
	bg.StartAllBrokers()
	runTestOnFsBroker(t, bname)
	bg.StopAllBrokers()
	bg.MustRmBroker(bname)
}

func TestFsSync(t *testing.T) {
	bname := "test-fs-sync"
	bg.MustAddBroker(Sync(bname, GetFsSendFunc(bname)))
	bg.StartAllBrokers()
	runTestOnFsBroker(t, bname)
	bg.StopAllBrokers()
	bg.MustRmBroker(bname)
}

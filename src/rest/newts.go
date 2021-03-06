package rest

import "fmt"
import "time"

import "io/ioutil"
import "math/rand"
import "encoding/json"

import "net/http"
import "bytes"
import "os"
import "strconv"

var r *rand.Rand

const seed int64 = 99
const nrJobs = 2000
const nrWorkers = 8
const nrSitesPerWorker = 6250

type TSNewts struct {
	mSlice []Measurement

	Id        int64
	Metric    string
	IdxSeries int
	Tags      map[string]string
	Val       int64
}

type Measurement struct {
	siteID int32   // site id between 0 and 49999
	mType  int32   // measurement type between 0 and 99
	t      int64   // unix time at which instance was generated by genMeasurement
	v      float64 // values are random in [0, 100) range
	tags   map[string]string
}

// Marshaller for newts
// Comment this out and replace it with your own marshaller
type NewtsResource struct {
	ID   string            `json:"id"`
	attr map[string]string `json:"attributes"`
}

func (m *Measurement) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Timestamp int64         `json:"timestamp"`
		Resource  NewtsResource `json:"resource"`
		Name      string        `json:"name"`
		Type      string        `json:"type"`
		Value     float64       `json:"value"`
	}{
		Timestamp: m.t / int64(time.Millisecond),
		Resource:  NewtsResource{ID: "localhost:chassis:temps", attr: m.tags},
		Name:      fmt.Sprintf("Sensor %v-%v", m.siteID, m.mType),
		Type:      "GAUGE",
		Value:     m.v,
	})
}

func (newts *TSNewts) Init(id int64) {
	newts.Id = id
	r = rand.New(rand.NewSource(seed))
}

func (newts *TSNewts) Create(name string, metric string, stamp int64, value float64, tags map[string]string) {
	//newts.mSlice = append(newts.mSlice, Measurement{int32(newts.Id)*int32(nrSitesPerWorker) + r.Int31n(nrSitesPerWorker), r.Int31n(100), stamp, value, tags})
	// Single site
	newts.mSlice = append(newts.mSlice, Measurement{5, 5, stamp / int64(time.Millisecond), value, tags})
}

func (newts *TSNewts) Add(host string, port int64) {
	var url string = "http://"
	var cmd string = "samples"

	url += host
	url += ":"
	url += strconv.FormatInt(port, 10)
	url += "/"
	url += cmd

	mJson, _ := json.Marshal(newts.mSlice)
	//fmt.Println(string(mJson))

	if mJson != nil {

	}

	resp, _ := http.Post(url, "application/json", bytes.NewReader(mJson))
	if resp == nil {
		fmt.Print("No response")
	} else {
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s", err)
			fmt.Printf("%s\n", string(contents))
			os.Exit(1)
		}

		if resp.StatusCode != 201 {
			fmt.Println()
			fmt.Println("Response code: ", resp.StatusCode) //Uh-oh this means our test failed
			fmt.Println()
			os.Exit(1)
		}

		defer resp.Body.Close()
	}

	newts.Reset()
}

func (newts *TSNewts) Reset() {
	newts.mSlice = []Measurement{}
}

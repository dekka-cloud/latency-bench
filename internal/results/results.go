// Package results defines what a probe run records, so runs from different people,
// networks and days can be pooled and compared. Every duration is in milliseconds.
package results

import "time"

// Version is the format of a Run. It changes only when a field changes meaning.
const Version = 1

// Run is one probe run from one vantage point against every target it was given.
type Run struct {
	Version int       `json:"version"`
	Probe   string    `json:"probe"`   // the probe's version
	Started time.Time `json:"started"` // UTC
	Vantage Vantage   `json:"vantage"`
	Targets []Target  `json:"targets"`
}

// Vantage describes where the measurements came from. The probe never looks up the
// person's own address; they describe their network themselves.
type Vantage struct {
	City    string `json:"city"`           // such as Accra
	Country string `json:"country"`        // ISO 3166 alpha-2, such as GH
	Network string `json:"network"`        // such as MTN, Telecel, AT, or a fibre provider
	Access  string `json:"access"`         // fibre, cable, dsl, 5g, 4g, 3g, satellite or other
	Note    string `json:"note,omitempty"` // anything else worth knowing, such as "evening peak"
}

// Target is everything measured against one deployment of the target service.
type Target struct {
	Name     string `json:"name"`     // the label from the targets file, such as dekka-accra
	URL      string `json:"url"`      // the base URL that was measured
	Location string `json:"location"` // what the target says about itself, from /info
	Provider string `json:"provider"`
	Error    string `json:"error,omitempty"` // set when the target couldn't be measured at all

	Connect  []Connect  `json:"connect"`  // fresh connections, one per round
	Warm     Stats      `json:"warm"`     // round trips on one kept-alive connection
	Chatty   []Chatty   `json:"chatty"`   // sequences of dependent API calls
	Page     []Page     `json:"page"`     // a page load: HTML, assets in parallel, then API calls
	Download []Transfer `json:"download"` // throughput, smallest size first
	Upload   []Transfer `json:"upload"`
	UDP      *UDP       `json:"udp,omitempty"`   // nil when UDP was skipped or blocked
	Route    *Route     `json:"route,omitempty"` // nil when tracing wasn't possible
}

// Connect is the cost of reaching the target from nothing, as a first visit pays it.
type Connect struct {
	DNS     float64 `json:"dns"`     // resolving the name; 0 when the URL is an address
	TCP     float64 `json:"tcp"`     // the TCP handshake
	TLS     float64 `json:"tls"`     // the TLS handshake; 0 over plain HTTP
	TTFB    float64 `json:"ttfb"`    // from sending the request to the first byte of the reply
	Total   float64 `json:"total"`   // everything, until the whole small reply arrived
	Proto   string  `json:"proto"`   // HTTP/1.1 or HTTP/2.0
	Address string  `json:"address"` // the address connected to
	Error   string  `json:"error,omitempty"`
}

// Stats summarises repeated samples.
type Stats struct {
	N      int     `json:"n"`
	Min    float64 `json:"min"`
	Median float64 `json:"median"`
	P90    float64 `json:"p90"`
	P99    float64 `json:"p99"`
	Max    float64 `json:"max"`
	Mean   float64 `json:"mean"`
	Jitter float64 `json:"jitter"` // the mean absolute difference between consecutive samples
}

// Chatty is a run of API calls made one after another on one connection, as an app does
// when each call needs the last one's answer. Total grows with distance times Calls.
type Chatty struct {
	Calls int     `json:"calls"`
	Total float64 `json:"total"`
}

// Page approximates loading an app page on a new connection: the HTML, then six assets
// in parallel, then API calls one after another.
type Page struct {
	Calls int     `json:"calls"`
	Total float64 `json:"total"`
	HTML  float64 `json:"html"`
}

// Transfer is one download or upload of Bytes.
type Transfer struct {
	Bytes int64   `json:"bytes"`
	Time  float64 `json:"time"`
	Mbps  float64 `json:"mbps"`
	Error string  `json:"error,omitempty"`
}

// UDP is a stream of small packets at a steady rate, as a call or a game sends them.
type UDP struct {
	Sent     int     `json:"sent"`
	Received int     `json:"received"`
	Loss     float64 `json:"loss"` // the share of packets that never came back, from 0 to 1
	RTT      Stats   `json:"rtt"`
}

// Route is the path packets took, hop by hop, from a traceroute.
type Route struct {
	Tool string `json:"tool"` // traceroute, tracepath or tracert
	Hops []Hop  `json:"hops"`
}

// Hop is one router on the path.
type Hop struct {
	N       int     `json:"n"`
	Address string  `json:"address,omitempty"` // empty when the hop didn't answer
	Name    string  `json:"name,omitempty"`    // its reverse DNS name, which often names a city
	RTT     float64 `json:"rtt,omitempty"`
	Country string  `json:"country,omitempty"` // from a GeoIP database, when the probe was given one
}

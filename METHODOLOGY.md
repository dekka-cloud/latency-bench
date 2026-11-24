# Methodology

This page says how each measurement is taken, what it includes, and what can skew it, so results from different people can be compared fairly.

## The targets

Every target runs the same Go program and answers the same requests. Measurement endpoints send `Cache-Control: no-store`, so nothing along the way can answer for the target, and downloads are random bytes, so no compression can shrink them. A study should run every target on the same size of machine with nothing else busy on it, so that differences come from distance and the networks in between.

## First connection

The probe makes a brand new connection for each round, with its own HTTP client, and times a request for `/ping` with Go's `httptrace`:

- **DNS** runs from looking up the name to the answer. The operating system's resolver may answer from its cache, so the first round usually pays the most.
- **TCP** is the handshake, one round trip.
- **TLS** is the handshake after it: one round trip with TLS 1.3, two with TLS 1.2.
- **Time to first byte** runs from the request being sent to the first byte of the answer, one more round trip plus the target's work, which is almost nothing.
- **Total** runs until the whole small answer has arrived.

The browser page reads the same steps from Resource Timing, which browsers expose across origins because every target sends `Timing-Allow-Origin: *`. A browser can't open a fresh connection to the server that sent the page, so that one target shows as already open.

## Round trips and calls one after another

Round trips are thirty small requests, one after another on a connection that's already open, so they show the network's own delay without any setup. The probe reports their median, 90th and 99th percentiles and jitter, the mean difference between one round trip and the next.

Calls one after another are 10, 20 and 30 requests for a small JSON object, each sent only once the last has arrived, which is how most apps load data when one answer decides the next request. Their total is close to the number of calls times the round trip, and it's where distance becomes something a person notices.

## Page loads

A page load opens a new connection, fetches the HTML, then fetches six assets in parallel, then makes 10 or 20 calls one after another. Over HTTP/2 the assets share one connection; over HTTP/1.1 the client opens more. It approximates a single-page app starting up and is repeated each round.

## Throughput

Downloads and uploads use the kept-alive connection. TCP starts slowly and speeds up once data flows, and it speeds up once per round trip, so on a long path a small transfer never reaches the link's full speed. That's why 100 KB and 1 MB transfers show distance more clearly than 10 MB ones.

## UDP jitter and loss

The probe sends 100 small UDP packets 20 ms apart, as a voice call does, to the target's echo on port 7001, and times each reply. A packet not back within a second counts as lost. Some networks and firewalls block or throttle UDP, so a run with every packet lost is worth checking against a run from another network before reading it as loss.

## Route

The route comes from the machine's own `traceroute`, `tracepath` or `tracert`, one probe per hop. Hops that don't answer show without an address. Each answering hop's reverse DNS name is looked up, because routers' names often include a city or an airport code. With `--geoip`, each hop is placed in a country from the database you give; router locations in GeoIP databases are approximate, so treat a single surprising hop with care and look at the pattern across runs. Traffic between two places in one country can leave it, and the route is how you see that.

## Running a study

- Measure from every kind of network your users have, such as each mobile operator and fibre, and say which with `--network` and `--access`.
- Repeat at different times of day, including the evening peak, and on different days. Use `--note` to record anything unusual.
- Run every target in one probe run, so they share the same conditions.
- Report medians and the 90th percentile, and say how many runs each comes from. The report does this.

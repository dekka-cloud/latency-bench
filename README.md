# latency-bench

latency-bench measures what distance costs the people using an app. You run the same small service in several places, such as Accra, Frankfurt, Virginia and Johannesburg, and measure it from where your users are, on the networks they really use. The result shows how much of an app's speed is decided by where its server sits, from the first connection to a page that makes twenty calls one after another.

It has three parts. The **target** is the service you deploy in each location with Docker Compose. The **probe** is a command-line tool you run from your users' side of the network. The **report** pools many probe runs into tables you can compare. Every target also serves a **browser test page**, so anyone can measure from their own phone or laptop and share the result.

## What it measures

| Measure | What it tells you |
| --- | --- |
| First connection | DNS lookup, TCP connect, TLS handshake and time to first byte for a new visitor, each timed separately |
| Round trip | A small request on a connection that's already open, repeated thirty times, with its median and jitter |
| Calls one after another | 10, 20 and 30 API calls where each waits for the last, as most apps make them, which multiplies every round trip |
| Page load | A page on a new connection: its HTML, six assets in parallel, then 10 or 20 API calls |
| Throughput | Downloads of 100 KB, 1 MB and 10 MB and uploads of 100 KB and 1 MB, where long round trips slow TCP's ramp-up |
| UDP jitter and loss | 100 small packets 20 ms apart, as a voice call sends them, with how many came back and how steadily |
| Route | The hops between you and the target, with their names and, given a GeoIP database, the countries the traffic crosses |

[METHODOLOGY.md](METHODOLOGY.md) explains how each one is taken and what can skew it.

## Deploy a target

On each machine, with Docker and Compose installed:

```sh
git clone https://github.com/dekka-cloud/latency-bench && cd latency-bench/deploy
cp .env.example .env        # say where this machine is and who hosts it
docker compose up -d
```

The target listens on port 80 for HTTP, 443 for HTTPS and 7001 for the UDP echo, so open those in the machine's firewall. Without a domain it serves HTTPS with a self-signed certificate, which the probe measures with `--insecure` at the same cost as a real one. For the browser page you need a real certificate: point a domain at the machine, set `DOMAIN` in `.env`, move the target's ports as `.env.example` shows, and run `docker compose --profile tls up -d` to have Caddy get one from Let's Encrypt.

Use the same size of machine in every location, so the comparison is about distance and not hardware.

## Run the probe

```sh
go install github.com/dekka-cloud/latency-bench/cmd/probe@latest
probe --targets targets.json --city Accra --country GH --network MTN --access 4g
```

`targets.json` lists the targets to measure, as in [targets.example.json](targets.example.json). The probe measures each in turn, writes everything to `results-<time>.json` and prints a short summary. Describe your network with `--network` and `--access` on every run, because the same target looks very different over fibre and over 3G. To name the countries a route crosses, pass `--geoip` with an MMDB country database, such as DB-IP's free IP to Country Lite. The route is traced with the machine's own `traceroute`, `tracepath` or `tracert`.

## Compare results

```sh
go install github.com/dekka-cloud/latency-bench/cmd/report@latest
report results-*.json > report.md
report --csv results-*.json > report.csv
```

The report groups runs by where they were taken from and shows the median of every measure against every target.

## The browser test page

Every target serves a page at its root that measures each target listed in its `deploy/targets.json` from the visitor's browser. It times first connections through the browser's Resource Timing, round trips, twenty calls in a row and transfers, and sends nothing anywhere unless the visitor ticks the box to share. Shared results are kept only by a target started with `BENCH_COLLECT=/data`, in `results.jsonl` on its volume, with what the visitor entered and the timings and never their address.

## RIPE Atlas

RIPE Atlas probes hosted on networks across a country can ping and trace the route to every target without anyone running anything. [atlas/README.md](atlas/README.md) shows how to schedule those measurements.

## Studies

[studies/2026-11-accra](studies/2026-11-accra) holds every result from Dekka's 22-hour study from Accra in November 2026, published as open data under CC BY 4.0, with the script that turns them into the figures in [The round trip from Accra](https://dekka.cloud/research/round-trip-from-accra).

## Licence

latency-bench is released under the [Apache License 2.0](LICENSE), and the study results under their own licence, as each study's README says. It's built by [Dekka](https://dekka.cloud) to study what hosting near users does for them, and anyone is welcome to run it and publish what they find.

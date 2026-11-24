# RIPE Atlas

[RIPE Atlas](https://atlas.ripe.net) is a network of small measurement devices hosted by volunteers and networks around the world. Its probes in a country can ping and trace the route to every latency-bench target from many networks at once, which complements probe runs by people. Scheduling measurements spends credits, which you earn by hosting a probe or can ask RIPE NCC for; create an API key with permission to schedule measurements at [atlas.ripe.net/keys](https://atlas.ripe.net/keys/).

This schedules a ping and a traceroute from 20 probes in Ghana to one target, repeating every four hours for a week. Change `target`, `value` and the dates to suit your study, and send one request per target.

```sh
curl -sS -X POST "https://atlas.ripe.net/api/v2/measurements/" \
  -H "Authorization: Key $ATLAS_KEY" -H "Content-Type: application/json" \
  -d '{
    "definitions": [
      { "type": "ping", "af": 4, "target": "frankfurt.bench.example.com", "description": "latency-bench ping", "interval": 14400, "packets": 5 },
      { "type": "traceroute", "af": 4, "target": "frankfurt.bench.example.com", "description": "latency-bench route", "interval": 14400, "protocol": "TCP", "port": 443 }
    ],
    "probes": [ { "requested": 20, "type": "country", "value": "GH" } ],
    "start_time": "2026-12-01T00:00:00Z",
    "stop_time": "2026-12-08T00:00:00Z",
    "is_oneoff": false
  }'
```

The answer lists the new measurements' ids. Their results are at `https://atlas.ripe.net/api/v2/measurements/<id>/results/`, as JSON with each probe's network (`prb_id`, and its ASN from the probe's own record) and every round trip. Tracing over TCP to port 443 follows the path HTTPS traffic takes, which some networks route differently from ICMP.

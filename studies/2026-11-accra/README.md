# Accra, 24 and 25 November 2026

These are the raw results behind Dekka Research's report [The round trip from Accra](https://dekka.cloud/research/round-trip-from-accra). For 22 hours, from 12:48 UTC on 24 November to 11:00 UTC on 25 November, we measured the same latency-bench target running in Accra, Johannesburg, Frankfurt and Northern Virginia, from a data centre in Accra and from Globalping's volunteer machines in five West African cities.

| Path | What it holds |
| --- | --- |
| `results/probe-*.json` | 90 runs of the latency-bench probe from a Dekka machine in Accra, one every 15 minutes, each in the JSON format the probe writes |
| `results/globalping-*.jsonl` | 23 hourly rounds of Globalping measurements from Accra, Lagos, Cotonou, Ouagadougou and Freetown: 16 pings and an HTTPS request to each target every hour, and an mtr trace every third hour, one measurement per line as Globalping's API returned it |
| `targets.json` | The four targets the probe measured |
| `analyze.py` | The script that turns the results into the report's figures |
| `analysis/summary.json`, `analysis/summary.md` | What `analyze.py` produced from these results |

The targets in Accra ran on a Dekka server in our own data centre, and the three abroad on Google Cloud e2-small machines in africa-south1, europe-west3 and us-east4. The machines abroad and the `bench-*.dekka.cloud` names were deleted when the study ended, so the addresses in the results no longer lead anywhere.

To check the figures, run `python3 analyze.py` in this directory. It needs only Python's standard library and writes `analysis/` again from `results/`. The `report` command in this repository also reads the probe files.

## Personal data

There is none. The probe ran on our own machine, and Globalping describes each of its volunteer machines by city, network and the city's coordinates, never by the address of the person who runs it. No one shared a result from the browser test page during the study.

## Licence

The results, summaries and this description are released under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/), so you can use them for anything if you credit "Dekka, The round trip from Accra (2026)". `analyze.py` is under the repository's Apache License 2.0.

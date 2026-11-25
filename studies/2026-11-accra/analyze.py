"""Summarise the latency study's results.

Reads every probe-*.json and globalping-*.jsonl in results/, and writes analysis/summary.json
and analysis/summary.md: medians and 90th percentiles per target for each measure from the
Accra data centre, the same by hour of day, and Globalping's pings, HTTPS timings and routes
from probes in Ghana and nearby cities.
"""

from __future__ import annotations

import glob
import json
import os
import statistics
from collections import defaultdict

HERE = os.path.dirname(os.path.abspath(__file__))
RESULTS = os.path.join(HERE, "results")
OUT = os.path.join(HERE, "analysis")

TARGETS = {
    "bench-acc.dekka.cloud": "Accra",
    "bench-jnb.dekka.cloud": "Johannesburg",
    "bench-fra.dekka.cloud": "Frankfurt",
    "bench-iad.dekka.cloud": "Virginia",
}
ORDER = ["Accra", "Johannesburg", "Frankfurt", "Virginia"]


def short(name_or_url: str) -> str:
    for host, city in TARGETS.items():
        if host in name_or_url or city in name_or_url:
            return city
    return name_or_url


def q(values: list[float], p: float) -> float | None:
    values = sorted(v for v in values if v is not None)
    if not values:
        return None
    if len(values) == 1:
        return values[0]
    k = (len(values) - 1) * p
    lo, hi = int(k), min(int(k) + 1, len(values) - 1)
    return values[lo] + (values[hi] - values[lo]) * (k - lo)


def stats(values: list[float]) -> dict:
    values = [v for v in values if v is not None]
    return {"n": len(values), "median": q(values, 0.5), "p90": q(values, 0.9), "min": min(values) if values else None, "max": max(values) if values else None}


def probe_runs() -> list[dict]:
    runs = []
    for path in sorted(glob.glob(os.path.join(RESULTS, "probe-*.json"))):
        with open(path) as f:
            runs.append(json.load(f))
    return runs


def summarise_probe(runs: list[dict]) -> dict:
    m: dict[str, dict[str, list[float]]] = defaultdict(lambda: defaultdict(list))
    hourly: dict[str, dict[int, list[float]]] = defaultdict(lambda: defaultdict(list))
    hourly_chatty: dict[str, dict[int, list[float]]] = defaultdict(lambda: defaultdict(list))
    routes: dict[str, list[list[dict]]] = defaultdict(list)
    for run in runs:
        hour = int(run["started"][11:13])
        for t in run["targets"]:
            city = short(t["name"])
            v = m[city]
            for c in t.get("connect") or []:
                if c.get("error"):
                    v["connect_errors"].append(1)
                    continue
                for k in ("dns", "tcp", "tls", "ttfb", "total"):
                    v[f"connect_{k}"].append(c.get(k))
            if t.get("error"):
                v["target_errors"].append(1)
            w = t.get("warm") or {}
            if w.get("n") and w.get("median") is not None:
                v["rtt_median"].append(w["median"])
                v["rtt_p90"].append(w["p90"])
                v["rtt_jitter"].append(w["jitter"])
                hourly[city][hour].append(w["median"])
            for c in t.get("chatty") or []:
                v[f"chatty_{c['calls']}"].append(c["total"])
                if c["calls"] == 20:
                    hourly_chatty[city][hour].append(c["total"])
            for p in t.get("page") or []:
                v[f"page_{p['calls']}"].append(p["total"])
                v["page_html"].append(p.get("html"))
            for d in t.get("download") or []:
                if not d.get("error"):
                    v[f"down_{d['bytes']}_mbps"].append(d.get("mbps"))
                    v[f"down_{d['bytes']}_ms"].append(d.get("time"))
            for u in t.get("upload") or []:
                if not u.get("error"):
                    v[f"up_{u['bytes']}_mbps"].append(u.get("mbps"))
            u = t.get("udp") or {}
            if u.get("sent"):
                v["udp_loss"].append(100 * (u.get("loss") or 0))
                if u.get("rtt") and u["rtt"].get("n"):
                    v["udp_rtt"].append(u["rtt"].get("median"))
                    v["udp_jitter"].append(u["rtt"].get("jitter"))
            r = t.get("route") or {}
            if r.get("hops"):
                routes[city].append(r["hops"])
    out = {city: {k: stats(vals) for k, vals in sorted(v.items())} for city, v in m.items()}
    by_hour = {city: {h: statistics.median(vals) for h, vals in sorted(hs.items())} for city, hs in hourly.items()}
    by_hour_chatty = {city: {h: statistics.median(vals) for h, vals in sorted(hs.items())} for city, hs in hourly_chatty.items()}
    # The most recent route per target, with every hop that answered.
    last_routes = {city: rs[-1] for city, rs in routes.items()}
    return {"runs": len(runs), "first": runs[0]["started"] if runs else None, "last": runs[-1]["started"] if runs else None, "targets": out, "rtt_by_hour": by_hour, "chatty20_by_hour": by_hour_chatty, "routes": last_routes}


def globalping() -> dict:
    ping: dict[tuple, list] = defaultdict(list)
    loss: dict[tuple, list] = defaultdict(list)
    http: dict[tuple, dict[str, list]] = defaultdict(lambda: defaultdict(list))
    mtr: dict[tuple, list] = {}
    probes: dict[str, dict] = {}
    for path in sorted(glob.glob(os.path.join(RESULTS, "globalping-*.jsonl"))):
        with open(path) as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    meas = json.loads(line)
                except json.JSONDecodeError:
                    continue
                kind, target = meas.get("type"), short(meas.get("target", ""))
                for r in meas.get("results", []):
                    p, res = r["probe"], r["result"]
                    where = f"{p['city']}, {p['country']} ({p['network']})"
                    probes[where] = {"city": p["city"], "country": p["country"], "network": p["network"], "asn": p.get("asn")}
                    if res.get("status") != "finished":
                        continue
                    key = (where, target)
                    if kind == "ping" and res.get("stats"):
                        s = res["stats"]
                        if s.get("avg") is not None:
                            ping[key].append(s.get("min"))
                        loss[key].append(s.get("loss"))
                    elif kind == "http" and res.get("timings"):
                        for k, val in res["timings"].items():
                            http[key][k].append(val)
                    elif kind == "mtr" and res.get("hops"):
                        mtr[key] = [
                            {"host": h.get("resolvedHostname") or h.get("resolvedAddress"), "address": h.get("resolvedAddress"), "asn": h.get("asn"), "avg": (h.get("stats") or {}).get("avg")}
                            for h in res["hops"]
                        ]
    return {
        "probes": probes,
        "ping_min": {f"{w} -> {t}": stats(v) for (w, t), v in sorted(ping.items())},
        "ping_loss": {f"{w} -> {t}": stats(v) for (w, t), v in sorted(loss.items())},
        "http": {f"{w} -> {t}": {k: stats(v) for k, v in ks.items()} for (w, t), ks in sorted(http.items())},
        "mtr": {f"{w} -> {t}": hops for (w, t), hops in sorted(mtr.items())},
    }


def fmt(x: float | None, unit: str = "ms") -> str:
    if x is None:
        return "–"
    if unit == "ms":
        return f"{x:,.0f} ms" if x >= 10 else f"{x:.1f} ms"
    if unit == "mbps":
        return f"{x:,.1f} Mbps"
    if unit == "%":
        return f"{x:.0f}%"
    return f"{x}"


def markdown(p: dict, g: dict) -> str:
    rows = [
        ("Round trip on an open connection", "rtt_median", "ms"),
        ("Jitter between round trips", "rtt_jitter", "ms"),
        ("New connection, from DNS to last byte", "connect_total", "ms"),
        ("  TCP handshake", "connect_tcp", "ms"),
        ("  TLS handshake", "connect_tls", "ms"),
        ("  Time to first byte", "connect_ttfb", "ms"),
        ("10 calls one after another", "chatty_10", "ms"),
        ("20 calls one after another", "chatty_20", "ms"),
        ("30 calls one after another", "chatty_30", "ms"),
        ("Page, 6 assets and 10 calls", "page_10", "ms"),
        ("Page, 6 assets and 20 calls", "page_20", "ms"),
        ("100 KB download", "down_102400_mbps", "mbps"),
        ("1 MB download", "down_1048576_mbps", "mbps"),
        ("10 MB download", "down_10485760_mbps", "mbps"),
        ("1 MB upload", "up_1048576_mbps", "mbps"),
        ("UDP round trip", "udp_rtt", "ms"),
        ("UDP loss", "udp_loss", "%"),
    ]
    lines = [f"# Latency study summary\n\n{p['runs']} probe runs from the Accra data centre, {p['first']} to {p['last']}. Median, with the 90th percentile in brackets.\n"]
    cities = [c for c in ORDER if c in p["targets"]]
    lines.append("| Measure | " + " | ".join(cities) + " |")
    lines.append("| --- |" + " --- |" * len(cities))
    for label, key, unit in rows:
        cells = []
        for c in cities:
            s = p["targets"][c].get(key)
            cells.append("–" if not s or s["median"] is None else f"{fmt(s['median'], unit)} ({fmt(s['p90'], unit)})")
        lines.append(f"| {label} | " + " | ".join(cells) + " |")
    lines.append("\n## Round trip by hour (UTC, which is Ghana's time)\n")
    hours = sorted({h for c in cities for h in p["rtt_by_hour"].get(c, {})})
    lines.append("| Hour | " + " | ".join(cities) + " |")
    lines.append("| --- |" + " --- |" * len(cities))
    for h in hours:
        lines.append(f"| {h:02d}:00 | " + " | ".join(fmt(p["rtt_by_hour"].get(c, {}).get(h)) for c in cities) + " |")
    lines.append("\n## Globalping: lowest ping per round, median across rounds\n")
    wheres = sorted({k.split(" -> ")[0] for k in g["ping_min"]})
    lines.append("| Probe | " + " | ".join(ORDER) + " |")
    lines.append("| --- |" + " --- |" * len(ORDER))
    for w in wheres:
        cells = []
        for t in ORDER:
            s = g["ping_min"].get(f"{w} -> {t}")
            cells.append(fmt(s["median"]) if s and s["median"] is not None else "–")
        lines.append(f"| {w} | " + " | ".join(cells) + " |")
    lines.append("\n## Globalping: HTTPS request timings, median (total, TCP, TLS, first byte)\n")
    lines.append("| Probe | " + " | ".join(ORDER) + " |")
    lines.append("| --- |" + " --- |" * len(ORDER))
    for w in wheres:
        cells = []
        for t in ORDER:
            h = g["http"].get(f"{w} -> {t}")
            if not h:
                cells.append("–")
                continue
            cells.append(" / ".join(fmt((h.get(k) or {}).get("median")) for k in ("total", "tcp", "tls", "firstByte")))
        lines.append(f"| {w} | " + " | ".join(cells) + " |")
    lines.append("\n## Routes from the Accra data centre (latest)\n")
    for c in cities:
        hops = p["routes"].get(c)
        if not hops:
            continue
        lines.append(f"**{c}**: " + " → ".join(f"{h.get('name') or h.get('address') or '*'} ({h.get('rtt', 0):.0f} ms)" if h.get("address") else "*" for h in hops))
    lines.append("\n## Routes from Globalping probes (mtr, latest)\n")
    for k, hops in g["mtr"].items():
        lines.append(f"**{k}**: " + " → ".join(f"{h['host']} AS{h['asn'][0] if isinstance(h['asn'], list) and h['asn'] else h['asn'] or '?'} ({h['avg']:.0f} ms)" if h.get("avg") is not None and h.get("host") else "*" for h in hops))
    return "\n".join(lines) + "\n"


def main() -> None:
    os.makedirs(OUT, exist_ok=True)
    p = summarise_probe(probe_runs())
    g = globalping()
    with open(os.path.join(OUT, "summary.json"), "w") as f:
        json.dump({"probe": p, "globalping": g}, f, indent=1, default=str)
    with open(os.path.join(OUT, "summary.md"), "w") as f:
        f.write(markdown(p, g))
    print(markdown(p, g))


if __name__ == "__main__":
    main()

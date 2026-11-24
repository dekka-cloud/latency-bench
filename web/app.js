// Times each target from this browser. Resource Timing gives the first connection's DNS,
// connect and TLS steps (the targets send Timing-Allow-Origin so the browser exposes them),
// and repeated fetches give round trips, sequences of calls and transfer speeds.

const $ = (id) => document.getElementById(id);
const WARM_ROUNDS = 15;
const CHATTY_CALLS = 20;

let lastResult = null;

$("about").addEventListener("submit", async (event) => {
  event.preventDefault();
  $("start").disabled = true;
  try {
    lastResult = await run(Object.fromEntries(new FormData(event.target)));
    $("share").hidden = false;
  } catch (e) {
    status(`The test stopped: ${e.message}`);
  } finally {
    $("start").disabled = false;
  }
});

$("consent").addEventListener("change", (e) => ($("send").disabled = !e.target.checked));

$("copy").addEventListener("click", async () => {
  await navigator.clipboard.writeText(JSON.stringify(lastResult, null, 2));
  status("Copied the result.");
});

$("send").addEventListener("click", async () => {
  const response = await fetch("results", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(lastResult),
  });
  status(response.ok ? "Thank you, the result was shared." : "This server doesn't collect results; copy them instead.");
  $("send").disabled = true;
});

async function run(about) {
  const targets = await (await fetch("targets.json", { cache: "no-store" })).json();
  const rows = $("results").querySelector("tbody");
  rows.innerHTML = "";
  $("results").hidden = false;
  $("explain").hidden = false;
  const result = {
    version: 1,
    source: "browser",
    started: new Date().toISOString(),
    vantage: { city: about.city || "", network: about.network || "", access: about.access || "" },
    userAgent: navigator.userAgent,
    targets: [],
  };
  for (const t of targets) {
    const base = (t.url || location.origin).replace(/\/$/, "");
    status(`Measuring ${t.name}…`);
    const row = rows.insertRow();
    row.insertCell().textContent = t.name;
    try {
      const measured = await measure(base);
      result.targets.push({ name: t.name, url: base, ...measured });
      const first = measured.connect.reused
        ? "already open"
        : `${ms(measured.connect.total)} (TLS ${ms(measured.connect.tls)})`;
      for (const value of [
        first,
        ms(measured.warm.median),
        ms(measured.chatty.total),
        mbps(measured.download.mbps),
        mbps(measured.upload.mbps),
      ]) {
        const cell = row.insertCell();
        cell.className = "num";
        cell.textContent = value;
      }
    } catch (e) {
      result.targets.push({ name: t.name, url: base, error: e.message });
      const cell = row.insertCell();
      cell.colSpan = 5;
      cell.textContent = `Couldn't reach it: ${e.message}`;
    }
  }
  status("Done. Run it again on another network to compare.");
  return result;
}

async function measure(base) {
  const info = await (await fetch(`${base}/info`, { cache: "no-store" })).json();
  return {
    location: info.location,
    provider: info.provider,
    // The /info request above opened the connection, so this one is timed on a fresh one.
    connect: await firstConnection(`${base}/ping?first=${crypto.randomUUID()}`),
    warm: await warm(base),
    chatty: await chatty(base),
    download: await download(base, 1 << 20),
    upload: await upload(base, 512 << 10),
  };
}

// The first request to a new origin, read from Resource Timing. A fresh query string and
// a fresh origin per target keep the browser from answering it from a cache.
async function firstConnection(url) {
  await fetch(url, { cache: "no-store" });
  // The browser records a request's timing just after it finishes, so wait for the entry.
  let t;
  for (let i = 0; i < 20 && !t; i++) {
    [t] = performance.getEntriesByName(url);
    if (!t) await new Promise((resolve) => setTimeout(resolve, 50));
  }
  if (!t) return { total: NaN, tls: NaN, reused: false };
  return {
    dns: t.domainLookupEnd - t.domainLookupStart,
    tcp: t.connectEnd - t.connectStart - (t.secureConnectionStart > 0 ? t.connectEnd - t.secureConnectionStart : 0),
    tls: t.secureConnectionStart > 0 ? t.connectEnd - t.secureConnectionStart : 0,
    ttfb: t.responseStart - t.requestStart,
    total: t.responseEnd - t.startTime,
    reused: t.connectEnd === t.connectStart,
  };
}

async function warm(base) {
  const samples = [];
  await fetch(`${base}/ping`, { cache: "no-store" });
  for (let i = 0; i < WARM_ROUNDS; i++) {
    const started = performance.now();
    await fetch(`${base}/ping?n=${i}`, { cache: "no-store" });
    samples.push(performance.now() - started);
  }
  samples.sort((a, b) => a - b);
  return { n: samples.length, min: samples[0], median: samples[Math.floor(samples.length / 2)], max: samples.at(-1) };
}

async function chatty(base) {
  const started = performance.now();
  for (let i = 0; i < CHATTY_CALLS; i++) {
    const r = await fetch(`${base}/api/items/${i}`, { cache: "no-store" });
    await r.json();
  }
  return { calls: CHATTY_CALLS, total: performance.now() - started };
}

async function download(base, bytes) {
  const started = performance.now();
  const body = await (await fetch(`${base}/bytes?n=${bytes}`, { cache: "no-store" })).arrayBuffer();
  const time = performance.now() - started;
  return { bytes: body.byteLength, time, mbps: (body.byteLength * 8) / (time * 1000) };
}

async function upload(base, bytes) {
  const data = new Uint8Array(bytes);
  crypto.getRandomValues(data.subarray(0, Math.min(bytes, 65536)));
  const started = performance.now();
  await fetch(`${base}/upload`, { method: "POST", body: data, cache: "no-store" });
  const time = performance.now() - started;
  return { bytes, time, mbps: (bytes * 8) / (time * 1000) };
}

const ms = (v) => (Number.isFinite(v) ? `${Math.round(v)} ms` : "n/a");
const mbps = (v) => (Number.isFinite(v) ? `${v.toFixed(1)} Mbps` : "n/a");
const status = (text) => ($("status").textContent = text);

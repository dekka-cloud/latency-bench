# Latency study summary

90 probe runs from the Accra data centre, 2026-11-24T12:48:54.120779206Z to 2026-11-25T11:00:00.733196518Z. Median, with the 90th percentile in brackets.

| Measure | Accra | Johannesburg | Frankfurt | Virginia |
| --- | --- | --- | --- | --- |
| Round trip on an open connection | 1.8 ms (4.4 ms) | 169 ms (174 ms) | 109 ms (109 ms) | 171 ms (172 ms) |
| Jitter between round trips | 0.4 ms (2.0 ms) | 0.3 ms (12 ms) | 0.3 ms (1.4 ms) | 0.4 ms (1.4 ms) |
| New connection, from DNS to last byte | 121 ms (138 ms) | 616 ms (637 ms) | 431 ms (446 ms) | 618 ms (632 ms) |
|   TCP handshake | 0.8 ms (1.6 ms) | 168 ms (173 ms) | 107 ms (108 ms) | 169 ms (171 ms) |
|   TLS handshake | 19 ms (24 ms) | 178 ms (183 ms) | 117 ms (119 ms) | 179 ms (182 ms) |
|   Time to first byte | 2.0 ms (3.4 ms) | 169 ms (175 ms) | 108 ms (110 ms) | 171 ms (172 ms) |
| 10 calls one after another | 20 ms (45 ms) | 1,699 ms (1,799 ms) | 1,089 ms (1,116 ms) | 1,712 ms (1,723 ms) |
| 20 calls one after another | 39 ms (99 ms) | 3,398 ms (3,553 ms) | 2,176 ms (2,226 ms) | 3,423 ms (3,443 ms) |
| 30 calls one after another | 64 ms (155 ms) | 5,108 ms (5,313 ms) | 3,265 ms (3,317 ms) | 5,133 ms (5,160 ms) |
| Page, 6 assets and 10 calls | 165 ms (211 ms) | 3,024 ms (3,101 ms) | 1,950 ms (1,984 ms) | 3,009 ms (3,035 ms) |
| Page, 6 assets and 20 calls | 191 ms (256 ms) | 4,722 ms (4,879 ms) | 3,038 ms (3,108 ms) | 4,721 ms (4,761 ms) |
| 100 KB download | 159.7 Mbps (239.1 Mbps) | 1.2 Mbps (1.2 Mbps) | 1.9 Mbps (1.9 Mbps) | 1.2 Mbps (1.2 Mbps) |
| 1 MB download | 427.9 Mbps (456.8 Mbps) | 12.4 Mbps (12.5 Mbps) | 19.3 Mbps (19.4 Mbps) | 12.3 Mbps (12.3 Mbps) |
| 10 MB download | 77.5 Mbps (150.1 Mbps) | 94.8 Mbps (112.2 Mbps) | 150.8 Mbps (166.2 Mbps) | 94.8 Mbps (110.9 Mbps) |
| 1 MB upload | 181.8 Mbps (185.8 Mbps) | 16.3 Mbps (16.6 Mbps) | 25.5 Mbps (25.7 Mbps) | 16.3 Mbps (23.9 Mbps) |
| UDP round trip | – | 168 ms (173 ms) | 107 ms (108 ms) | 169 ms (171 ms) |
| UDP loss | 100% (100%) | 0% (0%) | 0% (0%) | 0% (0%) |

## Round trip by hour (UTC, which is Ghana's time)

| Hour | Accra | Johannesburg | Frankfurt | Virginia |
| --- | --- | --- | --- | --- |
| 00:00 | 1.7 ms | 171 ms | 109 ms | 171 ms |
| 01:00 | 1.9 ms | 168 ms | 108 ms | 171 ms |
| 02:00 | 3.1 ms | 173 ms | 109 ms | 171 ms |
| 03:00 | 2.0 ms | 169 ms | 109 ms | 171 ms |
| 04:00 | 1.7 ms | 169 ms | 109 ms | 171 ms |
| 05:00 | 1.4 ms | 169 ms | 109 ms | 171 ms |
| 06:00 | 1.9 ms | 170 ms | 109 ms | 171 ms |
| 07:00 | 1.5 ms | 169 ms | 108 ms | 171 ms |
| 08:00 | 1.7 ms | 169 ms | 109 ms | 172 ms |
| 09:00 | 6.0 ms | 168 ms | 109 ms | 171 ms |
| 10:00 | 1.8 ms | 172 ms | 109 ms | 171 ms |
| 11:00 | 4.7 ms | 168 ms | 109 ms | 171 ms |
| 12:00 | 1.0 ms | 167 ms | 109 ms | 171 ms |
| 13:00 | 1.8 ms | 170 ms | 109 ms | 171 ms |
| 14:00 | 2.4 ms | 173 ms | 109 ms | 171 ms |
| 15:00 | 1.8 ms | 169 ms | 108 ms | 171 ms |
| 16:00 | 1.9 ms | 168 ms | 109 ms | 171 ms |
| 17:00 | 2.1 ms | 170 ms | 109 ms | 170 ms |
| 18:00 | 2.4 ms | 168 ms | 109 ms | 171 ms |
| 19:00 | 1.9 ms | 170 ms | 109 ms | 171 ms |
| 20:00 | 1.3 ms | 170 ms | 109 ms | 171 ms |
| 21:00 | 2.3 ms | 170 ms | 109 ms | 171 ms |
| 22:00 | 1.6 ms | 168 ms | 108 ms | 171 ms |
| 23:00 | 1.5 ms | 172 ms | 108 ms | 171 ms |

## Globalping: lowest ping per round, median across rounds

| Probe | Accra | Johannesburg | Frankfurt | Virginia |
| --- | --- | --- | --- | --- |
| Accra, GH (Workonline Communications(Pty)) | 0.6 ms | 168 ms | 106 ms | 169 ms |
| Cotonou, BJ (ISOCEL) | 202 ms | 261 ms | 112 ms | 177 ms |
| Freetown, SL (MICHCOM) | 97 ms | 198 ms | 135 ms | 212 ms |
| Lagos, NG (AfriCloud) | 126 ms | 75 ms | 109 ms | 189 ms |
| Lagos, NG (Melbikomas) | 205 ms | 78 ms | 109 ms | 161 ms |
| Lagos, NG (SiteHUB Agency) | 7.4 ms | 77 ms | 112 ms | 161 ms |
| Lagos, NG (Workonline Communications(Pty)) | 7.1 ms | 76 ms | 112 ms | 160 ms |
| Ouagadougou, BF (Burkina Faso Internet EXchange Point (BFIX)) | 152 ms | 102 ms | 124 ms | 185 ms |

## Globalping: HTTPS request timings, median (total, TCP, TLS, first byte)

| Probe | Accra | Johannesburg | Frankfurt | Virginia |
| --- | --- | --- | --- | --- |
| Accra, GH (Workonline Communications(Pty)) | 35 ms / 2.0 ms / 27 ms / 4.0 ms | 531 ms / 168 ms / 188 ms / 171 ms | 348 ms / 108 ms / 127 ms / 111 ms | 533 ms / 170 ms / 189 ms / 172 ms |
| Cotonou, BJ (ISOCEL) | 665 ms / 202 ms / 228 ms / 206 ms | 796 ms / 262 ms / 270 ms / 263 ms | 349 ms / 112 ms / 120 ms / 117 ms | 546 ms / 179 ms / 186 ms / 180 ms |
| Freetown, SL (MICHCOM) | 432 ms / 109 ms / 130 ms / 111 ms | 659 ms / 212 ms / 218 ms / 213 ms | 562 ms / 148 ms / 154 ms / 149 ms | 720 ms / 211 ms / 218 ms / 212 ms |
| Lagos, NG (AfriCloud) | 581 ms / 127 ms / 150 ms / 134 ms | 401 ms / 76 ms / 82 ms / 77 ms | 509 ms / 110 ms / 116 ms / 111 ms | 565 ms / 183 ms / 189 ms / 184 ms |
| Lagos, NG (Melbikomas) | 805 ms / 218 ms / 265 ms / 212 ms | 321 ms / 79 ms / 97 ms / 82 ms | 362 ms / 108 ms / 125 ms / 111 ms | 515 ms / 162 ms / 179 ms / 165 ms |
| Lagos, NG (SiteHUB Agency) | 267 ms / 9.0 ms / 42 ms / 13 ms | 420 ms / 79 ms / 96 ms / 81 ms | 528 ms / 113 ms / 129 ms / 115 ms | 665 ms / 162 ms / 179 ms / 164 ms |
| Lagos, NG (Workonline Communications(Pty)) | 216 ms / 7.5 ms / 34 ms / 12 ms | 258 ms / 78 ms / 86 ms / 80 ms | 486 ms / 113 ms / 121 ms / 114 ms | 506 ms / 162 ms / 170 ms / 164 ms |
| Ouagadougou, BF (Burkina Faso Internet EXchange Point (BFIX)) | 498 ms / 157 ms / 178 ms / 159 ms | 327 ms / 105 ms / 111 ms / 106 ms | 387 ms / 126 ms / 132 ms / 128 ms | 575 ms / 189 ms / 196 ms / 190 ms |

## Routes from the Accra data centre (latest)

**Johannesburg**: 102.68.83.254 (0 ms) → 196.44.100.237 (15 ms) → gw-wacs02.ecoband.net.gh (0 ms) → 197.157.66.117 (1 ms) → 41.78.188.139 (1751 ms) → 41.78.188.178 (94 ms) → 195.66.224.125 (94 ms) → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → *
**Frankfurt**: 102.68.83.254 (0 ms) → 196.44.100.237 (2 ms) → gw-wacs02.ecoband.net.gh (0 ms) → 197.157.66.117 (1 ms) → 41.78.188.137 (7955 ms) → 41.78.188.178 (94 ms) → 195.66.224.125 (94 ms) → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → *
**Virginia**: 102.68.83.254 (0 ms) → 196.44.100.237 (28 ms) → gw-wacs02.ecoband.net.gh (0 ms) → 197.157.66.117 (2 ms) → 41.78.188.139 (3012 ms) → 41.78.188.178 (94 ms) → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → * → *

## Routes from Globalping probes (mtr, latest)

**Accra, GH (Workonline Communications(Pty)) -> Accra**: _gateway AS? (0 ms) → 197.157.75.25 AS37271 (1 ms) → 41.78.188.138 AS37271 (1 ms) → 197.157.66.118 AS37271 (0 ms) → gw1-acc.ecoband.net.gh AS327814 (7 ms) → 196.44.100.238 AS327814 (0 ms) → bench-acc.dekka.cloud AS327813 (10 ms)
**Accra, GH (Workonline Communications(Pty)) -> Frankfurt**: _gateway AS? (0 ms) → 197.157.75.25 AS37271 (1 ms) → 41.78.188.178 AS37271 (94 ms) → 195.66.224.125 AS? (94 ms) → bench-fra.dekka.cloud AS396982 (107 ms)
**Accra, GH (Workonline Communications(Pty)) -> Johannesburg**: _gateway AS? (0 ms) → 197.157.75.25 AS37271 (1 ms) → 41.78.188.178 AS37271 (94 ms) → google1.lonap.net AS? (94 ms) → bench-jnb.dekka.cloud AS396982 (169 ms)
**Accra, GH (Workonline Communications(Pty)) -> Virginia**: _gateway AS? (0 ms) → 197.157.75.25 AS37271 (1 ms) → 41.78.188.178 AS37271 (94 ms) → 195.66.224.125 AS? (94 ms) → bench-iad.dekka.cloud AS396982 (171 ms)
**Cotonou, BJ (ISOCEL) -> Accra**: _gateway AS37090 (1 ms) → 41.86.224.214 AS37090 (102 ms) → 154.66.130.254 AS37090 (102 ms) → 193.251.142.59 AS5511 (101 ms) → 193.251.133.92 AS5511 (101 ms) → ae35.ear2.par1.sp.lumen.tech AS3356 (114 ms) → ae1.3118.ear5.lon1.neo.colt.net AS3356 (113 ms) → level3.ae0-2116.cr1.lon1.uk.ipv4.kaiaglobal.com AS3356 (108 ms) → 41.78.188.179 AS37271 (202 ms) → 41.78.188.138 AS37271 (202 ms) → 197.157.66.118 AS37271 (201 ms) → gw1-acc.ecoband.net.gh AS327814 (233 ms) → 196.44.100.238 AS327814 (202 ms) → bench-acc.dekka.cloud AS327813 (206 ms)
**Cotonou, BJ (ISOCEL) -> Frankfurt**: _gateway AS37090 (1 ms) → 41.86.224.214 AS37090 (102 ms) → 154.66.130.254 AS37090 (103 ms) → 193.251.142.59 AS5511 (101 ms) → 193.251.133.92 AS5511 (120 ms) → 72.14.218.40 AS15169 (101 ms) → bench-fra.dekka.cloud AS396982 (112 ms)
**Cotonou, BJ (ISOCEL) -> Johannesburg**: _gateway AS37090 (1 ms) → 41.86.224.214 AS37090 (102 ms) → 154.66.130.254 AS37090 (102 ms) → 193.251.142.59 AS5511 (103 ms) → 193.251.133.92 AS5511 (101 ms) → 72.14.202.232 AS15169 (101 ms) → bench-jnb.dekka.cloud AS396982 (263 ms)
**Cotonou, BJ (ISOCEL) -> Virginia**: _gateway AS37090 (1 ms) → 41.86.224.214 AS37090 (102 ms) → 154.66.130.254 AS37090 (105 ms) → 193.251.142.59 AS5511 (101 ms) → 193.251.133.92 AS5511 (121 ms) → 72.14.202.100 AS15169 (101 ms) → bench-iad.dekka.cloud AS396982 (178 ms)
**Freetown, SL (MICHCOM) -> Accra**: _gateway AS209049 (0 ms) → 172.31.46.177 AS? (0 ms) → 172.31.46.210 AS? (0 ms) → 172.31.6.1 AS? (1 ms) → 196.216.220.141 AS329109 (2 ms) → 185.35.140.164 AS60171 (76 ms) → 185.10.164.45 AS60171 (172 ms) → 185.35.142.39 AS60171 (172 ms) → 196.60.12.15 AS? (102 ms) → bench-acc.dekka.cloud AS327813 (100 ms)
**Freetown, SL (MICHCOM) -> Frankfurt**: _gateway AS209049 (0 ms) → 172.31.46.177 AS? (0 ms) → 172.31.46.169 AS? (0 ms) → 172.31.6.1 AS? (0 ms) → 196.216.220.141 AS329109 (1 ms) → 185.35.140.164 AS60171 (75 ms) → bench-fra.dekka.cloud AS396982 (136 ms)
**Freetown, SL (MICHCOM) -> Johannesburg**: _gateway AS209049 (0 ms) → 172.31.46.177 AS? (0 ms) → 172.31.46.210 AS? (0 ms) → 172.31.6.1 AS? (1 ms) → 196.216.220.141 AS329109 (1 ms) → 185.35.140.164 AS60171 (76 ms) → bench-jnb.dekka.cloud AS396982 (200 ms)
**Freetown, SL (MICHCOM) -> Virginia**: _gateway AS209049 (0 ms) → 172.31.46.177 AS? (0 ms) → 172.31.46.169 AS? (0 ms) → 172.31.6.1 AS? (1 ms) → 196.216.220.141 AS329109 (1 ms) → 185.35.140.164 AS60171 (88 ms) → bench-iad.dekka.cloud AS396982 (213 ms)
**Lagos, NG (AfriCloud) -> Accra**: _gateway AS209179 (0 ms) → 102.211.155.80 AS37468 (11 ms) → * → * → be2444.rcr51.opo01.atlas.cogentco.com AS174 (82 ms) → be2313.ccr31.bio02.atlas.cogentco.com AS174 (90 ms) → be3077.ccr31.mrs02.atlas.cogentco.com AS174 (109 ms) → be4515.rcr51.mrs01.atlas.cogentco.com AS174 (109 ms) → 149.11.240.202 AS174 (198 ms) → * → 41.78.188.179 AS37271 (188 ms) → 41.78.188.138 AS37271 (189 ms) → gw-wacs02.ecoband.net.gh AS327814 (126 ms) → gw1-acc.ecoband.net.gh AS327814 (142 ms) → 196.60.12.15 AS? (126 ms) → bench-acc.dekka.cloud AS327813 (131 ms)
**Lagos, NG (Melbikomas) -> Accra**: _gateway AS8849 (0 ms) → 10.32.1.34 AS? (3 ms) → 154.113.97.161 AS37282 (2 ms) → 41.75.80.26 AS37282 (2 ms) → 212.133.84.17 AS3356 (69 ms) → ae1.3505.edge4.mrs1.neo.colt.net AS3356 (96 ms) → 213.242.121.6 AS3356 (93 ms) → * → 41.78.188.179 AS37271 (206 ms) → 41.78.188.136 AS37271 (205 ms) → 197.157.66.118 AS37271 (206 ms) → gw1-acc.ecoband.net.gh AS327814 (215 ms) → 196.44.100.238 AS327814 (205 ms) → bench-acc.dekka.cloud AS327813 (220 ms)
**Lagos, NG (Melbikomas) -> Frankfurt**: _gateway AS8849 (0 ms) → ixpn-lagos-02.google.com AS? (1 ms) → bench-fra.dekka.cloud AS396982 (107 ms)
**Lagos, NG (Melbikomas) -> Johannesburg**: _gateway AS8849 (0 ms) → ixpn-lagos-02.google.com AS? (1 ms) → bench-jnb.dekka.cloud AS396982 (78 ms)
**Lagos, NG (Melbikomas) -> Virginia**: _gateway AS8849 (0 ms) → ixpn-lagos-02.google.com AS? (1 ms) → bench-iad.dekka.cloud AS396982 (161 ms)
**Lagos, NG (SiteHUB Agency) -> Accra**: _gateway AS214354 (1 ms) → 197.157.72.69 AS37271 (1 ms) → 41.78.188.170 AS37271 (8 ms) → 197.157.66.118 AS37271 (7 ms) → gw1-acc.ecoband.net.gh AS327814 (61 ms) → 196.44.100.238 AS327814 (7 ms) → bench-acc.dekka.cloud AS327813 (15 ms)
**Lagos, NG (SiteHUB Agency) -> Frankfurt**: _gateway AS214354 (1 ms) → 197.157.72.69 AS37271 (0 ms) → * → 41.78.188.55 AS37271 (13 ms) → bench-fra.dekka.cloud AS396982 (113 ms)
**Lagos, NG (SiteHUB Agency) -> Johannesburg**: _gateway AS214354 (1 ms) → 197.157.72.69 AS37271 (0 ms) → * → 41.78.188.55 AS37271 (7 ms) → bench-jnb.dekka.cloud AS396982 (79 ms)
**Lagos, NG (SiteHUB Agency) -> Virginia**: _gateway AS214354 (5 ms) → 197.157.72.69 AS37271 (0 ms) → * → 41.78.188.55 AS37271 (9 ms) → bench-iad.dekka.cloud AS396982 (164 ms)
**Lagos, NG (Workonline Communications(Pty)) -> Accra**: _gateway AS? (0 ms) → 197.157.75.33 AS37271 (0 ms) → 41.78.188.170 AS37271 (7 ms) → 197.157.66.118 AS37271 (7 ms) → gw1-acc.ecoband.net.gh AS327814 (24 ms) → 196.44.100.238 AS327814 (7 ms) → bench-acc.dekka.cloud AS327813 (24 ms)
**Lagos, NG (Workonline Communications(Pty)) -> Frankfurt**: _gateway AS? (0 ms) → 197.157.75.33 AS37271 (0 ms) → 41.78.188.33 AS37271 (16 ms) → 142.251.206.190 AS15169 (1 ms) → bench-fra.dekka.cloud AS396982 (112 ms)
**Lagos, NG (Workonline Communications(Pty)) -> Johannesburg**: _gateway AS? (0 ms) → 197.157.75.33 AS37271 (0 ms) → 41.78.188.33 AS37271 (10 ms) → 142.251.206.190 AS15169 (1 ms) → bench-jnb.dekka.cloud AS396982 (76 ms)
**Lagos, NG (Workonline Communications(Pty)) -> Virginia**: _gateway AS? (0 ms) → 197.157.75.33 AS37271 (0 ms) → * → 41.78.188.55 AS37271 (13 ms) → bench-iad.dekka.cloud AS396982 (164 ms)
**Ouagadougou, BF (Burkina Faso Internet EXchange Point (BFIX)) -> Accra**: _gateway AS328010 (1 ms) → et-0-0-13.0.pe2.oua2.vts.bf AS37721 (1 ms) → ae353-353.bb2.oua2.vts.bf AS37721 (1 ms) → ae26-2026.bb2.dak1.vts.bf AS37721 (11 ms) → 102.67.99.156 AS37721 (15 ms) → et-0-0-3-716.br2.acc1.vts.bf AS37721 (16 ms) → 196.44.100.238 AS327814 (153 ms) → bench-acc.dekka.cloud AS327813 (177 ms)
**Ouagadougou, BF (Burkina Faso Internet EXchange Point (BFIX)) -> Frankfurt**: _gateway AS328010 (0 ms) → et-0-0-13.0.pe2.oua2.vts.bf AS37721 (1 ms) → ae353-353.bb2.oua2.vts.bf AS37721 (1 ms) → ae50-0.bb2.dak1.vts.bf AS37721 (4 ms) → ae32-2022.bb1.acc1.vts.bf AS37721 (16 ms) → et-0-0-12-2084.bb2.los1.vts.bf AS37721 (25 ms) → bench-fra.dekka.cloud AS396982 (125 ms)
**Ouagadougou, BF (Burkina Faso Internet EXchange Point (BFIX)) -> Johannesburg**: _gateway AS328010 (1 ms) → et-0-0-13.0.pe2.oua2.vts.bf AS37721 (1 ms) → ae253-253.bb1.oua1.vts.bf AS37721 (1 ms) → et-1-1-5.2089.bb1.dak1.vts.bf AS37721 (7 ms) → ae32-2022.bb1.acc1.vts.bf AS37721 (17 ms) → et-0-0-12-2084.bb2.los1.vts.bf AS37721 (25 ms) → bench-jnb.dekka.cloud AS396982 (102 ms)
**Ouagadougou, BF (Burkina Faso Internet EXchange Point (BFIX)) -> Virginia**: _gateway AS328010 (1 ms) → et-0-0-13.0.pe2.oua2.vts.bf AS37721 (1 ms) → ae253-253.bb1.oua1.vts.bf AS37721 (1 ms) → et-1-1-5.2089.bb1.dak1.vts.bf AS37721 (4 ms) → irb-2070.bb1.acc1.vts.bf AS37721 (16 ms) → et-0-0-11-2031.bb1.los1.vts.bf AS37721 (39 ms) → bench-iad.dekka.cloud AS396982 (186 ms)

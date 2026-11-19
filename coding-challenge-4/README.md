# 🚚 Coding Challenge #4: Merger Day at Mini Delivery/works

## 🎯 1. Context

### 1.1 Introduction

Mini Delivery/works is a fictional delivery company in Vietnam. When a shop sends a parcel, the company quotes a price and a delivery date. You will build the engine behind that quote.

On 1 July 2025 Vietnam merged its 63 provinces into 34 and removed the district level. Your engine serves orders from both sides of that day, and two months later it takes one more rule change.

> ⏱️ **Timebox: 6–8 hours total.**

We expect you to work with an AI coding assistant and explain how your team used it during the presentation.

### 1.2 System Overview

```
  order ──> pricing ──> dispatch ──> quote
             │            │
             │            hub, pickup day, promised date
             lane, chargeable weight, fuel surcharge

  merger day  from 1 July 2025 addresses and lanes change, old orders keep their price
  distance    from 1 September 2025 the vehicle depends on the shortest road route
  service     a repeated order gets the same quote, 1,000,000 orders in under a minute
```

A quote depends only on the order, the data files, and the earlier orders of the same run. The engine never asks what day it is today: the date on the order decides which rules apply.

---

## 🧩 2. Details

### 2.1 Example Data

The folder [`data/`](data/) holds an example data set. The format is fixed but the values are only examples: the judges use their own folder, so no value from these files belongs in your code. You may build your own data sets, as long as your README says where the values come from.

| File | What is in it |
| --- | --- |
| `fees.json` | fees, transit days, block sizes and the fuel surcharge settings |
| `provinces.json`, `regions.json` | the 63 old province names (the 34 new ones are among them) and their region |
| `fuel-prices.json` | petrol and diesel prices, one board per change |
| `capacity.json` | daily kg per hub, one board per change. A motorbike hub is a province, a truck hub is a region. |
| `mapping.csv` | old ward to new ward, and the merger date (`effective_from`, the same on every row) |
| `roads.json` | roads between provinces in km, and the settings of the distance rule |

A board applies from its date until the next board. Look up fuel prices with the order date and hub capacity with the pickup day.

Every data set keeps a few promises. Each capacity board is complete: it lists every hub, except that a hub retired by the merger is left out from the merger date. The first capacity board is no later than the first fuel board. A later board never lowers a hub, a merged hub is at least as large as each hub it takes over, and orders only use hubs that are in the file. So a hub never gets smaller on a later day. A new ward name appears once per province. All wards of an old province go to one new province, all old wards of a new ward come from one old province, and no ward is split. No two roads join the same pair. Names are in Unicode NFC, and input names differ from them only by prefix, case and spacing. Hyphenated names never appear in the input.

### 2.2 Pricing Rules

1. Start with the lane (AC-01), which says how the two addresses relate.

```
same province?  no ──> same region?  no ──> CROSS_REGION   (truck)
   │ yes                   │ yes
   │                       └──────────────> SAME_REGION    (truck)
same district?  no ───────────────────────> SAME_PROVINCE  (motorbike)
   │ yes
   └──────────────────────────────────────> LOCAL          (motorbike)
```

   The vehicle in the diagram holds until the distance rule starts (2.5). Compare names, from the data or from the input, with these steps in order:

```
1. normalise to Unicode NFC
2. trim, and turn every run of spaces into one space
3. ignore case
4. drop one leading prefix: Tỉnh, Thành phố, TP, Quận, Q, Huyện, H, Thị xã, TX, Phường, P, Xã, X, Thị trấn, TT
   a prefix counts only when a dot, a space, or both follow it

"Quận 1", "Q.1", "Q. 1" and "q 1" are the same name. "HCM" is not Hồ Chí Minh.
```

2. Then work out the base fee (AC-02). A parcel is charged by the larger of its actual and its volume weight. The first `first_block_g` grams cost `first_block`, and every started `extra_block_g` after that adds `per_extra_block`.

```
volume_g   = length_cm x width_cm x height_cm / 5, rounded up     (30 x 20 x 20 / 5 = 2,400 g)
chargeable = max(weight_g, volume_g)                               (no dimensions: weight_g)
2,200 g    = first_block + 3 x per_extra_block                     (1,200 g over 1,000, three started 500 g)
```

3. Add the fuel surcharge (AC-03). A motorbike follows the petrol price and a truck the diesel price. For each full 1,000 VND above the reference price, add `surcharge_percent_per_step` (3%) of the base fee, up to `cap_percent` (15%), and round the surcharge down. Below the reference there is no surcharge.
4. The total (AC-03) is the base fee plus the surcharge, rounded to the nearest 1,000 VND, half up. So there are three roundings, each in its own step: the volume weight up to a gram, the surcharge down to a dong, and the total half up to 1,000 VND. Every value in between is a whole number.
5. Before all this, check the input (AC-04). Run the checks in this order and return the first that fails: NO_PRICE_BOARD, UNKNOWN_PROVINCE, MISSING_DISTRICT (only before the merger), BAD_WEIGHT, BAD_DIMENSIONS. Each check looks at both addresses before the next check runs. The ward is not used before the merger.

### 2.3 Dispatch (AC-05)

1. A motorbike order leaves from the hub of its origin province, and a truck order from the hub of its origin region.
2. The pickup day is the first day, from the order date on, when the hub still has room for the chargeable weight. Count per hub and per day, in order of arrival. A day may be filled exactly, and a rejected order takes no room.
3. The promised date is the pickup day plus the `transit_days` of the lane.
4. An order heavier than the daily capacity of its hub on the order date is rejected with TOO_HEAVY. This is the last check.

```
hub of 60 kg per day, three orders of 25 kg on 06-28
06-28  [25][25]      the third would make 75 > 60
06-29  [25]
```

### 2.4 Merger Day (AC-06)

The merger date is the `effective_from` of `mapping.csv`. Read it from the file and never write it in your code.

1. When an order is dated on or after the merger, convert both addresses with the mapping before you work out the lane. An address with a district is in the old format. An address with no district (absent or empty) is in the new format, and its province and ward must be a new address of the mapping, otherwise the order is rejected with UNKNOWN_NEW_ADDRESS. One order may mix the two formats.
2. From the merger date, LOCAL means the same province and the same ward. Nothing else about lanes and fees changes.
3. An order dated before the merger is not converted. It needs a district on both addresses and keeps the district rule.
4. An old-format address whose province, district and ward are not a row of the mapping is rejected with UNKNOWN_WARD.
5. The mapping checks come after the input checks. Check `from` first, then `to`.
6. Hubs follow provinces. From the merger date a motorbike order uses the hub of its new province. An older order that slips past the merger date is picked up from the hub its province has on that day: a Bình Dương order that slips to 1 July takes room in the Hồ Chí Minh hub, while its lane and price stay the same. Truck hubs do not change.

Backward compatibility (AC-07) is the heart of this challenge, and we ask for a test of each of its two rules.

1. Old orders keep their price. An order dated before the merger keeps its addresses, lane, base fee, surcharge, total, vehicle and transit days, exactly as if there were no mapping. Only dispatch looks ahead: a pickup day on or after the merger uses the hub of that day (rule 6), so a slipping order can get a later pickup date.
2. Old addresses still work. After the merger, the old and the new address of a place give the same quote, apart from `order_id`. Compare them in two runs that start from the same state, because in one run the first order takes room that the second one then lacks.

### 2.5 Distance (AC-10)

From a second date, the `effective_from` of `roads.json`, the vehicle is chosen by road distance and no longer by lane.

1. Every address has a depot, named after an old province. For an old-format address it is simply the province. For a new-format address it is the old province of the wards that became its ward, which you find in the mapping: "Thủ Dầu Một, Hồ Chí Minh" has the depot Bình Dương.
2. The distance is the length of the shortest route between the two depots, over as many roads as needed. Roads go both ways, and the same depot is 0 km.

```
Bình Phước ───120─── Đồng Nai        direct road            120 km
     │                  │
     85                 25           through Bình Dương     85 + 25 = 110 km   <── the distance
     └─── Bình Dương ───┘
```

3. When no route joins the two depots, reject the order with NO_ROUTE. A depot that is not in `roads.json` has no route to any other depot. This check comes after the mapping checks.
4. The order goes by motorbike when the distance is at most `motorbike_max_km`, and by truck otherwise. The vehicle decides the fuel and the hub as before, and the motorbike hub is still the new province, not the depot. The lane and the base fee do not change.
5. Long routes take longer: add one day to the promised date for every full `extra_day_km` of the distance.
6. An order dated before the rule starts keeps the vehicle of its lane, gets no extra days and has `distance_km` null, whenever it is picked up.

You may build your own `roads.json` from a map or a geocoding service. The engine still reads distances from its data folder.

### 2.6 Worked Examples

Each block starts from empty hubs. All 22 example orders, with the lines to print, are in [`examples/worked-examples.json`](examples/worked-examples.json).

```
order-0001  2025-06-30  1,500 g  Phú Cường, Thủ Dầu Một, Bình Dương -> Bến Nghé, Quận 1, Hồ Chí Minh
  SAME_REGION, truck          base 30,000 + 1 x 5,000 = 35,000
  diesel 19,707               1 step above 18,000, 3% = 1,050, total 36,000 (36,050 rounded)
  pickup 06-30, promised 07-02

order-0007  2025-07-07  the same order after the merger
  SAME_PROVINCE, motorbike    base 22,000 + 1 x 4,000 = 26,000
  petrol 19,906               below 20,000, no surcharge, total 26,000
  pickup 07-07, promised 07-08
order-0008  2025-07-07  Thủ Dầu Một, Hồ Chí Minh -> Sài Gòn, Hồ Chí Minh       the same quote as order-0007

order-0018  2025-09-08  1,500 g  Trấn Biên, Đồng Nai -> Sài Gòn, Hồ Chí Minh
  SAME_REGION, 35 km          motorbike, hub Đồng Nai, base 35,000
  petrol 21,150               1 step, 3% = 1,050, total 36,000
  pickup 09-08, promised 09-10
  dated 08-31                 truck, diesel 20,480, 2 steps, 6% = 2,100, total 37,000
```

### 2.7 Input and Output

Your `run.sh` (or `run.cmd`) reads one order per line as JSON on stdin and writes one quote per line on stdout. It takes the data from `./data`, or from the folder given as its first argument.

Every input line is valid JSON. Its keys are below. `order_id`, `created_at`, `from` and `to` are always right. Nothing is promised about the weight and the dimensions: the input checks decide.

| Key | What it holds |
| --- | --- |
| `order_id` | a string |
| `created_at` | a real date, `YYYY-MM-DD` |
| `from`, `to` | objects with a string `province`, and an optional string `district` and `ward` |
| `weight_g` | grams. BAD_WEIGHT unless it is a positive whole number |
| `length_cm`, `width_cm`, `height_cm` | optional. When at least one is there, BAD_DIMENSIONS unless all three are positive whole numbers |

A whole number is a JSON number with no fraction, so `1500` and `1500.0` are the same. A missing key, `null`, a string such as `"1500"`, `true` and `1.5` are not whole numbers. A dimension that is `null` counts as missing, and one or two dimensions alone give BAD_DIMENSIONS.

```json
{"order_id": "order-0001", "created_at": "2025-06-30", "weight_g": 1500,
 "from": {"ward": "Phú Cường", "district": "Thủ Dầu Một", "province": "Bình Dương"},
 "to":   {"ward": "Bến Nghé",  "district": "Quận 1",       "province": "Hồ Chí Minh"}}
```

```json
{"order_id": "order-0001", "status": "quoted", "lane": "SAME_REGION", "chargeable_g": 1500, "base": 35000, "surcharge": 1050, "total": 36000,
 "distance_km": null, "pickup_date": "2025-06-30", "promised_date": "2025-07-02", "duplicate": false, "reason": null}
```

A rejected order keeps its `order_id` and `duplicate`, has `status: "rejected"` and a `reason`, and null for the eight values of the quote:

```json
{"order_id": "order-0004", "status": "rejected", "lane": null, "chargeable_g": null, "base": null, "surcharge": null, "total": null,
 "distance_km": null, "pickup_date": null, "promised_date": null, "duplicate": false, "reason": "NO_PRICE_BOARD"}
```

A quoted order has `reason: null`. `distance_km` is null before the distance rule starts. `duplicate` is false the first time an `order_id` is seen. Numbers are JSON integers, so print `36000`, not `36000.0`.

| Reason | When |
| --- | --- |
| NO_PRICE_BOARD | dated before the first fuel board |
| UNKNOWN_PROVINCE | province not in `provinces.json` |
| MISSING_DISTRICT | dated before the merger, district missing |
| BAD_WEIGHT | weight not a positive whole number |
| BAD_DIMENSIONS | some dimensions given, but not all three as positive whole numbers |
| UNKNOWN_WARD | old-format address not in the mapping |
| UNKNOWN_NEW_ADDRESS | new-format address not in the mapping |
| NO_ROUTE | no route between the two depots |
| TOO_HEAVY | heavier than the daily capacity of the hub |

### 2.8 Running as a Service

1. Shops sometimes send an order twice (AC-08). When an `order_id` was already seen in this run, quoted or rejected, print its first output line again with `duplicate: true`, even if the rest of the order is different. A duplicate takes no capacity.
2. On the 11.11 sale day the shops send 1,000,000 orders (AC-09). `./run.sh load/data < load/orders-1m.jsonl` must finish in under 60 seconds on a laptop. Unzip the file in `load/` first. The hubs fill up for a week.

### 🚫 2.9 Out of Scope

You do not need live fuel prices or free-text addresses. Parcels are not routed from hub to hub. There is no second merger round. Payment, a database, an API, login and a UI are not needed.

---

## 📋 3. Deliverables and Evaluation

### 3.1 Deliverables

On top of the season README, please hand in:

- `run.sh`, as described in 2.7
- tests for the two backward compatibility rules, a hub that fills up, a duplicate order, and a route that is not the direct road
- one command that runs `examples/worked-examples.json` against your engine
- a README with a diagram of your engine, how you handle the merger date and why, your answer to the scale question, and a few lines on how you worked with AI

For the lines on AI, tell us what you gave the assistant (keep the instruction files in your repository), what it did and what you did, and how you checked its work, with one mistake it made and how you found it.

At the demo, the judge hands you one small rule change. You get 10 minutes, with any tool you like, then the judge runs a scenario file for it.

The scale question: next year the sale brings 20,000,000 orders in one day, on several machines. Which parts can run in parallel, which must keep the order of arrival, and where does the state live?

### 🏆 3.2 Evaluation

**The submission accounts for 70% of the score. The remaining 30% comes from the team presentation.**

| Assessment | Weight | What we look for |
| --- | --- | --- |
| Submission | 70% | Correct results, pricing and temporal models, service design, tests, and documentation. |
| Presentation | 30% | The team's understanding of the problem, explanation of its solution and design choices, use of engineering best practices, and how it applied AI to this challenge. |

In the presentation, explain the system design and its trade-offs, demonstrate the solution, and show how your team used and checked AI-generated work. Be ready to answer questions and complete the live rule change described above. Bonuses follow the season README.

The submission's 70 points are allocated as follows:

| Category | Points |
| --- | ---: |
| Correct results on the judges' scenarios | 18 |
| Pricing model design | 8 |
| Temporal design | 8 |
| Service design | 8 |
| Working with AI | 7 |
| Tests & Reliability | 10 |
| Documentation & Communication | 11 |
| **Submission total** | **70** |

Working with AI in the submission is assessed from the team's documented use and checks. The presentation assesses how the team explains and demonstrates that work.

Correct results are required. We also look for these design choices:

- **Pricing model design:** every fee, block size, fuel setting, region, capacity and road is read from data. Lane, chargeable weight, base fee, surcharge and total are separate steps. One function normalises every name, and the shortest route is its own function. Money stays in whole numbers and is rounded only where 2.2 says.
- **Temporal design:** one "latest board on or before the date" lookup serves fuel and capacity. The order date chooses the rule set, and the merger and roads dates are read from data and handled the same way. Pricing before the merger does not convert addresses. Nothing reads the system clock.
- **Service design:** the first output of each `order_id` is kept by key. Input is read and written one line at a time. Data is loaded once, and a distance is searched once per pair of depots.

### 💡 3.3 Creative Bonus Ideas

- An LLM that turns a messy address into the structured input
- Your own `roads.json` for all 63 provinces, built from a map or from coordinates
- Real fuel prices read from a public page, or a third vehicle type

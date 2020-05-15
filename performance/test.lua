json = require "json"

-- This test script sends a HTTP 1.1. POST with 2580 bytes body to Jabba
-- with 8 concurrent threads and records various performance metrics such as requests per second
-- in an output file

wrk.method = "POST"
wrk.body   = [[{
    \"port\": 8080,
    \"mode\": \"TLS\",
    \"connection\": {
        \"server\": {
            \"readTimeoutSeconds\": 121,
            \"roundTripTimeoutSeconds\": 242
        },
        \"client\": {
            \"socketTimeoutSeconds\": 7,
            \"readTimeoutSeconds\": 121,
            \"keepAliveTimeoutSeconds\": 300,
            \"maxAttempts\": 4,
            \"poolSize\" : 1024
        }
    },
    \"policies\": {
        \"ab\": [
            {
                \"label\": \"green\",
                \"weight\": 0.8
            },
            {
                \"label\": \"blue\",
                \"weight\": 0.2
            }
        ]
    },
    \"routes\": [
        {
            \"path\": \"/about\",
            \"alias\": \"aboutJabba\"
        },
        {
            \"path\": \"/customer\",
            \"alias\": \"customer\",
            \"policy\": \"ab\"
        },
        {
            \"path\": \"/billing\",
            \"alias\": \"billing\",
            \"policy\": \"ab\"
        },
        {
            \"path\": \"/posting\",
            \"alias\": \"posting\"
        }
    ],
    \"resources\": {
        \"customer\": [
            {
                \"labels\": [
                    \"xblue\"
                ],
                \"upstream\": {
                    \"scheme\": \"http\",
                    \"host\": \"localhost\",
                    \"port\": 8081
                }
            },
            {
                \"labels\": [
                    \"xgreen\"
                ],
                \"upstream\": {
                    \"scheme\": \"http\",
                    \"host\": \"localhost\",
                    \"port\": 8082
                }
            }
        ],
        \"billing\": [
            {
                \"labels\": [
                    \"blue\"
                ],
                \"upstream\": {
                    \"scheme\": \"http\",
                    \"host\": \"localhost\",
                    \"port\": 8083
                }
            },
            {
                \"labels\": [
                    \"green\"
                ],
                \"upstream\": {
                    \"scheme\": \"http\",
                    \"host\": \"localhost\",
                    \"port\": 8083
                }
            }
        ],
        \"posting\": [
            {
                \"upstream\": {
                    \"scheme\": \"http\",
                    \"host\": \"localhost\",
                    \"port\": 8083
                }
            }
        ]
    }
}]]
wrk.headers["Content-Type"] = "application/json"

function setup(thread)
   thread0 = thread0 or thread
end

function init(args)
   file = args[1] or "performancetest_results.json"
end

function done(summary, latency, requests)
   file = io.open(thread0:get("file"), "w")

   percentiles = {}

   for _, p in pairs({ 1, 5, 10, 25, 50, 75, 90, 95, 99, 99.999 }) do
      k = string.format("%g%%", p)
      v = latency:percentile(p)
      percentiles[k] = v
   end

   file:write(json.encode({
       duration = summary.duration,
       requests = summary.requests,
       requestspersecond = summary.requests/(summary.duration//1000000),
       bytespersecond = summary.bytes/(summary.duration//1000000),
       bytes    = summary.bytes,
       errors   = summary.errors,
       latency  = {
          min         = latency.min,
          max         = latency.max,
          mean        = latency.mean,
          stdev       = latency.stdev,
          percentiles = percentiles,
       },
   }))
   file:close()
end

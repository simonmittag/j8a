json = require "json"
math = require "math"

-- This test script sends a HTTP 1.1. POST with 2580 bytes body to j8a
-- with 8 concurrent threads and records various performance metrics such as requests per second
-- in an output file

wrk.method = "POST"
wrk.body = [[{
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
            \"alias\": \"aboutj8a\"
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
wrk.headers["Authorization"] = "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6IjBmM2EzMmMxLTZiYTQtNGFmZi1iZTZmLWI1YjNlYzdlMjc5OSIsImlhdCI6MTYwODkzODEzMn0.VYPpEo0DQGfGUMbWj4LwbhGHpn49tsA6xCHXMoSYLGCee3pnrt2ZUNUizZyUUE2dY95eev7c6TdaHFuMy9l8_ILROfF22njEGzzPrIelAMGrYwJto3p6xyWVjgW3iTWKUODDgcKuvCtP6hMGj5wAw0fGdAXP3PXaUoyYUn3YEDFgBTV8pFIgy2Lwe2uNkqQ8PAYHqwPJdMyi8lejmWvHVTBgxrBanGEvkIovMTVwABV5AF5g4Hh5TOqvAR-Ap6h-pd5tFJr5K3hMmAfWL-IW379l0PLJQ508sUsbadUTLid6O-TdwsylfUYKRQEs-ymG8iK2Crm52jKMDro2caAMo_PQ1rGjDSZgfs7mKe-TnkKGkxfpx40uVs-EidcGnw05s972i3vOZTzCWuXrNQFDj6u4y76xdKb-W1q7KdqHKlRl1Sit_rti44TXGKumxytXcbzyss4TtJ72REdP_oaPLxHpYNY1W0KV86yJ9q30kLYZrxx8UD1cGbT2EKgr78Z-GmXZOqpSKDYOZ49Icma2UZzCxFAR8bYq5IMWW9NDy2WmbQ2rkvxxZbxgSKGhgZSDdUh-pVk-2utcSVl0cyvhH-sI4edYH2KEnKi4CFtIOB3AiatKqLRGRzNUTzA7BvVCiQ8hrGjIwjOWDN139GVoZ5ge-wizxy1bZ_5v8BTdvgY"

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
        requestspersecond = math.floor(summary.requests / (summary.duration / 1000000)),
        bytespersecond = math.floor(summary.bytes / (summary.duration / 1000000)),
        bytes = summary.bytes,
        errors = summary.errors,
        latency = {
            min = latency.min,
            max = latency.max,
            mean = latency.mean,
            stdev = latency.stdev,
            percentiles = percentiles,
        },
    }))
    file:close()
end

json = require "json"
math = require "math"

-- This test script sends a HTTP 1.1. POST with 2580 bytes body to j8a
-- with 8 concurrent threads and records various performance metrics such as requests per second
-- in an output file

wrk.method = "GET"
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

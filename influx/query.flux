from(bucket: "testbucket")
    |> range(start: 2024-12-11T00:00:00Z, stop: 2024-12-11T00:05:00Z)
    |> filter(fn: (r) => r._measurement == "bleTemperature")
    |> aggregateWindow(every: 1m, fn:mean)
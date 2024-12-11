from(bucket: "testbucket")
    |> range(start: 2024-12-11T00:00:00Z, stop: 2024-12-11T00:02:00Z)
    |> filter(fn: (r) => r._measurement == "bleTemperature" and (r._field == "TempF" or r._field == "Humidity" or r._field == "BatteryPercent"))
    |> window(every: 1m)
    |> reduce(
    identity: {count: 0.0, sum: 0.0, avg: 0.0, min: 1000.0, max: -1000.0},
    fn: (r, accumulator) => ({
        // Increment the counter on each reduce loop
        count: accumulator.count + 1.0,
        // Add the _value to the existing sum
        sum: accumulator.sum + r._value,
        // Divide the existing sum by the existing count for a new average
        avg: (accumulator.sum + r._value) / (accumulator.count + 1.0),
        min: if r._value < accumulator.min then r._value else accumulator.min,
        max: if r._value > accumulator.max then r._value else accumulator.max,
    }),
    )
    |> drop(columns: ["sum"])
    |> duplicate(column: "_start", as: "_time")
    |> window(every: inf)
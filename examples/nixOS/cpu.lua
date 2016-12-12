hosts = getFieldChoice("host", "host")
for i, host in ipairs(hosts) do
    events = { }
    setToken("host", host, "host")

    totalCPU = math.random() + math.random(options["minCPU"], options["maxCPU"])
    pctUserAll = round(math.random() * totalCPU, 2)
    pctSystemAll = round(totalCPU - pctUserAll, 2)
    pctIowaitAll = round(math.random() * 0.1, 2)
    pctIdleAll = round(100 - pctUserAll - pctSystemAll, 2)

    setToken("pctUserAll", pctUserAll)
    setToken("pctSystemAll", pctSystemAll)
    setToken("pctIowaitAll", pctIowaitAll)
    setToken("pctIdleAll", pctIdleAll)
    l = getLine(0)
    l = replaceTokens(l)
    table.insert(events, l)

    for i=1,tonumber(options["numCPUs"]),1
    do
        oneCPU = round(math.random() * totalCPU, 2)
        pctUser = round(math.random() * oneCPU, 2)
        pctSystem = round(oneCPU - pctUser, 2)
        pctIowait = round(math.random() * 0.1, 2)
        pctIdle = round(100 - pctUser - pctSystem, 2)
        setToken("CPU", oneCPU)
        setToken("pctUser", pctUser)
        setToken("pctSystem", pctSystem)
        setToken("pctIowait", pctIowait)
        setToken("pctIdle", pctIdle)
        l = getLine(1)
        l = replaceTokens(l)
        table.insert(events, l)
    end
    send(events)
end
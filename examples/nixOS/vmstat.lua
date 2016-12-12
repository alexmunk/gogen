hosts = getFieldChoice("host", "host")
for i, host in ipairs(hosts) do
    setToken("host", host, "host")
    events = { }
    l = getLine(0)
    l = replaceTokens(l)
    table.insert(events, l)

    memTotalMB = options["totalMB"]
    memUsedPct = round(math.random() + math.random(options["minMemUsedPct"], options["maxMemUsedPct"]),1)
    memFreePct = round(100-memUsedPct, 1)
    memFreeMB = memTotalMB * memFreePct / 100
    memUsedMB = memTotalMB - memFreeMB
    pgPageOut = math.random(1000000, 10000000)
    swapUsedPct = round(math.random(0, 200) / 100, 1)
    pgSwapOut = math.random(10000, 100000)
    cSwitches = math.random(1000000,5000000)
    interrupts = math.random(1000000,4000000)
    forks = math.random(10000,100000)
    processes = math.random(100,500)
    threads = math.random(1000,3000)
    loadAvg1mi = math.random(100,500) / 100
    waitThreads = 0
    interruptsPS = math.random(100000, 1000000) / 100
    pgPageInPS = math.random(1000, 15000) / 100
    pgPageOutPS = math.random(100000, 2000000) / 100

    tokens = { "memTotalMB", "memUsedPct", "memFreePct", "memFreeMB", "memUsedMB", "pgPageOut", "swapUsedPct", "pgSwapOut", "cSwitches", "interrupts", "forks", "processes", "threads", "loadAvg1mi", "waitThreads", "interruptsPS", "pgPageInPS", "pgPageOutPS"}
    for i, t in ipairs(tokens) do
        setToken(t, _G[t])
    end
    l = getLine(1)
    l = replaceTokens(l)
    table.insert(events, l)
    send(events)
end
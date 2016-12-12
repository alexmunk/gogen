hosts = getFieldChoice("host", "host")
disks = getChoice("disks")
maxOps = tonumber(options["maxOps"])
avgKB = tonumber(options["avgKB"])
maxTime = tonumber(options["maxTime"])
for i, host in ipairs(hosts) do
    setToken("host", host, "host")
    
    events = { }
    l = getLine(0)
    l = replaceTokens(l)
    table.insert(events, l)

    for i=1,#disks do
        if options["highWrites"] > 0 and optoins["highReads"] > 0 then
            rrps = (math.random(0, 50) / 100) * maxOps
            wrps = maxOps - rrps
            qtime =(math.random(50, 100) / 100) * maxTime
        else if options["highWrites"] > 0 then
            rrps = (math.random(0, 10) / 100) * maxOps
            wrps = (math.random(0, 80) / 100) * maxOps
            qtime =(math.random(50, 80) / 100) * maxTime
        else if options["highReads"] > 0 then
            rrps = (math.random(0, 80) / 100) * maxOps
            wrps = (math.random(0, 10) / 100) * maxOps
            qtime =(math.random(50, 80) / 100) * maxTime
        else
            rrps = (math.random(0, 2) / 100) * maxOps
            wrps = (math.random(0, 2) / 100) * maxOps
            qtime = 0
        end end end
        device = disks[i]
        rkbps = round(rrps * (math.random(0, 20) / 100) * avgKB, 2)
        wkbps = round(wrps * (math.random(0, 20) / 100) * avgKB, 2)
        avgsvc = round((math.random(0, 1) / 100) * maxTime, 2)
        avgwait = round(avgsvc + qtime, 2)
        bwutil = round(((rrps + wrps) * avgsvc / 1000) * 100, 2)

        tokens = { "device", "rrps", "wrps", "rkbps", "wkbps", "avgwait", "avgsvc", "bwutil" }
        for i, t in ipairs(tokens) do
            setToken(t, _G[t])
        end
        l = getLine(1)
        l = replaceTokens(l)
        table.insert(events, l)
    end
    send(events)
end
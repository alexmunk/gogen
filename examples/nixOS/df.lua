hosts = getFieldChoice("host", "host")
mounts = getFieldChoice("disks", "mount")
disks = getFieldChoice("disks", "disk")
for i, host in ipairs(hosts) do
    setToken("host", host, "host")
    events = { }
    l = getLine(0)
    l = replaceTokens(l)
    table.insert(events, l)

    for i=1,#disks do
        usedPct = round(math.random() + math.random(options["minDiskUsedPct"], options["maxDiskUsedPct"]),2)
        usedGB = round((usedPct/100) * options["totalGBperDisk"])
        availGB = round(options["totalGBperDisk"] - usedGB)
        setToken("usedPct", usedPct)
        setToken("usedGB", usedGB)
        setToken("availGB", availGB)
        setToken("totalGB", options["totalGBperDisk"])
        if availGB < 1 then
            setToken("fs", "/dev/sdb1")
            setToken("mnt", "var")
        else
            setToken("fs", disks[i])
            setToken("mnt", mounts[i])
        end
        l = getLine(1)
        l = replaceTokens(l)
        table.insert(events, l)
    end
    send(events)
end

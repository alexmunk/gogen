hosts = getFieldChoice("host", "host")
for i, host in ipairs(hosts) do
    setToken("host", host, "host")
    events = { }
    l = getLine(0)
    l = replaceTokens(l)
    table.insert(events, l)

    kbps = math.random(options["minKBPS"], options["maxKBPS"])
    packets = kbps / 65.535
    receivedPct = math.random(1000, 4000) / 1000
    for i=1,options["numNICs"] do
        nic = "eth"..tostring(i)
        rx_p = round(receivedPct * packets, 2)
        tx_p = round(packets - rx_p, 2)
        rx_kb = round(receivedPct * kbps, 2)
        tx_kb = round(kbps - rx_kb, 2)

        tokens = { "nic", "rx_p", "tx_p", "rx_kb", "tx_kb" }
        for i, t in ipairs(tokens) do
            setToken(t, _G[t])
        end
        l = getLine(1)
        l = replaceTokens(l)
        table.insert(events, l)
    end
    send(events)
end
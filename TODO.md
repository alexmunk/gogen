# Todo

* Build a Splunk app which proxies stdout to modinput and allows for easily adding any Gogen via a modinput
* Unit test coverage 90%
* Lua generators need to support deconstruction in GitHub pull
* Add caching of GitHub API calls
* Implement checkpointing state
    * Create channels back to each imer thread
    * Outputters should acknowledge output and that should increment state counters
    * Each timer thread should write current state after ack
    * This can also be used for performance counters
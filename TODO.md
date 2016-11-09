# Todo

* Build a Splunk app which proxies stdout to modinput and allows for easily adding any Gogen via a modinput
* Unit test coverage 90%
* Integrate LUA interpret for custom generators and raters
    * Need to remember different replacements across stages in user lua generator
    * Add ability to have state mantained on the generator thread rather than the sample to allow for multi-threading
    * Build example where we model a user with different log messages, pulled from the lines map, played at differing intervals depending on the choices
      an individual user makes.  Model an optional number of users with an optional set of choices for how many intervals they will go through before
      various actions are taken.
* Add caching of GitHub API calls
* Implement checkpointing state
    * Create channels back to each imer thread
    * Outputters should acknowledge output and that should increment state counters
    * Each timer thread should write current state after ack
    * This can also be used for performance counters
# Todo

* Create a concept of Mixes which are configs which only contain links to other samples.
* Transliterate a bunch of TA configs and publish them
* Remove OutputTemplates
* Consider finding a way to break up config package and refactor using better interface design
* Unit test coverage 90%
* Lua generators need to support deconstruction in GitHub pull
* Implement checkpointing state
    * Create channels back to each imer thread
    * Outputters should acknowledge output and that should increment state counters
    * Each timer thread should write current state after ack
    * This can also be used for performance counters
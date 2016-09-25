# Todo

* Implement rating for event counts and rated fields
    * Rating should support lua based raters as well
* Implement header and footer for output templates to allow for CSV & XML generation
* Implement replay generator
* Integrate LUA interpret for custom generators and raters
* Implement catalog of configs, easy upload from command line to repo
* Implement pulling configs via HTTP
* Implement checkpointing state
    * Create channels back to each imer thread
    * Outputters should acknowledge output and that should increment state counters
    * Each timer thread should write current state after ack
    * This can also be used for performance counters
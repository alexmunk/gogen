# Todo

* Unit test coverage 90%
* Implement rating for event counts and rated fields
    * Rating should support lua based raters as well
* Implement replay generator
* Integrate LUA interpret for custom generators and raters
    * Add Sample as reflected global in Lua interpreter
    * Add options map to Sample to pass optional parameters to the generator
    * Build example where we model a user with different log messages, pulled from the lines map, played at differing intervals depending on the choices
      an individual user makes.  Model an optional number of users with an optional set of choices for how many intervals they will go through before
      various actions are taken.
* Add caching of GitHub API calls
* Implement checkpointing state
    * Create channels back to each imer thread
    * Outputters should acknowledge output and that should increment state counters
    * Each timer thread should write current state after ack
    * This can also be used for performance counters


Bugs:

http output not working
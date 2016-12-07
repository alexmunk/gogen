# Configuration Philosophy

Gogen is the spiritual successor to my original [Eventgen](https://github.com/splunk/eventgen), and as such it shares many configuration concepts in common.  However, Gogen as the successor has been designed to work around a number of deficiencies in Eventgen's original configuration format.  

Gogen was designed to be configured from a single file.  This makes moving configurations around very simple.  Managing the data Gogen references is painful from a single file, so it also allows for referencing other files.  For example, choice token types allow choosing items from fields in other samples, and these samples can be referenced to files on the file system.  When publishing the configurations, Gogen will take it's in-memory representation which has joined all the configuration data together, and generate a single file version of this configuration.  Later, if desired, Gogen can deconstruct this single file representation back into component files to make editing easier.

## Config File Overview

Gogen is configured via a YAML or JSON based configuration.  Lets look at a very simple example configuration:

    samples:
      - name: tutorial1
        interval: 1
        endIntervals: 5
        count: 1
        randomizeEvents: true
        
        tokens:
        - name: ts
          format: template                                                                                                                     
          type: timestamp
          replacement: "%b/%d/%y %H:%M:%S"

        lines:
        - _raw: $ts$ line1
        - _raw: $ts$ line2
        - _raw: $ts$ line3

This example is in YAML.  Gogen configurations are made up of Samples, which contain some configuration, tokens, and lines.  In this example, we will generate 1 event (`count: 1`) from a random line (`randomizeEvents: true`) every 1 second (`interval: 1`) for a total of 5 intervals (`endIntervals 5`).  When `endIntervals` is set, we will go back that number of intervals and just work as fast as we can to generate that number of events.  Gogen can also keep generating and generate in realtime, which we'll cover a bit later.  


TODO:

Simple replay
Single JSON Document
Translog
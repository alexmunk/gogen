# Python Eventgen
    python bin/eventgen.py --multiprocess --generators 4 --disableOutputQueue -v tests/perf/eventgen.conf.perfsampleweblog
    2016-09-25 10:51:54,473 INFO module='main' sample='null': GlobalEventsPerSec=6340.8 KilobytesPerSec=2266.661328 GigabytesPerDay=186.767138
    
    python bin/eventgen.py --multiprocess --generators 4 --disableOutputQueue -v tests/perf/eventgen.conf.perfweblog
    2016-09-25 10:50:56,125 INFO module='main' sample='null': GlobalEventsPerSec=18400.0 KilobytesPerSec=6556.238867 GigabytesPerDay=540.217436
 
    python bin/eventgen.py --multiprocess --generators 4 --disableOutputQueue -v tests/perf/eventgen.conf.perfcweblog
    2016-09-25 10:49:17,479 INFO module='main' sample='null': GlobalEventsPerSec=213706.8 KilobytesPerSec=74564.968945 GigabytesPerDay=6143.964116

# Gogen
    gogen -g 6 -o 2 gen -s weblog-regex -ei 10000 -o devnull
    2016-09-26T13:14:30.449-07:00 ROT  Events/Sec: 14000.00 Kilobytes/Sec: 4956.60 GB/Day: 408.41

    gogen -g 6 -o 2 gen -s weblog -ei 10000 -o devnull
    2016-09-26T13:15:30.409-07:00 ROT  Events/Sec: 113000.00 Kilobytes/Sec: 40066.22 GB/Day: 3301.35 

Appears we're about 2-6x faster.
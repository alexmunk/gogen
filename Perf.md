# Python Eventgen
    python bin/eventgen.py --multiprocess --generators 4 --disableOutputQueue -v tests/perf/eventgen.conf.perfsampleweblog
    2016-09-25 10:51:54,473 INFO module='main' sample='null': GlobalEventsPerSec=6340.8 KilobytesPerSec=2266.661328 GigabytesPerDay=186.767138
    
    python bin/eventgen.py --multiprocess --generators 4 --disableOutputQueue -v tests/perf/eventgen.conf.perfweblog
    2016-09-25 10:50:56,125 INFO module='main' sample='null': GlobalEventsPerSec=18400.0 KilobytesPerSec=6556.238867 GigabytesPerDay=540.217436
 
    python bin/eventgen.py --multiprocess --generators 4 --disableOutputQueue -v tests/perf/eventgen.conf.perfcweblog
    2016-09-25 10:49:17,479 INFO module='main' sample='null': GlobalEventsPerSec=213706.8 KilobytesPerSec=74564.968945 GigabytesPerDay=6143.964116

# Gogen
    gogen -g 4 --doq gen -ei 10000 -s weblog-regex | python ~/local/projects/sa-eventgen/bin/eventcount.py
    2016-09-25 10:55:56 Events/Sec: 14054.2 Kilobytes/Sec: 4953.600000 GB/Day: 408.164062

    gogen -g 4 --doq gen -ei 10000 -s weblog | python ~/local/projects/sa-eventgen/bin/eventcount.py
    2016-09-25 10:54:48 Events/Sec: 98857.8 Kilobytes/Sec: 34841.600000 GB/Day: 2870.859375

Appears we're about 4-6x faster.
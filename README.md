# Still to be done

In the absence of a real README, here's a few commands to get you started:

    gogen
    gogen -c examples/weblog/weblog.yml 
    gogen -c examples/csv/csv.yml

To see what we're doing behind the scenes:

    gogen -v gen -s translog -c 1 -ei 1

To see how good your laptop is:

    gogen -v -g 4 -c tests/perf/weblog.yml gen -s weblog -o devnull 
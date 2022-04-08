# stopwatch-go
Collect timestamps of events and report them as CSV. Written in Go.

Useful for manually recording occurrences of various phenomena.  Analyzing the
timestamps may then help user to figure out the cause or source of the
phenomena.

Example: you have a mechanical machine that suddenly starts to make worrying
clicking noises. You start the stopwatch and every time you hear a click,
you immediately press the record key (see *Usage* below).
When you feel confident that you have collected enough information, you stop
the program and import the resulting CSV file in your favourite data analysis
tool (`pandas`, for example). Next, you compute the average frequency of the clicking
and notice that this frequency matches closely the angular frequency an axle in
in your machine. Perhaps its bearing needs replacing? Or a gear has broken off?
You are not sure yet, but you feel reassured that your further diagnostics will
be directed towards the right part or location inside the machine.

**NOTE**: This tool is quite "ad-hoc" in principle; it knows nothing of the
phenomena you are recording. Additionally, the accuracy of the recorded
timestamps depend mostly on the latency of the meatspace operator connected to
the `stdin` file descriptor. If you need to collect information reliably or
accurately, I recommend setting up actual automated electronic sensors and
recording equipment (not covered by this application).

# Install

    $ go install github.com/MawKKe/stopwatch-go@latest

This will place the binary `stopwatch-go` into your `$GOPATH/bin/`.
If that path is in your `$PATH`, you are good to go. Next, see `Usage` below.

# Usage

Run the program; write events into `stdout`:

    $ stopwatch-go

Run the program; write events into file named `foo.csv`:

    $ stopwatch-go -o foo.csv

**NOTE**: in this mode, the previous file will be overwritten. **Be careful.**

When the program is running, you record timestamp of a "events" by
pressing `<enter>`. You can press enter as many times as you like. To stop
the program, press either `<ctrl+d>` or `<ctrl+c>`.

**NOTE**: this program does not analyze the data for you. You must do that
with some other tool.

Example output:

    $ stopwatch-go
    # Record: <enter>, Exit: <ctrl+d> or <ctrl-c>
    >> Waiting... [1]:
    >> Waiting... [2]:
    >> Waiting... [3]:
    >> Waiting... [4]: ^C
    seq,ts,what
    0,2022-04-08T20:12:36.928118021+03:00,enter
    1,2022-04-08T20:12:37.774229977+03:00,tick
    2,2022-04-08T20:12:38.74224978+03:00,tick
    3,2022-04-08T20:12:39.758276309+03:00,tick
    4,2022-04-08T20:12:40.790300244+03:00,exit

Here you may notice that each record is separated by approximately one second,
simulating a phenomena occurring at frequency of 1 Hertz. The recording was
stopped by pressing `<ctrl+c>` while the program was waiting for a fourth event.

# Dependencies

The program is written in Go, version 1.18. It may compile with older compiler versions.
The program does not have any third party dependencies.

# License

Copyright 2022 Markus Holmström (MawKKe)

The works under this repository are licenced under Apache License 2.0.
See file `LICENSE` for more information.

# Contributing

This project is hosted at https://github.com/MawKKe/stopwatch-go

You are welcome to leave bug reports, fixes and feature requests. Thanks!


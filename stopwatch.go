// Copyright 2022 Markus Holmström (MawKKe) markus@mawkke.fi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// stopwatch-go: Collect timestamps of events and report them as CSV
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

// Represents an event to be recorded
type Event struct {
	Seq       int32     `csv:"seq"`  // sequence number of the event
	Timestamp time.Time `csv:"ts"`   // when the event happened
	What      string    `csv:"what"` // description of the event
}

// Convert an Event into a slice of strings. Used for writing Event as CSV record.
func (e Event) Row() []string {
	return []string{fmt.Sprintf("%v", e.Seq), e.Timestamp.Format(time.RFC3339Nano), e.What}
}

// Produce a slice of column names from Event. Used for writing CSV header.
func GetEventColumnNames() []string {
	var hdr []string
	etype := reflect.TypeOf(Event{})
	for i := 0; i < etype.NumField(); i++ {
		field := etype.Field(i)
		hdr = append(hdr, field.Tag.Get("csv"))
	}
	return hdr
}

// Write a slice of Events into output file in CSV mode.
func DumpCSV(out io.Writer, events []Event, comment string) error {
	w := csv.NewWriter(out)
	if comment != "" {
		_, err := fmt.Fprintln(out, fmt.Sprintf("# %v", comment))
		if err != nil {
			return err
		}
	}
	var rows [][]string
	rows = append(rows, GetEventColumnNames())
	for _, evt := range events {
		rows = append(rows, evt.Row())
	}
	err := w.WriteAll(rows)
	if err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func main() {
	outFile := flag.String("o", "-", "Output file path (Optional, default: stdout)")
	outComment := flag.String("c", "", "Comment for the output file. Optional")
	flag.Parse()

	chanSig := make(chan os.Signal, 1)
	chanEvt := make(chan time.Time)

	signal.Notify(chanSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var events []Event
	var ctr int32

	tick := func(what string, t time.Time) {
		events = append(events, Event{Seq: ctr, Timestamp: t, What: what})
		ctr++
	}

	go func() {
		for {
			var s string
			_, err := fmt.Scanln(&s)
			/*
			   pressing only enter will return err == "unexpected newline",
			   but pressing ctrl-d will cause err == io.EOF
			*/
			if err == io.EOF {
				chanSig <- syscall.SIGTERM
				return
			}
			chanEvt <- time.Now()
		}
	}()

	fmt.Fprintln(os.Stderr, "# Record: <enter>, Exit: <ctrl+d> or <ctrl+c>")

	tick("enter", time.Now())

loop:
	for {
		fmt.Fprint(os.Stderr, fmt.Sprintf("# Waiting for [%v]> ", ctr))
		select {
		case <-chanSig: // got signal, exit
			break loop
		case t := <-chanEvt: // got event, record, continue
			tick("tick", t)
			continue
		}

	}

	tick("exit", time.Now())

	// Make sure next print will be on a fresh line
	fmt.Fprintln(os.Stderr, "")

	// Write events into file; either stdout or
	if *outFile == "-" || *outFile == "" {
		err := DumpCSV(os.Stdout, events, *outComment)
		if err != nil {
			panic(err)
		}
	} else {
		f, err := os.Create(*outFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		err = DumpCSV(f, events, *outComment)
		if err != nil {
			panic(err)
		}
	}
}

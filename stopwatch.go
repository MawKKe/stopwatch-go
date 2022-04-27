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
	"context"
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

// Event represents an event to be recorded
type Event struct {
	Seq       int       `csv:"seq"`  // sequence number of the event
	Timestamp time.Time `csv:"ts"`   // when the event happened
	What      string    `csv:"what"` // description of the event
}

// Row converts an Event into a slice of strings. Used for writing Event as CSV record.
func (e Event) Row() []string {
	return []string{fmt.Sprintf("%d", e.Seq), e.Timestamp.Format(time.RFC3339Nano), e.What}
}

// GetEventColumnNames produces a slice of column names from Event. Used for
// writing CSV header.
func GetEventColumnNames() []string {
	var hdr []string
	etype := reflect.TypeOf(Event{})
	for i := 0; i < etype.NumField(); i++ {
		field := etype.Field(i)
		hdr = append(hdr, field.Tag.Get("csv"))
	}
	return hdr
}

// EventsToRecords converts a sequence of events to string representation
func EventsToRecords(events []Event) [][]string {
	var rows [][]string
	rows = append(rows, GetEventColumnNames())
	for _, evt := range events {
		rows = append(rows, evt.Row())
	}
	return rows
}

// DumpCSV writes a sequence of records into output file in CSV mode.
// Filenames "" and "-" are interpreted as stdout. Comment parameter (if non-empty)
// will be written as "# <comment>" on the first line of the file.
func DumpCSV(outFile string, events []Event, comment string) error {
	if outFile == "-" || outFile == "" {
		return MarshallEventsCSV(os.Stdout, events, comment)
	}
	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()
	return MarshallEventsCSV(f, events, comment)
}

func MarshallEventsCSV(out io.Writer, events []Event, comment string) error {

	// convert records to text form
	records := EventsToRecords(events)

	w := csv.NewWriter(out)
	if comment != "" {
		_, err := fmt.Fprintf(out, "# %s\n", comment)
		if err != nil {
			return err
		}
	}
	return w.WriteAll(records)
}

func collect(ctx context.Context, tickChan <-chan struct{}) (events []Event) {
	var ctr int

	// Print all info messages to stderr, as data might be printed to stdout
	fmt.Fprintln(os.Stderr, "# Record: <enter>, Exit: <ctrl+d> or <ctrl+c>")

	tick := func(what string) {
		events = append(events, Event{Seq: ctr, Timestamp: time.Now(), What: what})
		ctr++
	}

	tick("enter")
loop:
	for {
		fmt.Fprintf(os.Stderr, "# Waiting for [%v]> ", ctr)
		select {
		case <-ctx.Done():
			break loop // plain 'break' would break from select, not the loop.
		case <-tickChan:
			tick("tick")
		}
	}
	tick("exit")

	// Make sure next print will be on a fresh line
	fmt.Fprintln(os.Stderr, "")
	return
}

func main() {
	outFile := flag.String("o", "", "Output file path (Optional, default: stdout)\n"+
		"Values \"\" and \"-\" are interpreted as stdout")
	outComment := flag.String("c", "", "Comment for the output file. Optional")
	flag.Parse()

	// capture signals and handle cancellation via Context
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	defer func() {
		cancel()
	}()

	tickChan := make(chan struct{})

	go func() {
		for {
			var s string
			_, err := fmt.Scanln(&s)
			/*
			   pressing only enter will return err == "unexpected newline",
			   but pressing ctrl-d will cause err == io.EOF
			*/
			if err == io.EOF {
				// tell main loop we are done.
				cancel()
				return
			}

			// (new)line received, notify collector
			tickChan <- struct{}{}
		}
	}()

	events := collect(ctx, tickChan)

	// In case we exited loop due to a signal, the stdin goroutine
	// is still running. Here we close stdin manually to signal the
	// goroutine to exit. The goroutine will receive EOF, and call
	// cancel() on the context (again?)
	os.Stdin.Close()

	// Write events into file; either stdout or
	if err := DumpCSV(*outFile, events, *outComment); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: problem writing CSV:", err)
		os.Exit(1)
	}
	os.Exit(0)
}

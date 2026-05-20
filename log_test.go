package main

import (
	"regexp"
	"testing"
)

func TestGetTimeStamp(t *testing.T) {
	ts := getTimeStamp()
	matched, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}$`, ts)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("timestamp %q doesn't match expected format YYYY-MM-DD HH:MM:SS.mmm", ts)
	}
}

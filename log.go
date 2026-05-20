package main

import (
	"fmt"
	"time"
)

func logMessage(log string, args ...any) {
	fmt.Printf("%s - %s\n", getTimeStamp(), fmt.Sprintf(log, args...))
}

func getTimeStamp() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

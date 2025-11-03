package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var (
    debugLog  *log.Logger
    loggerMux sync.Mutex
    loggerOnce sync.Once
)

// setupLogger initializes the package debug logger once. It's safe to call
// multiple times; only the first call performs initialization.
func setupLogger(enableDebug bool) {
    loggerOnce.Do(func() {
        loggerMux.Lock()
        defer loggerMux.Unlock()

        if enableDebug {
            // Open debug.log file
            logFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
            if err != nil {
                fmt.Printf("Error opening debug.log: %v\n", err)
                // fallback to discard logger
                debugLog = log.New(io.Discard, "", 0)
                return
            }
            debugLog = log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)
        } else {
            // If debug is not enabled, set debugLog to a no-op logger
            debugLog = log.New(io.Discard, "", 0)
        }
    })
}

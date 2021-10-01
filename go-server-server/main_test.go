package main

import (
  "testing"
  "log"
)

// Test started when the test binary is started. Only calls main.
func TestSystem(t *testing.T) {
	log.Printf("info: Starting Tests...")
    main()
	log.Printf("info: Ending Tests...")
}

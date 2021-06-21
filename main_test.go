package main

import (
	"fmt"
	"os"
	"testing"
)

func TestReadConfig(t *testing.T) {

	// set test environment
	if err := os.Setenv("ENV", "test"); err != nil {
		t.Error(err)
	}

	p := Package{}
	fmt.Println(os.Getenv("ENV"))
	if err := p.readConfig(); err != nil {
		t.Errorf("%v", err)
	}

}

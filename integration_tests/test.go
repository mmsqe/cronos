package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	i := 0
	for {
		cmd := exec.Command(
			"pytest", "-v", "-s",
			// "test_basic.py::test_tx_inclusion",
			// "test_filters.py::test_pending_transaction_filter",
			"test_mempool.py",
		)
		var stdBuffer bytes.Buffer
		mw := io.MultiWriter(os.Stdout, &stdBuffer)
		cmd.Stdout = mw
		cmd.Stderr = mw
		if err := cmd.Run(); err != nil {
			log.Panic(err)
		}
		if i += 1; i%3 == 0 {
			exec.Command("rm", "-rf", "/private/tmp/pytest-of-mavistan").Run()
		}
	}
}

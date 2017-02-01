package main

import (
	"fmt"
	"os/exec"

	"github.com/appscode/log"
)

func runCmd(path string) error {
	log.Infoln("calling boot file to execute")
	output, err := exec.Command("sh", "-c", path).CombinedOutput()
	msg := fmt.Sprintf("%v", string(output))
	log.Infoln("Output:\n", msg)
	if err != nil {
		log.Errorln("failed to run cmd")
		return fmt.Errorf("error restarting %v: %v", msg, err)
	}
	log.Infoln("boot file executed")
	return nil
}

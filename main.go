package main

import (
	"fmt"
	"os"
	"runtime/trace"
	"time"

	"github.com/getsavvyinc/savvy-cli/cmd"
)

var cf, pf, tf *os.File

func init() {
	now := time.Now().UnixMilli()
	n := fmt.Sprintf("%d", now)
	var err error

	pf, err = os.OpenFile("/Users/shantanu/.savvy_cpu_pprof"+n+".prof", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}

	tf, err = os.OpenFile("/Users/shantanu/.savvy_trace_pprof"+n+".out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
		// return nil, err
	}

	//	if err := pprof.StartCPUProfile(pf); err != nil {
	//		panic(err)
	//	}
	//
	//	if err := trace.Start(tf); err != nil {
	//		panic(err)
	//	}
}

func main() {
	defer tf.Close()
	defer pf.Close()

	defer trace.Stop()

	fmt.Println("#Starting the CLI")
	cmd.Execute()
	// pprof.StopCPUProfile()
	// pf.Close()
}

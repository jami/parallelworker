package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	numCurrentWorker int
	workerExec       []string
	workerArgList    string
	outputFile       string
)

type logOutput struct {
	id       int
	start    time.Time
	duration time.Duration
	cmd      string
	err      string
	stderr   string
	stdout   string
	exit     int
}

func writeOutput(logWriter <-chan logOutput, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	var (
		of  *os.File
		err error
	)

	if len(outputFile) > 0 {
		of, err = os.Create(outputFile)
		if err != nil {
			fmt.Println("Unable to open output file.", err)
		} else {
			defer of.Close()
		}
	}

	for l := range logWriter {
		buffer := fmt.Sprintf("Worker id: %d\n", l.id)
		buffer = buffer + fmt.Sprintf("Command: %s\n", l.cmd)
		buffer = buffer + fmt.Sprintf("Start at: %s\n", l.start.String())
		buffer = buffer + fmt.Sprintf("Duration: %s\n", l.duration.String())
		buffer = buffer + fmt.Sprintf("Exit code: %d\n", l.exit)
		buffer = buffer + fmt.Sprintf("Error: %s\n", l.err)
		buffer = buffer + fmt.Sprintf("StdOut: %s\n", strings.TrimSpace(l.stdout))
		buffer = buffer + fmt.Sprintf("StdErr: %s\n", strings.TrimSpace(l.stderr))

		if of != nil {
			fmt.Fprintln(of, buffer)
		} else {
			fmt.Println(buffer)
		}
	}
}

func readArgumentList(fp string) ([]string, error) {
	res := []string{}

	file, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}

	return res, nil
}

func runWorker(id int, jobs <-chan string, log chan logOutput, execCmd []string, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for job := range jobs {
		var stderr, stdout bytes.Buffer
		t0 := time.Now()

		prog := execCmd[0]
		args := strings.Join(execCmd[1:], " ")
		args = strings.Replace(args, "%arg", job, -1)

		cmdExec := exec.Command(prog, args)
		cmdExec.Env = os.Environ()
		cmdExec.Stderr = &stderr
		cmdExec.Stdout = &stdout

		err := cmdExec.Run()
		errStr := ""
		exitCode := 0
		if err != nil {
			errStr = err.Error()
			if exiterr, ok := err.(*exec.ExitError); ok {
				waitStatus := exiterr.Sys().(syscall.WaitStatus)
				exitCode = waitStatus.ExitStatus()
			}
		}

		t1 := time.Now()
		log <- logOutput{
			id:       id,
			start:    t0,
			cmd:      prog + " " + args,
			err:      errStr,
			exit:     exitCode,
			duration: t1.Sub(t0),
			stderr:   stderr.String(),
			stdout:   stdout.String(),
		}
	}
}

func init() {
	flag.IntVar(&numCurrentWorker, "numworker", 1, "Number of concurrent worker")
	flag.StringVar(&workerArgList, "args", "", "Argument list file")
	flag.StringVar(&outputFile, "output", "", "Output file")

	flag.Parse()

	if len(workerArgList) == 0 {
		fmt.Println("Missing argument list")
		os.Exit(1)
	}

	workerExec = flag.Args()
	if len(workerExec) == 0 {
		fmt.Println("Missing worker exec command")
		os.Exit(1)
	}
}

func main() {
	fmt.Println("Parallel worker execution")
	fmt.Printf("Number of workers %d\n", numCurrentWorker)
	fmt.Printf("Worker exec command %s\n", workerExec)
	fmt.Printf("Worker arg list %s\n", workerArgList)

	var wg sync.WaitGroup
	var wgLog sync.WaitGroup

	var argList []string
	var err error
	argChan := make(chan string)
	logChan := make(chan logOutput)

	go writeOutput(logChan, &wgLog)

	if argList, err = readArgumentList(workerArgList); err != nil {
		fmt.Println("Invalid argument list")
		os.Exit(1)
	}

	for i := 0; i < numCurrentWorker; i++ {
		go runWorker(i, argChan, logChan, workerExec, &wg)
	}

	for _, arg := range argList {
		argChan <- arg
	}

	close(argChan)
	wg.Wait()

	close(logChan)
	wgLog.Wait()
}

# Parallel worker

Parallel worker is a tool to run batch jobs concurrently. It is very similar to GNU parallel but with less functionallity and less complex. You can just specify a command and a argument list. So the commands got an argument injected and run on a single worker.

## Build

    make

or 

    make build

## Examples

Create a argument list e.g.

    for i in {1..50}; do echo $i; done > arglist

Use php to "process" each argument in pool with 10 worker

    ./bin/parallelworker --numworker 10 --args arglist -- php -r "echo '%arg';"

## Output

Each job produces an output block with some information

    Worker id: <workerId>
    Command: php -r echo '50';
    Start at: 2018-04-10 17:58:58.538964901 +0200 CEST m=+0.432039113
    Duration: 49.578596ms
    Exit code: 0
    Error:
    StdOut: 50
    StdErr:
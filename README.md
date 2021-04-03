# Cynic ![build](https://github.com/psyomn/cynic/workflows/build/badge.svg)

Simple monitoring, contract and heuristic tool. Dependency free!

## TODO

I plan to move this to my big monorepo, with the rest of my projects
in there called `github.com/psyomn/ecophagy`, and eventually archive
this.

## Usage

For detailed usage take a look at `cynic/cynic.go`.

For usage of the storage dumper look at `cynic-store/main.go`.

## Examples

I want to:

- Run an event in 10 seconds: [examples/ten_sec.go][1]
- Run an event every 10 seconds:  [examples/every\_ten\_sec.go][2]
- Run an event immediately, and every 10 seconds:
  [examples/imm\_ten\_sec.go][3]
- Run an event every second, and if timestamps are odd, issue alert:
  [examples/alert.go][4]
- Run an event and access query results in an http endpoint:
  [examples/status_cache.go][5]
- Run an event every 10 seconds, store in http endpoint, and take
  snapshots every one minute, and write it to disk every 2 minutes:
  [examples/snapshot.go][6]

The above should give you enough context to figure out how to do more
complex things, by combining a number of configurations (as shown
above).

To build, simply run: `make examples`

[1]: examples/ten_sec.go
[2]: examples/every_ten_sec.go
[3]: examples/imm_ten_sec.go
[4]: examples/alert.go
[5]: examples/status_cache.go
[6]: examples/snapshot.go

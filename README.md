# Cynic [![Build Status](https://travis-ci.org/psyomn/cynic.svg?branch=master)](https://travis-ci.org/psyomn/cynic)

Simple monitoring, contract and heuristic tool.

## Usage

Rather simple to get it up and running. You first need some
configuration file like this (cynic.config):

```json
[
    {
	"url": "http://localhost:9001/one",
	"secs": 10,
		"contracts": [
			"$.Name"
		]
	},
    {
	"url": "http://localhost:9001/two",
	"secs": 20,
		"contracts": [
			"$.Name",
			"$.Age"
		]
    },
    {
	"url": "http://localhost:9001/flappyerror",
	"secs": 3,
		"contracts": [
			"$.Name",
			"$.Age"
		]
    }
]
```

Then you just need to start a service issuing this comand:

```bash
cynic \
	--config="./test/fixtures/example_cynic_config.json" \
	--log="/some/log/file.log"
```

It is also possible to make cynic post in services like Slack. You
need a webhook to make it post things.

```bash
cynic \
	--config="./test/fixtures/example_cynic_config.json" \
	--slack-hook="http://secret-slack-hook-here.com/hook"
```

Once `cynic` has ran for a while, you can access a http service to find
out about the values it has observed. Notice that these are the values
as required by the jsonpath contracts previously specified:

```bash
$ curl -X GET '0:9999/status'
{"http://localhost:9001/flappyerror":{"$.Age":[12,13],"$.Name":["jon","mary"]},
 "http://localhost:9001/one":{"$.Name":"jon"},
 "http://localhost:9001/two":{"$.Age":[12,13],"$.Name":["jon","mary"]}}
```

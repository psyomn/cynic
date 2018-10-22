# Cynic [![Build Status](https://travis-ci.org/psyomn/cynic.svg?branch=master)](https://travis-ci.org/psyomn/cynic)

Simple monitoring, contract and heuristic tool. Dependency free!

## Usage

For detailed usage take a look at `cynic/cynic.go`.

## Specs

Cynic is designed as a service that can monitor over REST endpoints,
exposing different data.

- A minimum time interval that you can have is 1 second
- A maximum time interval for timer ticks is 1 year (365 days -- not
  caring about leap years).

## Notes

This section will go away when ready to merge the changes. If you see
this, I did a booboo and this shouldn't have been here.

- Timers repeat: timers repeat after `n` time units.
    - How do we represent this in a circular buffer? For example,
      service repeats every 7 seconds; how do we manage this in a 60
      second bufffer? (either dynamically inserting and removing
      timers per tick, or badly distribute 'best-possible'

    - Re-insert timer on expiry, with time setting of now +
      time_interval.

- Timers of 1 second: if not done smart, we'd be inserting a timer
  each second slot, each expiry (which is not great). Might be worth
  hacking in some support for very short timers in this respect (known
  1second timers all get lumped in a 'standard' buffer which gets
  executed every second). On the other hand, might be too paranoid.

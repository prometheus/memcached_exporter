## 0.8.0 / 2020-12-04

* [FEATURE] Support MySQL's InnoDB memcached plugin (by handling their multi-word stats settings values)
* [FEATURE] Make exporter logic available as standalone library package #97
* [ENHANCEMENT] Add --version flag and version metric #99
* [ENHANCEMENT] Update prometheus client library

## 0.7.0 / 2020-07-24

* [CHANGE] Switch logging to go-kit #73
* [CHANGE] Register `memcached_lru_crawler_starts_total` metric correctly (formerly `namespace_lru_crawler_starts`) #83
* [ENHANCEMENT] Add `memcached_time_seconds` metric #74
* [ENHANCEMENT] Add slab metrics related to hot/warm/cold/temp LRUs #76
* [BUGFIX] Fix `memcached_slab_mem_requested_bytes` metric in newer memcached versions #70

## 0.6.0 / 2019-08-25

* [CHANGE] Handle non-existent metrics without NaN values #53
* [ENHANCEMENT] Do not run as root by default in Docker #54
* [ENHANCEMENT] Update prometheus client library

## 0.5.0 / 2018-10-17

* [FEATURE] Add memcached_connections_yielded_total metric #35
* [FEATURE] Add memcached_connections_listener_disabled_total metric #36
* [ENHANCEMENT] Update prometheus client library removing outdated metrics #31

## 0.4.1 / 2018-02-01

* [BUGFIX] Handle connection errors gracefully in all cases

## 0.4.0 / 2018-01-23

* [CHANGE] Use the standard prometheus log library
* [CHANGE] Use the standard prometheus flag library

## 0.3.0 / 2016-10-15

* [CHANGE] Tarball includes a directory
* [CHANGE] Use unique default port 9150
* [FEATURE] Use common build system
* [FEATURE] Export extended slab metrics. Thanks @ipstatic
* [FEATURE] Add -version flag and use common version format
* [FEATURE] Add memcached_max_connections metric

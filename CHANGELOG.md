## 0.15.4 / 2025-10-14

* [CHANGE] Update dependencies

## 0.15.3 / 2025-05-28

* [CHANGE] Update dependencies

## 0.15.2 / 2025-03-21

* [CHANGE] Update dependencies

This addresses CVE-2025-22870

## 0.15.1 / 2025-02-24

* [CHANGE] Update dependencies

This addresses CVE-2024-45337 and CVE-2024-45338

## 0.15.0 / 2024-11-08

* [CHANGE] Update dependencies
* [ENHANCEMENT] Add metric for `direct_reclaims` #227

## 0.14.4 / 2024-06-24

* [CHANGE] Update dependencies

This addresses CVE-2023-45288

## 0.14.3 / 2024-03-22

* [CHANGE] Update dependencies

This addresses CVE-2024-24786 which is not exploitable in the exporter, but set off security scanners.

## 0.14.2 / 2023-12-22

* [CHANGE] Update dependencies

This addresses CVE-2023-48795 which is not exploitable in the exporter, but set off security scanners.

## 0.14.1 / 2023-12-06

* [CHANGE] Build with Go 1.21 #190
* [BUGFIX] Add missing `_total` suffix for metrics for failure to store items #191

## 0.14.0 / 2023-12-06

* [FEATURE] Add metrics for failure to store items #184

## 0.13.1 / 2023-12-06

* [CHANGE] Update dependencies

This addresses CVE-2023-3978 which is not exploitable in the exporter, but set off security scanners.

## 0.13.0 / 2023-06-02

* [FEATURE] Multi-target scrape support #143, #173

## 0.12.0 / 2023-06-02

* [ENHANCEMENT] Add `memcached_extstore_io_queue_depth` #169
* [BUGFIX] Fix exposing `memcached_extstore_pages_free` #169

## 0.11.3 / 2023-04-12

* [ENHANCEMENT] Better error messaging when TLS server name is required #162
* [CHANGE] Update dependencies & build with Go 1.20 to avoid upstream CVEs #166

## 0.11.2 / 2023-03-08

* [BUGFIX] Fix connections via UNIX domain socket #157
* [CHANGE] Update dependencies, including exporter toolkit #161

## 0.11.1 / 2023-02-13

* [FEATURE] Add metric to indicate if memcached is accepting connections #137
* [FEATURE] Support TLS for connection to memcached #153
* [FEATURE] Support systemd socket activation #147
* [ENHANCEMENT] Miscellaneous dependency updates #151 #147 #146 #140

Release 0.11.0 failed due to CI issues.

## 0.10.0 / 2022-06-21

* [FEATURE] Add rusage and rejected_connection metrics #109
* [FEATURE] Add extstore metrics #117

## 0.9.0 / 2021-03-25

* [FEATURE] Add TLS and basic authentication #101

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

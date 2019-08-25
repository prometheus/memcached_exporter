## v0.6.0 / 2019-08-25

* [CHANGE] Handle non-existent metrics without NaN values #53
* [ENHANCEMENT] Do not run as root by default in Docker #54
* [ENHANCEMENT] Update prometheus client library

## v0.5.0 / 2018-10-17

* [FEATURE] Add memcached_connections_yielded_total metric #35
* [FEATURE] Add memcached_connections_listener_disabled_total metric #36
* [ENHANCEMENT] Update prometheus client library removing outdated metrics #31

## v0.4.1 / 2018-02-01

* [BUGFIX] Handle connection errors gracefully in all cases

## v0.4.0 / 2018-01-23

* [CHANGE] Use the standard prometheus log library
* [CHANGE] Use the standard prometheus flag library

## v0.3.0 / 2016-10-15

* [CHANGE] Tarball includes a directory
* [CHANGE] Use unique default port 9150
* [FEATURE] Use common build system
* [FEATURE] Export extended slab metrics. Thanks @ipstatic
* [FEATURE] Add -version flag and use common version format
* [FEATURE] Add memcached_max_connections metric

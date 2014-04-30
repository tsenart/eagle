# Eagle

A black-box HTTP testing framework.

  * reports metrics using [Prometheus][1]
  * uses Service Discovery to find endpoints to test
  * provides a server stub for comparision of routing layer metrics

![eagle](https://cloud.githubusercontent.com/assets/3432/2821618/3730c7b4-cf08-11e3-860c-854e153b7e6e.jpg)

## Usage

```
Usage of ./eagle:
  -listen=":7800": Server listen address.
  -test.name="unknown": Name of the test to run.
  -test.path="/": Path to hit on the targets
  -test.rate=100: Number of requests to send during test duration.
  -test.target=: Target to hit by the test with the following format: -test.target="NAME address/url"
```

The `-test.target` flag can be repeated to provide a set of targets to load test.

## Todo

  * API for load tests
  * vary on request/response body sizes

## Author

SoundCloud, Tobias Schmidt, Alexander Simmerl

[1]: https://github.com/prometheus/prometheus

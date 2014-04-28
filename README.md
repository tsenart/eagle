# Eagle

A black-box HTTP testing framework.

  * reports metrics using [Prometheus][1]
  * uses Service Discovery to find endpoints to test
  * provides a server stub for comparision of routing layer metrics

![eagle](https://cloud.githubusercontent.com/assets/3432/2821618/3730c7b4-cf08-11e3-860c-854e153b7e6e.jpg)

## Usage

Create a config file and start the server using the `-config` flag. Example:

```toml
name = "loadtest"

[tests.direct]
address = "http.web.prod.liebling.srv"

[tests.loadbalancer]
url = "http://liebling"
```

## Todo

  * API for load tests
  * vary on request/response body sizes

## Author

SoundCloud, Tobias Schmidt

[1]: https://github.com/prometheus/prometheus

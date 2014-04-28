# Eagle

A black-box HTTP testing framework. Inspired by [SoundCloud Canary][1].

  * reports metrics using [Prometheus][2]
  * uses [Service Discovery][3] to find endpoints to test
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

## Maintainer

ISS <[iss@soundcloud.com](mailto:iss@soundcloud.com)>

[1]: https://github.com/soundcloud/canary
[2]: https://github.com/prometheus/prometheus
[3]: http://go/service-discovery

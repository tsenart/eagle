# Eagle

A black-box HTTP testing framework. Inspired by [SoundCloud Canary][1].

  * reports metrics using [Prometheus][2]
  * uses [Service Discovery][3] to find endpoints to test
  * provides a server stub for comparision of routing layer metrics

## Usage

Configuration example:

```toml
[eagle-squirrel]
type = permanent
address = http.squirrel.prod.eagle.

[liebling-web]
type = test
address = http.web.prod.liebling.
```

## API

### api.eagle/<type>/<name>?<params>

Creates/updates environment and starts new instances with that env.

## Todo

  * vegeta library integration to make http calls
  * prometheus library integration to export results
  * service discovery integration to detect endpoints automatically
  * API / load tests
  * vary on request/response body sizes

## Maintainer

ISS <[iss@soundcloud.com](mailto:iss@soundcloud.com)>

[1]: https://github.com/soundcloud/canary
[2]: https://github.com/prometheus/prometheus
[3]: http://go/service-discovery

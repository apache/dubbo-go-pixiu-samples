# Traffic Filter

`TRAFFIC_V1_ADDR`, `TRAFFIC_V2_ADDR`, and `TRAFFIC_V3_ADDR` configure the
backend listeners. `TRAFFIC_GATEWAY_URL` configures the gateway URL used by
tests and the listener port rendered into Pixiu's configuration. Omitted
variables use `:1315`, `:1316`, `:1317`, and `http://localhost:8888`.

Backend values must be `host:port` addresses. A wildcard listen host such as
`:1315` or `0.0.0.0:1315` is converted to `127.0.0.1:1315` for Pixiu's
same-machine upstream. The gateway value must be an `http://host:port` URL
without a path, query, or fragment. Invalid values fail before startup. The
checked-in YAML files remain directly runnable with defaults; non-default
YAML is generated only into the integration build directory (or the explicit
output path below), never back into `pixiu/`.

## Default ports

From the samples repository root, start the backends:

```sh
go run ./dubbogo/simple/traffic/server/app
```

From the Pixiu repository root, start the canary gateway:

```sh
go run ./cmd/pixiu gateway start -c "$PWD/../dubbo-go-pixiu-samples/dubbogo/simple/traffic/pixiu/conf.yaml"
```

To run the canary tests from the samples repository root:

```sh
go test -count=1 -tags integration -run '^TestCanary' ./dubbogo/simple/traffic/test/...
```

For header routing, stop the first gateway and run this command from the
Pixiu repository root:

```sh
go run ./cmd/pixiu gateway start -c "$PWD/../dubbo-go-pixiu-samples/dubbogo/simple/traffic/pixiu/header-conf.yaml"
```

Then run the header tests from the samples repository root:

```sh
go test -count=1 -tags 'integration manual' -run '^TestHeader' ./dubbogo/simple/traffic/test/...
```

## Non-default ports

This example uses backend ports `11315`/`11316`/`11317` and gateway port
`18888`. The same four values must be present in the backend, config renderer,
gateway client, and tests. From the samples repository root, render a canary
config into a temporary build directory:

```sh
export TRAFFIC_V1_ADDR=127.0.0.1:11315
export TRAFFIC_V2_ADDR=127.0.0.1:11316
export TRAFFIC_V3_ADDR=127.0.0.1:11317
export TRAFFIC_GATEWAY_URL=http://127.0.0.1:18888
mkdir -p /tmp/pixiu-traffic-build
go run ./dubbogo/simple/traffic/configure \
  -input ./dubbogo/simple/traffic/pixiu/conf.yaml \
  -output /tmp/pixiu-traffic-build/conf.yaml
```

In a second terminal at the samples repository root, start the backends with
the same settings:

```sh
TRAFFIC_V1_ADDR=127.0.0.1:11315 \
TRAFFIC_V2_ADDR=127.0.0.1:11316 \
TRAFFIC_V3_ADDR=127.0.0.1:11317 \
TRAFFIC_GATEWAY_URL=http://127.0.0.1:18888 \
go run ./dubbogo/simple/traffic/server/app
```

From the Pixiu repository root, start the gateway with the generated config:

```sh
go run ./cmd/pixiu gateway start -c /tmp/pixiu-traffic-build/conf.yaml
```

In another samples-root terminal, run the canary tests with those same
settings:

```sh
TRAFFIC_V1_ADDR=127.0.0.1:11315 \
TRAFFIC_V2_ADDR=127.0.0.1:11316 \
TRAFFIC_V3_ADDR=127.0.0.1:11317 \
TRAFFIC_GATEWAY_URL=http://127.0.0.1:18888 \
go test -count=1 -tags integration -run '^TestCanary' ./dubbogo/simple/traffic/test/...
```

For header routing, stop the gateway, then run the following in the
samples-root terminal where the four variables were exported:

```sh
go run ./dubbogo/simple/traffic/configure \
  -input ./dubbogo/simple/traffic/pixiu/header-conf.yaml \
  -output /tmp/pixiu-traffic-build/header-conf.yaml
```

From the Pixiu repository root, start the header gateway:

```sh
go run ./cmd/pixiu gateway start -c /tmp/pixiu-traffic-build/header-conf.yaml
```

From the samples repository root, run the header tests:

```sh
TRAFFIC_V1_ADDR=127.0.0.1:11315 \
TRAFFIC_V2_ADDR=127.0.0.1:11316 \
TRAFFIC_V3_ADDR=127.0.0.1:11317 \
TRAFFIC_GATEWAY_URL=http://127.0.0.1:18888 \
go test -count=1 -tags 'integration manual' -run '^TestHeader' ./dubbogo/simple/traffic/test/...
```

## Integration runner

The runner builds and configures Pixiu from the checkout passed through
`PIXIU_DIR`, starts only processes belonging to that run, and holds the shared
traffic integration lock through cleanup. From the samples repository root:

```sh
PIXIU_DIR=/path/to/dubbo-go-pixiu ./integrate_test.sh dubbogo/simple/traffic
PIXIU_DIR=/path/to/dubbo-go-pixiu \
  PIXIU_CONFIG=header-conf.yaml TEST_TAGS='integration manual' TEST_RUN='^TestHeader' \
  ./integrate_test.sh dubbogo/simple/traffic
```

Each runner invocation generates the selected configuration under its
temporary build directory. `start_integrate_test.sh` runs both canary and
header routing on the defaults and then repeats both suites on the non-default
ports above. To use the runner for the non-default ports, export the same four
variables from the non-default example before invoking these two commands.

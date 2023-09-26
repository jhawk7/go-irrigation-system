# go-irrigation-system
Plant Irrigation System built with Raspberry Pi using Go

# Use Buildx
* Set buildx as default in your enviroment to allow docker-compose to use buildx as default
`COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1`
* The platform is set to arm/v6 in the docker-compose file

# Environment variables
* `environment` - environment for telemetry purposes (defaults to `development`)
* `SERVICE_NAME` - service identifier for telemetry (defaults to `default_application`)
* `OTEL_EXPORTER_OTLP_ENDPOINT` - IP:PORT for running opentelemetry collector instance (defaults to  `0.0.0.0:4317`)
* `CHECK_DELAY_SECONDS` - time in seconds to delay moisture reads from ADC moisture sensor
* `CALLBACK_DELAY_SECONDS` - time in seconds to delay continous Opentel Gauge Observer function callbacks
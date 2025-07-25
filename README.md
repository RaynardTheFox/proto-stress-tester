# Protobuf Stress Test Tool

## Features

- Detailed metrics and statistics
- Dynamic load pattern control
- Support for multiple endpoints
- Template-based request generation
- FastHTTP
- Configurable via YAML/JSON/TOML
- Real-time metrics

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/protobuf-stress-test.git
cd protobuf-stress-test

# Install dependencies
go mod download

# Build the binary
go build -o stress-test ./cmd/stress-test
```

## Quick Start

1. Create a configuration file (e.g., `config.yaml`):

```yaml
duration: 5m
maxRPS: 1000
endpoints:
  - name: "example"
    url: "http://localhost:8080/api"
    method: "POST"
    headers:
      Content-Type: "application/json"
    body:
      message: "Hello, World!"
```

2. Run the stress test:

```bash
./stress-test -config config.yaml -workers 100 -duration 5m
```

## Configuration

The tool supports various configuration options through a config file:

### Basic Configuration

```yaml
duration: 5m          # Test duration
maxRPS: 1000         # Maximum requests per second
workers: 100         # Number of concurrent workers
```

### Load Pattern

```yaml
loadPattern:
  type: "ramp-up"    # Load pattern type
  startRPS: 100      # Initial RPS
  increment: 50      # RPS increment
  interval: 30s      # Interval between increments
```

### Endpoints

```yaml
endpoints:
  - name: "endpoint1"
    url: "http://api.example.com/v1/resource"
    method: "POST"
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer ${token}"
    queryParams:
      param1: "value1"
    body:
      field1: "value1"
      field2: "${dynamic_value}"
```

## Metrics

The tool provides detailed metrics including:

- Total Requests
- Successful/Failed Requests
- Current RPS
- Latency Statistics (Min, Max, Mean, P50, P95, P99)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [FastHTTP](https://github.com/valyala/fasthttp) for high-performance HTTP client
- [Viper](https://github.com/spf13/viper) for configuration management
- [Protocol Buffers](https://developers.google.com/protocol-buffers) for data serialization

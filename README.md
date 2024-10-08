# Go Prometheus Demo

This project demonstrates a simple Go application with Prometheus metrics, all orchestrated
using Docker Compose. The application generates its own load and reports metrics, which are
then collected by Prometheus and sent to a Oodle remote write endpoint.

## Prerequisites

- Docker
- Docker Compose
- The application requires 6767 and 9797 ports to be free. You can change
  the ports by updating `.env` file

## Project Structure

- `app.go`: The main Go application that generates metrics.
- `Dockerfile`: Used to build the Go application container.
- `docker-compose.yml`: Defines and configures the services.
- `prometheus.yml`: Configuration file for Prometheus.
- `prometheus-entrypoint.sh`: Script to process the Prometheus configuration file.

## Setup

1. Clone this repository:
   ```
   git clone https://github.com/oodle-ai/prom-go-demo.git
   cd prom-go-demo
   ```

2. Create a `.env` file in the `prom-go-demo` directory with the following content
   by replacing placeholders with your account-specific details:
   ```
   X_API_KEY=<API_KEY>
   REMOTE_WRITE_URL=https://<OODLE_COLLECTOR_ENDPOINT>/v1/prometheus/<OODLE_INSTANCE>/write
   ```

## Running the Application

1. Start the services:
   ```
   docker-compose up --build
   ```

   This command will build the Go application and start all services defined in the `docker-compose.yml` file.

2. The services will be available at the following addresses:
   - Go Application: http://localhost:6767
   - Prometheus: http://localhost:9797

3. On successful launch, metrics will be available for consumption in your Oodle UI. 

## Stopping the Application

To stop the application and remove the containers, use:

```
docker-compose down
```

## Troubleshooting

If you encounter any issues:

1. Ensure all required ports (6767, 9797) are free on your host machine.
   If you want to change the ports to be used, you can update `.env` file
   to specify `APP_PORT` and `PROMETHEUS_PORT`.
2. Check the Docker logs for any error messages:
   ```
   docker-compose logs
   ```
3. Verify that your API key is correctly set in the `.env` file.
4. Make sure the remote write endpoint is accessible and correctly configured.

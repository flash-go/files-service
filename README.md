# files-service

Microservice written in Go for managing files, based on the [Flash Framework](https://github.com/flash-go/flash) with Hexagonal Architecture.

## Features

- Support local filesystem store.
- Support HTTP transport.

## Setup

### 1. Install Task

```
go install github.com/go-task/task/v3/cmd/task@latest
```

### 2. Create .env files

```
task env
```

### 3. Setup .env.server

| Environment Variable | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| CONSUL_AGENT         | Full address (host:port) of the Consul agent (e.g., `localhost:8500`).      |
| SERVICE_NAME         | Name used to register the service in Consul.                                |
| SERVICE_HOST         | Host address under which the service is accessible for Consul registration. |
| SERVICE_PORT         | Port number under which the service is accessible for Consul registration.  |
| SERVER_HOST          | Host address the HTTP server should bind to (e.g., `0.0.0.0`).              |
| SERVER_PORT          | Port number the HTTP server should listen on (e.g., `8080`).                |
| LOG_LEVEL            | Logging level. See the log level table for details.                         |

#### Log Levels

| Level    | Value  | Description                                                                            |
|----------|--------|----------------------------------------------------------------------------------------|
| Trace    | -1     | Fine-grained debugging information, typically only enabled in development.             |
| Debug    | 0      | Detailed debugging information helpful during development and debugging.               |
| Info     | 1      | General operational entries about what's going on inside the application.              |
| Warn     | 2      | Indications that something unexpected happened, but the application continues to work. |
| Error    | 3      | Errors that need attention but do not stop the application.                            |
| Fatal    | 4      | Critical errors causing the application to terminate.                                  |
| Panic    | 5      | Severe errors that will cause a panic; useful for debugging crashes.                   |
| NoLevel  | 6      | No level specified; used when level is not explicitly set.                             |
| Disabled | 7      | Logging is turned off entirely.                                                        |

### 4. Setup .env.seed

| Environment Variable    | Description                                                                                 |
|-------------------------|---------------------------------------------------------------------------------------------|
| CONSUL_AGENT            | Full address (host:port) of the Consul agent (e.g., `localhost:8500`).                      |
| SERVICE_NAME            | The name of the service in Consul used to retrieve database connection configuration.       |
| OTEL_COLLECTOR_GRPC     | Address of the OpenTelemetry Collector for exporting traces via gRPC.                       |
| USERS_SERVICE_NAME      | User Management Service Name.                                                               |
| USERS_ADMIN_ROLE        | Administrator Role ID.                                                                      |
| STORE_LOCAL_ROOT_PATH   | Root path of local filesystem for store files.                                              |

### 5. Run seed

```
task seed
```

## Run

```
task
```

### View Swagger docs

```
http://[SERVER_HOST]:[SERVER_PORT]/swagger/index.html
```

## Full list of commands

```
task -l
```

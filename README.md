# Online Song Library

The Online Song Library is a web application designed to manage a collection of songs. It provides a RESTful API that allows users to retrieve, add, update, and delete songs, as well as fetch detailed information about specific songs from an external music info API.

## Table of Contents

- [Tech Stack](#tech-stack)
- [Installation](#installation)
- [Running the Application](#running-the-application)
  - [Using Docker](#using-docker)
  - [Without Docker](#without-docker)
- [API Documentation](#api-documentation)
- [Running Tests](#running-tests)
  - [Unit Tests](#unit-tests)
- [Database Migrations](#database-migrations)
- [Application Configuration](#application-configuration)
- [Contributing](#contributing)
- [License](#license)

## Tech Stack

- Golang
- PostgreSQL
- Docker

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/vadimbarashkov/online-song-library.git
    cd online-song-library
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

## Running the Application

### Using Docker

1. Prepare `.env` file:

    ```bash
    ENV=dev
    MUSIC_INFO_API=https://music.info.api

    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=postgres
    POSTGRES_HOST=db
    POSTGRES_DB=online_song_library
    ```

2. Build and start the services:

    ```bash
    docker-compose up -d --build
    ```

### Without Docker

1. Setup PostgreSQL.

2. Prepare `.env` file:

     ```bash
    ENV=dev
    MUSIC_INFO_API=https://music.info.api

    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=postgres
    POSTGRES_HOST=db
    POSTGRES_DB=online_song_library
    ```

3. Run the application:

    ```bash
    make run/server
    ```

## API Documentation

The application is documented using Swagger. You can explore the API using various tools or access the interactive Swagger UI by running the application and using these links:

- [Swagger UI for dev and test environments](http://localhost:8080/swagger/index.html)
- [Swagger UI for the prod environment](https://localhost:8443/swagger/index.html)

## Running Tests

### Unit Tests

To run unit tests:

```bash
make test/unit
```

## Database Migrations

The application automatically applies migrations from the `/migrations` directory, but you can run them manually using the `Makefile`:

```bash
# Create migration
make migrations/create $(MIGRATION_NAME)

# Run migrations
make migrations/up $(DATABASE_DSN)

# Rollback migrations
make migrations/down $(DATABASE_DSN)
```

## Application Configuration

The application is configured via a `.env` file, but you can specify the configuration path using the `CONFIG_PATH` environment variable.

Here is the basic structure of the configuration file:

```bash
# enum=[dev,test,prod], default=dev
ENV=dev

# default=migrations
MIGRATIONS_PATH=migrations
# required
MUSIC_INFO_API=https://music.info.api

# default=8080
HTTP_SERVER_PORT=8080
# default=5s
READ_TIMEOUT=5s
# default=10s
WRITE_TIMEOUT=10s
# default=1m
IDLE_TIMEOUT=1m
# default=1048576
MAX_HEADER_BYTES=1048576
CERT_FILE=./crts/example.pem
KEY_FILE=./crts/example-key.pem

# required
POSTGRES_USER=postgres
# required
POSTGRES_PASSWORD=postgres
# default=localhost
POSTGRES_HOST=localhost
# default=5432
POSTGRES_PORT=5432
# required
POSTGRES_DB=online_song_library
# default=disable
POSTGRES_SSLMODE=disable
```

The behavior of the application depends on the environment passed in the configuration file:

1. `dev` - http server doesn't use SSL/TLS certificates and logging is structured without JSON.
2. `test` - http server doesn't use SSL/TLS certificates and logging is structured with JSON.
3. `prod` - http server uses SSL/TLS certificates and logging is structured with JSON.

## Contributing

Contributions are welcome! Suggest your ideas in issues or pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

# weather-app

- [weather-app](#weather-app)
  - [API Documentation](#api-documentation)
  - [Database Schema](#database-schema)
    - [Unit Tests](#unit-tests)
  - [Server Configuration](#server-configuration)
  - [Development & Running locally](#development--running-locally)
    - [Install go to run the server locally](#install-go-to-run-the-server-locally)
    - [Running the UI locally](#running-the-ui-locally)
  - [Docker-compose](#docker-compose)
    - [Force re-build](#force-re-build)
    - [Getting logs](#getting-logs)
  - [Docker](#docker)
    - [Building the server image(s)](#building-the-server-images)
      - [Individually running the server image](#individually-running-the-server-image)
    - [Building the ui image](#building-the-ui-image)

## API Documentation

The backend's openapi definition is available in [openapi.yaml](./openapi.yaml). You can load this in [editor.swagger.io](https://editor.swagger.io/) to see it visually.

## Database Schema

While this project does not actually store data in a relational database, a potential schema is provided in [`types.go`](./server/types/types.go) as `WeatherSchema`.


### Unit Tests

```shell
go test ./...
```

## Server Configuration

The server needs to be configured to communicate with the various weather backends using their API keys.

Provide a `config.json` file in the following format:

```json
{
  "backends": {
    "accuweather": {
      "apiKey": "YOUR_API_KEY"
    },
    "openweathermap": {
      "apiKey": "YOUR_API_KEY"
    }
  }
}
```

## Development & Running locally

There are two ways to run the server; you can run it locally or you can run it in docker.


### Install go to run the server locally

- On a mac, `brew install go` is your best bet
- This app assumes go 1.12

You can then run `go run ./server/main.go` or `go build ./server/main.go` and then just run the generated `go-weather-app` executable.


### Running the UI locally

The easiest way to use this locally is to run `npm start` from the `ui` folder.

## Docker-compose

The easiest way to run the application is to just run `docker-compose up -d`. This should take care of building the images as well if they haven't already been built.

### Force re-build

```shell
docker-compose up -d --build
```

### Getting logs

Both images output their logs to standard out. So you can use `docker-compose logs -f` to follow the logs they output.

## Docker

### Building the server image(s)

You can build either the production or the development images.

```shell
docker build -t dotlou/go-weather-app:production . --target production -f server.Dockerfile
```

```shell
docker build -t dotlou/go-weather-app:development . --target=development -f server.Dockerfile
```

#### Individually running the server image

```shell
docker run --rm -it -p 8080:8080 -v $PWD/config.json:/config.json dotlou/go-weather-app:development
```

### Building the ui image

Since the UI requires caddy, we'll just use a the full image. For a true production use case, we should try to remove as much as we can from this image to reduce attack vectors.

```shell
docker build -t dotlou/react-weather-app:development -f ui.Dockerfile .
```

# GSS

GSS (Go serve SPA) is a containerized web server for single-page applications written in Go.

## Features

- Optimized for single-page apps.
- Automatically serves pre-compressed brotli and gzip files if available.
- Sensible default cache configuration.
- Optional out-of-the-box metrics.
- Deployable as a container.
- Lightweight.

## Usage

GSS works as a Docker container. By default it serves a directory in the container named `dist` at port `8080`.

### Running container directly

```sh
docker run -p [container-port]:8080 -v [local-folder-to-serve-path]:/dist lewislbr/gss
```

> Example:
>
> Having a local folder named `public` with SPA files:
>
> ```sh
> docker run -p 3000:8080 -v $PWD/public:/dist lewislbr/gss
> ```
>
> The server with the contents from `public` will be accessible at port `3000`.

### With a custom image

```Dockerfile
FROM lewislbr/gss:latest
COPY [local-folder-to-serve-path] ./dist
```

> Example:
>
> ```Dockerfile
> FROM lewislbr/gss:latest
> COPY /public ./dist
> ```
>
> ```sh
> docker build -t custom-image .
> ```
>
> ```
> docker run -p 3000:8080 custom-image
> ```
>
> The server with the contents from `public` will be accessible at port `3000`.

## Configuration options

Optionally, the server can be configured with a YAML file named `/gss.yaml`.

The configuration file should go into the container, such as:

> ```sh
> docker run -p 3000:8080 -v $PWD/gss.yaml:/gss.yaml -v $PWD/public:/dist lewislbr/gss
> ```

> ```Dockerfile
> FROM lewislbr/gss:latest
> COPY gss.yaml ./
> COPY /public ./dist
> ```

### Files port: `filesPort`

##### string: integer

Configures the port where files will be served. `8080` by default.

> Example:
>
> ```yaml
> # gss.yaml
>
> filesPort: 3000
> ```

### Metrics port: `metricsPort`

##### string: integer

Configures the port where metrics, if enabled, will be available. `8081` by default.

> Example:
>
> ```yaml
> # gss.yaml
>
> metricsPort: 3001
> ```

### Metrics collection: `metrics`

##### string: boolean

Enables metrics collection and exposes an endpoint at `:<metricsPort>/metrics`. Collected metrics include request duration, request status, total requests, and bytes written. False by default.

> Example:
>
> ```yaml
> # gss.yaml
>
> metrics: true
> ```

## Contributing

This project started as a way to learn and to solve a need I had. If you think it can be improved in any way, you are very welcome to contribute!

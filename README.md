# GSS

GSS (Go serve SPA) is a containerized web server for single-page applications written in Go.

## Features

- Optimized for single-page apps.
- Automatically serves pre-compressed brotli and gzip files if available.
- Sensible default cache configuration.
- Docker-based.
- Configurable via YAML.
- Lightweight.

## Usage

GSS works as a Docker image. By default it serves a directory in the container named `dist` at port `80`.

### Running container directly

```sh
docker run -p [container-port]:80 -v [local-folder-to-serve-path]:/dist lewislbr/gss
```

> Example:
>
> Having a local folder named `public` with SPA files:
>
> ```sh
> docker run -p 8080:80 -v $PWD/public:/dist lewislbr/gss
> ```
>
> The server with the contents from `public` will be accessible at port `8080`.

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
> docker run -p 8080:80 custom-image
> ```

## Configuration options

The server can be configured with a YAML file. File must be named `/gss.yaml`.

> Example:
>
> ```sh
> docker run -p 8080:80 -v $PWD/gss.yaml:/gss.yaml -v $PWD/public:/dist lewislbr/gss
> ```

> ```Dockerfile
> FROM lewislbr/gss:latest
> COPY gss.yaml ./
> COPY /public ./dist
> ```

### `headers`

Headers to add to the response.

> Example:
>
> ```yaml
> # gss.yaml
>
> headers:
>   Referrer-Policy: "strict-origin-when-cross-origin"
>   Server: GSS
>   Strict-Transport-Security: "max-age=63072000; includeSubDomains; preload"
> ```

## Contributing

This project started as a way to learn and to solve a need I had. If you think it can be improved in any way, you are very welcome to contribute!

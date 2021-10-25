# GSS

GSS (Go serve SPA) is a containerized web server for single-page applications written in Go.

## Features

- Optimized for single-page apps.
- Automatically serves pre-compressed brotli and gzip files if available.
- Docker-based.
- Configurable via YAML.
- Lightweight.

## Usage

GSS works as a Docker image.

```sh
docker run -p [port-where-to-serve]:80 -v [folder-to-serve]:/dist lewislbr/gss
```

By default it serves a directory in the container named `dist` at port `80`, but you can change this values with a YAML configuration file.

> File must be named `gss.yaml`

> Example:
>
> ```yaml
> directory: public
> port: 8080
> ```
>
> ```sh
> docker run -p 8080:8080 -v $PWD/gss.yaml:/gss.yaml -v $PWD/web/dist:/public lewislbr/gss
> ```
>
> The folder `public` will be accessible at port `8080`.

## Configuration options

### `directory`

Container path to the directory to serve.

Default: `dist`.

> Example:
>
> ```yaml
> directory: public
> ```

### `headers`

Headers to add to the response.

> Example:
>
> ```yaml
> headers:
>   Referrer-Policy: "strict-origin-when-cross-origin"
>   Strict-Transport-Security: "max-age=63072000; includeSubDomains; preload"
> ```

### `port`

Port where to run the server.

Default: `80`.

> Example:
>
> ```yaml
> port: 8080
> ```

## Contributing

This project started as a way to learn and to solve a need I had. If you think it can be improved in any way, you are very welcome to contribute!

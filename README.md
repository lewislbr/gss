# GSS

GSS (Go serve SPA) is a web server for single-page applications written in Go using the standard library.

## Features

- Optimized for single-page apps.
- Automatically serves pre-compressed brotli and gzip files if available.
- Docker-based.
- Configurable via CLI.
- Lightweight.

## Usage

GSS works as a Docker image. By default it serves a directory in the container named `dist` at port `80`, but you can change this values or mount any folder and publish the container at any port.

With the CLI:

```sh
docker run -p [port you want to use]:[container port] -v [absolute path to SPA build folder]:/[container folder] lewislbr/gss [options]
```

> Example:
>
> ```sh
> docker run -p 1234:80 -v $PWD/public:/dist lewislbr/gss
> ```

With a custom image:

```dockerfile
FROM lewislbr/gss:latest
COPY [path to SPA build folder] ./[container folder]
# Optional:
CMD [options]
```

```sh
docker build -t [image-name] .
docker run -p [port you want to use]:[container port] [image-name] [options]
```

> Example:
>
> ```dockerfile
> FROM lewislbr/gss:latest
> COPY public ./dist
> ```
>
> ```sh
> docker build -t test-app .
> docker run -p 1234:80 test-app
> ```

## Configuration

You can configure the server with CLI options defined after the image name when using the Docker CLI, or in a `CMD` statement if using a custom Dockerfile.

### `-d`

Container path to the directory to serve.

Default: `dist`.

> Example:
>
> ```sh
> docker run -p 1234:80 -v $PWD/public:/static-content lewislbr/gss -d static-content
> ```

### `-p`

Port where to run the server.

Default: `80`.

> Please note that this value should be the same as the container port.

> Example:
>
> ```sh
> docker run -p 1234:3000 -v $PWD/public:/dist lewislbr/gss -p 3000
> ```

## Contributing

This project started as a way to learn and to solve a need I had. If you think it can be improved in any way, you are very welcome to contribute!

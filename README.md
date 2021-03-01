# GSS

GSS (Go serve SPA) is a web server for single-page applications written in Go using the standard library.

## Features

- Optimized for single-page apps.
- Automatically serves pre-compressed brotli and gzip files if available.
- Docker-based.
- Configurable via YAML.
- Configurable via CLI.
- Lightweight.

## Usage

GSS works as a Docker image.

```sh
docker run -p [container-port]:80 -v [spa-folder]:/dist lewislbr/gss
```

By default it serves a directory in the container named `dist` at port `80`, but you can change this values via YAML file or CLI flags.

With YAML file (name must be `gss.yaml`):

```yaml
directory: [spa-folder]
port: [server-port]
```

```sh
docker run -p [container-port]:[server-port] -v [yaml-file]:/gss.yaml -v [spa-folder]:/[container-folder] lewislbr/gss
```

> Example:
>
> ```yaml
> directory: public
> port: 7892
> ```
>
> ```sh
> docker run -p 8080:7892 -v $PWD/gss.yaml:/gss.yaml -v $PWD/web/dist:/public lewislbr/gss:test
> ```

With CLI:

```sh
docker run -p [container-port]:[server-port] -v [spa-folder]:/[container-folder] lewislbr/gss [options]
```

> Example:
>
> ```sh
> docker run -p 8080:7891 -v $PWD/public:/dist lewislbr/gss -d public -p 7891
> ```

If both a YAML config and a CLI flag set up a configuration option, the CLI flag prevails.

## Configuration options

### `-d` (CLI), `directory` (YAML)

Container path to the directory to serve.

Default: `dist`.

> Example:
>
> YAML:
>
> ```yaml
> directory: public
> ```
>
> CLI:
>
> ```sh
> docker run -p 8080:80 -v $PWD/public:/public lewislbr/gss -d public
> ```

### `-p` (CLI), `port` (YAML)

Port where to run the server.

Default: `80`.

> Example:
>
> YAML:
>
> ```yaml
> port: 7892
> ```
>
> CLI:
>
> ```sh
> docker run -p 8080:7892 -v $PWD/public:/dist lewislbr/gss -p 7892
> ```

## Contributing

This project started as a way to learn and to solve a need I had. If you think it can be improved in any way, you are very welcome to contribute!

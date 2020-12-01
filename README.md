# GSS

GSS (Go serve SPA) is a web server for single-page applications written in Go using the standard library.

## Usage

GSS works with Docker. By default it serves a directory in the container named `dist` at port `80`, but you can change this values or mount any folder and publish the container at any port.

With the CLI:

```sh
docker run -p [port you want to use]:80 -v [absolute path to your SPA build folder]:/dist lewislbr/gss:latest [flags]
```

> Example:
>
> ```sh
> docker run -p 1234:9000 -v $PWD/public:/dist lewislbr/gss:latest -p 9000
> ```

With a Dockerile:

```dockerfile
FROM lewislbr/gss:latest
COPY [path to your SPA build folder] ./dist

# Optional:
CMD [flags]
```

```sh
docker build -t [your-image-name] .
docker run -p 1234:80 [your-image-name] [flags]
```

> Example:
>
> ```dockerfile
> FROM lewislbr/gss:latest
> COPY public ./dist
> CMD ["-p=9000"]
> ```
>
> ```sh
> docker build -t test-app .
> docker run -p 1234:9000 test-app
> ```

## Configuration

You can change the value of variables via CLI flags.

### `-d`

Container path to the directory to serve.

Default: `dist`.

### `-p`

Port where to run the server.

Default: `80`.

> Please note that this value should be the same as the container port.

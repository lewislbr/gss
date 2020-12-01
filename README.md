# GSS

GSS (Go serve SPA) is a web server for single-page applications written in Go using the standard library.

## Usage

GSS works with Docker. By default it serves a directory named `dist` at port `80`, but you can mount or copy any folder and publish the container at any port.

With only the CLI:

```sh
docker run -p [port you want to use]:80 -v [absolute path to your SPA build folder]:/dist lewislbr/gss:latest

# Example:
docker run -p 1234:80 -v $PWD/public:/dist lewislbr/gss:latest
```

With a Dockerile:

```dockerfile
FROM lewislbr/gss:latest
COPY [path to your SPA build folder] ./dist
```

```sh
docker build -t [your-image-name] .
docker run -p 1234:80 [your-image-name]
```

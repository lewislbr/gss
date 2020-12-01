build:
	@docker build \
	-t lewislbr/gss:$(shell git describe --tags --abbrev=0) \
	-t lewislbr/gss:latest .

run-test:
	@docker run -p 1234:80 -v $$PWD/dist:/dist lewislbr/gss:latest

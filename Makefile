.PHONY: test

build:
	@docker build -t lewislbr/gss:test .

run-default: build
	@docker run --rm -p 8080:8080 -p 9090:9090 -v $$PWD/test/public:/dist lewislbr/gss:test

run-yaml: build
	@docker run --rm -p 8080:8080 -p 9090:9090 -v $$PWD/test/gss.yaml:/gss.yaml -v $$PWD/test/public:/dist lewislbr/gss:test

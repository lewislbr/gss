build:
	@docker build -t lewislbr/gss:test .

run-cli: build
	@docker run --rm -p 8080:8081 -v $$PWD/test/web/dist:/public lewislbr/gss:test -d public -p 8081

run-default: build
	@docker run --rm -p 8080:80 -v $$PWD/test/web/dist:/dist lewislbr/gss:test

run-yaml: build
	@docker run --rm -p 8080:8081 -v $$PWD/test/gss.yaml:/gss.yaml -v $$PWD/test/web/dist:/public lewislbr/gss:test

build:
	@docker build -t lewislbr/gss:test .

run-cli:
	@docker run -p 8080:7891 -v $$PWD/web/dist:/public lewislbr/gss:test -d public -p 7891

run-default:
	@docker run -p 8080:80 -v $$PWD/web/dist:/dist lewislbr/gss:test

run-yaml:
	@docker run -p 8080:7892 -v $$PWD/gss.yaml:/gss.yaml -v $$PWD/web/dist:/public lewislbr/gss:test

.PHONY: test

build:
	@docker build -t lewislbr/gss:test .

run-default: build
	@docker run --rm -p 8080:80 -v $$PWD/test/public:/dist lewislbr/gss:test

run-yaml: build
	@docker run --rm -p 8080:80 -v $$PWD/test/gss.yaml:/gss.yaml -v $$PWD/test/public:/dist lewislbr/gss:test

test:
	@sed -i "" "s|"dist"|"test/public"|g" gss.go
	@go test
	@sed -i "" "s|"test/public"|"dist"|g" gss.go

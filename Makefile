.PHONY: test

build:
	@docker build -t lewislbr/gss:dev .

run: build
	@docker run --rm -p 8080:8080 -p 9090:9090 -v $$PWD/test/public:/dist lewislbr/gss:dev

run-yaml: build
	@docker run --rm -p 8080:8080 -p 9090:9090 -v $$PWD/test/gss.yaml:/gss.yaml -v $$PWD/test/public:/dist lewislbr/gss:dev

test:
	@sed -i "" "s|"dist"|"test/public"|g" gss.go && go test ./... -count=1 -race; sed -i "" "s|"test/public"|"dist"|g" gss.go

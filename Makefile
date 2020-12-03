build-test:
	@docker build -t lewislbr/gss:test .

run-test:
	@docker run -p 1234:80 -v $$PWD/web/dist:/dist lewislbr/gss:test

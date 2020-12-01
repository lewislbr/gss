build:
	@docker build -t gss:latest .

run-test:
	@docker run -p 1234:80 -v $$PWD/dist:/dist gss:latest

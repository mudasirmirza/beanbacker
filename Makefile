beautify:
	@go fmt

build:
	@go build -o bin/beanbacker src/main.go

build_lambda:
	GOOS=linux go build -o bin/beanbackerLambda src/main_lambda.go
	zip bin/beanbackerLambda.zip bin/beanbackerLambda

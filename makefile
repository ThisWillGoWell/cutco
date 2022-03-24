

# update the staging gql lambda without going though the pipeline
update:
	go build -o bootstrap cmd/lambda-api/main.go
	zip lambda.zip bootstrap
	aws lambda update-function-code --zip-file=fileb://lambda.zip --function-name=starket-staging-gql-lambda
	rm bootstrap
	rm lambda.zip
.PHONY: update


dev-logs:


.PHONY: dev-logs
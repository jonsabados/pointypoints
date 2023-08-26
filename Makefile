.DEFAULT_GOAL := build

dist/:
	mkdir dist

frontend/.env.local: frontend/gen_env.sh
	rm -f frontend/.env.local
	cd frontend && ./gen_env.sh

frontend/dist/index.html: $(shell find frontend/src) $(shell find frontend/public) frontend/.env.local
	cd frontend && npm run build

.PHONY: start-dynamo
start-dynamo:
	@docker run -d --name dynamo -p 8000:8000 amazon/dynamodb-local 2>&1 >/dev/null

.PHONY: stop-dynamo
stop-dynamo:
	@docker stop dynamo 2>&1 >/dev/null
	@docker rm dynamo 2>&1 >/dev/null

.PHONY: frontend-test
frontend-test:
	cd frontend && npm run test:unit

.PHONY: backend-test
backend-test:
	@./execute-tests.sh

.PHONY: test
test: frontend-test backend-test

.PHONY: clean
clean:
	rm -rf frontend/dist/ frontend/.env.local
	rm -rf dist

.PHONY: run
run: frontend/.env.local
	cd frontend && npm run serve

dist/corsLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/cors dist/corsLambda.zip

dist/newSessionLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/new dist/newSessionLambda.zip

dist/connectLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/connect dist/connectLambda.zip

dist/disconnectLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/disconnect dist/disconnectLambda.zip

dist/setFacilitatorSessionLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/setfacilitator dist/setFacilitatorSessionLambda.zip

dist/watchSessionLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/watch dist/watchSessionLambda.zip

dist/joinSessionLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/join dist/joinSessionLambda.zip

dist/voteLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/vote dist/voteLambda.zip

dist/updateSessionLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/update dist/updateSessionLambda.zip

dist/clearVotesLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/session/clearvotes dist/clearVotesLambda.zip

dist/pingLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/ping dist/pingLambda.zip

dist/authorizerLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/authorizer dist/authorizerLambda.zip

dist/profileReadLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/profile/read dist/profileReadLambda.zip

dist/profileWriteLambda.zip: dist/ $(shell find . -iname "*.go")
	./scripts/build_lambda.sh github.com/jonsabados/pointypoints/cmd/lambda/profile/write dist/profileWriteLambda.zip

build: frontend/dist/index.html dist/corsLambda.zip dist/newSessionLambda.zip dist/connectLambda.zip \
	dist/disconnectLambda.zip dist/setFacilitatorSessionLambda.zip dist/watchSessionLambda.zip \
	dist/joinSessionLambda.zip dist/voteLambda.zip dist/updateSessionLambda.zip dist/clearVotesLambda.zip \
	dist/pingLambda.zip dist/authorizerLambda.zip dist/profileReadLambda.zip dist/profileWriteLambda.zip
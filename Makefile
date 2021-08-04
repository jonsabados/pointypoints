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

dist/cors: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/cors github.com/jonsabados/pointypoints/cors/lambda

dist/corsLambda.zip: dist/cors
	cd dist && zip corsLambda.zip cors

dist/newSession: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/newSession github.com/jonsabados/pointypoints/session/new

dist/newSessionLambda.zip: dist/newSession
	cd dist && zip newSessionLambda.zip newSession

dist/connect: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/connect github.com/jonsabados/pointypoints/session/connect

dist/connectLambda.zip: dist/connect
	cd dist && zip connectLambda.zip connect

dist/disconnect: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/disconnect github.com/jonsabados/pointypoints/session/disconnect

dist/disconnectLambda.zip: dist/disconnect
	cd dist && zip disconnectLambda.zip disconnect

dist/setFacilitatorSession: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/setFacilitatorSession github.com/jonsabados/pointypoints/session/setfacilitatorsession

dist/setFacilitatorSessionLambda.zip: dist/setFacilitatorSession
	cd dist && zip setFacilitatorSessionLambda.zip setFacilitatorSession

dist/watchSession: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/watchSession github.com/jonsabados/pointypoints/session/watchsession

dist/watchSessionLambda.zip: dist/watchSession
	cd dist && zip watchSessionLambda.zip watchSession

dist/joinSession: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/joinSession github.com/jonsabados/pointypoints/session/joinsession

dist/joinSessionLambda.zip: dist/joinSession
	cd dist && zip joinSessionLambda.zip joinSession

dist/vote: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/vote github.com/jonsabados/pointypoints/session/vote

dist/voteLambda.zip: dist/vote
	cd dist && zip voteLambda.zip vote

dist/showVotes: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/showVotes github.com/jonsabados/pointypoints/session/showvotes

dist/showVotesLambda.zip: dist/showVotes
	cd dist && zip showVotesLambda.zip showVotes

dist/updateSession: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/updateSession github.com/jonsabados/pointypoints/session/update

dist/updateSessionLambda.zip: dist/updateSession
	cd dist && zip updateSessionLambda.zip updateSession

dist/clearVotes: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/clearVotes github.com/jonsabados/pointypoints/session/clearvotes

dist/clearVotesLambda.zip: dist/clearVotes
	cd dist && zip clearVotesLambda.zip clearVotes

dist/ping: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/ping github.com/jonsabados/pointypoints/ping

dist/pingLambda.zip: dist/ping
	cd dist && zip pingLambda.zip ping

dist/authorizer: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/authorizer github.com/jonsabados/pointypoints/profile/authorizer

dist/authorizerLambda.zip: dist/authorizer
	cd dist && zip authorizerLambda.zip authorizer

dist/profileRead: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/profileRead github.com/jonsabados/pointypoints/profile/read

dist/profileReadLambda.zip: dist/profileRead
	cd dist && zip profileReadLambda.zip profileRead

dist/profileWrite: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/profileWrite github.com/jonsabados/pointypoints/profile/write

dist/profileWriteLambda.zip: dist/profileWrite
	cd dist && zip profileWriteLambda.zip profileWrite

build: frontend/dist/index.html dist/corsLambda.zip dist/newSessionLambda.zip dist/connectLambda.zip \
	dist/disconnectLambda.zip dist/setFacilitatorSessionLambda.zip dist/watchSessionLambda.zip \
	dist/joinSessionLambda.zip dist/voteLambda.zip dist/updateSessionLambda.zip dist/clearVotesLambda.zip \
	dist/pingLambda.zip dist/authorizerLambda.zip dist/profileReadLambda.zip dist/profileWriteLambda.zip
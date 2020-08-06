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

.PHONY: test
test:
	cd frontend && npm run test:unit
	@./execute-tests.sh

.PHONY: clean
clean:
	rm -rf frontend/dist/ frontend/.env.local
	rm -rf dist

.PHONY: run
run: frontend/.env.local
	cd frontend && npm run serve

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

dist/loadFacilitatorSession: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/loadFacilitatorSession github.com/jonsabados/pointypoints/session/loadfacilitatorsession

dist/loadFacilitatorSessionLambda.zip: dist/loadFacilitatorSession
	cd dist && zip loadFacilitatorSessionLambda.zip loadFacilitatorSession

dist/loadSession: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/loadSession github.com/jonsabados/pointypoints/session/loadsession

dist/loadSessionLambda.zip: dist/loadSession
	cd dist && zip loadSessionLambda.zip loadSession

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

dist/clearVotes: dist/ $(shell find . -iname "*.go")
	GOOS=linux go build -o dist/clearVotes github.com/jonsabados/pointypoints/session/clearvotes

dist/clearVotesLambda.zip: dist/clearVotes
	cd dist && zip clearVotesLambda.zip clearVotes

build: frontend/dist/index.html dist/newSessionLambda.zip dist/connectLambda.zip dist/disconnectLambda.zip \
	dist/loadFacilitatorSessionLambda.zip dist/loadSessionLambda.zip dist/joinSessionLambda.zip dist/voteLambda.zip \
	dist/showVotesLambda.zip dist/clearVotesLambda.zip
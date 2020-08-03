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

build: frontend/dist/index.html dist/newSessionLambda.zip dist/connectLambda.zip dist/disconnectLambda.zip
.DEFAULT_GOAL := build

dist/:
	mkdir dist

frontend/.env.local:
	cd frontend && ./gen_env.sh

frontend/dist/index.html: $(shell find frontend/src) $(shell find frontend/public) frontend/.env.local
	cd frontend && npm run build

.PHONY: test
test:
	cd frontend && npm run test:unit
	cd backend/src/go && go test ./... --race --cover

.PHONY: clean
clean:
	rm -rf frontend/dist/ frontend/.env.local
	rm -rf dist

.PHONY: run
run: frontend/.env.local
	cd frontend && npm run serve

build: frontend/dist/index.html
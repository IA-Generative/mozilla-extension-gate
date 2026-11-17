run:
	docker compose -f docker-compose.yml up -d --build && make logs

seed:
	docker compose --profile seed -f docker-compose.yml up db-seed --build

stop:
	docker compose -f docker-compose.yml down

stop-all:
	docker compose --profile seed -f docker-compose.yml down -v

logs:
	docker compose -f docker-compose.yml logs -f api

shell:
	docker compose -f docker-compose.yml exec api /bin/sh

test:
	docker compose -f docker-compose.yml exec api pytest

test-coverage:
	docker compose -f docker-compose.yml exec api coverage run -m pytest --junitxml=junit.xml tests/

coverage-report:
	docker compose -f docker-compose.yml exec api coverage report

checkreq:
	pip install --upgrade pur
	pur --dry-run-changed -r requirements.dev.txt

installreq:
	pip install --upgrade uv
	uv pip install --native-tls -r requirements.dev.txt

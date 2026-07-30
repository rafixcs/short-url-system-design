.PHONY: perf-up perf-smoke perf-load perf-stress perf-down

PERF_COMPOSE := docker compose -f infra/docker-compose.k6.yaml

perf-up:
	$(PERF_COMPOSE) up -d --build app

perf-smoke: perf-up
	TEST_TYPE=smoke $(PERF_COMPOSE) run --rm k6

perf-load: perf-up
	TEST_TYPE=load $(PERF_COMPOSE) run --rm k6

perf-stress: perf-up
	TEST_TYPE=stress $(PERF_COMPOSE) run --rm k6

perf-down:
	$(PERF_COMPOSE) down --remove-orphans

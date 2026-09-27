# Tự động nạp các biến môi trường từ file .env nếu file này tồn tại
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Các cấu hình mặc định của bạn
ATLAS ?= atlas
ATLAS_ENV ?= local

.PHONY: migrate-diff migrate-apply migrate-status migrate-down migrate-lint

migrate-diff:
	@test -n "$(name)" || (echo "Usage: make migrate-diff name=<migration_name>"; exit 1)
	$(ATLAS) migrate diff "$(name)" --env "$(ATLAS_ENV)"

migrate-apply:
	@test -n "$$ATLAS_DB_URL" || (echo "ATLAS_DB_URL must be set in .env file"; exit 1)
	$(ATLAS) migrate apply --env "$(ATLAS_ENV)"

migrate-status:
	@test -n "$$ATLAS_DB_URL" || (echo "ATLAS_DB_URL must be set in .env file"; exit 1)
	$(ATLAS) migrate status --env "$(ATLAS_ENV)"

migrate-down:
	@test -n "$$ATLAS_DB_URL" || (echo "ATLAS_DB_URL must be set in .env file"; exit 1)
	$(ATLAS) migrate down --env "$(ATLAS_ENV)"

migrate-lint:
	$(ATLAS) migrate validate --env "$(ATLAS_ENV)"
	$(ATLAS) migrate lint --env "$(ATLAS_ENV)" --latest 1

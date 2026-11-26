.PHONY: generate clientes migrate

OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.2

generate:
	cd apps/api && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

clientes:
	cd apps/api && $(OAPI_CODEGEN) -generate types,client -package productos -o internal/clientes/productos/client.gen.go ../../contracts/externos/productos.yaml
	cd apps/api && $(OAPI_CODEGEN) -generate types,client -package stock -o internal/clientes/stock/client.gen.go ../../contracts/externos/stock.yaml

migrate:
	cd apps/api && go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir db/migrations postgres "$$DATABASE_URL" up

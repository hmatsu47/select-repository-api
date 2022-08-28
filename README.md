# select-repository-api

Go で内部管理用 API を作るテスト

## `.yaml`ファイルから API コードの枠組みを生成

```sh:install
go mod init github.com/hmatsu47/select-repository-api
mkdir docs
mkdir -p gen/openapi
（作成した`.yaml`ファイルを`openapi.yaml`として`docs`内にコピー）
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0
oapi-codegen --output-config --old-config-style --package="main" --generate="types,gin,spec" -alias-types docs/openapi.yaml > docs/config.yaml
oapi-codegen --config docs/config.yaml docs/openapi.yaml > gen/openapi/server.gen.go
go mod tidy
```

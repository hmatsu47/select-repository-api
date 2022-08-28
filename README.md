# select-repository-api

Go で内部管理用 API を作るテスト

## `.yaml`ファイルから API コードの枠組みを生成

- やり方を試行錯誤するため、一旦`go generate`を使わずに準備

```sh:install
go mod init github.com/hmatsu47/select-repository-api
mkdir internal
cd internal
（とりあえず生成したコード自体は実行コードとして使わないので internal に入れておく）
mkdir docs
mkdir -p gen/openapi
（作成した`.yaml`ファイルを`openapi.yaml`として`docs`内にコピー）
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0
oapi-codegen -output-config -old-config-style -package=openapi -generate=types -alias-types docs/openapi.yaml > docs/config-types.yaml
oapi-codegen -output-config -old-config-style -package=openapi -generate=gin,spec -alias-types docs/openapi.yaml > docs/config-server.yaml
oapi-codegen -config docs/config-types.yaml docs/openapi.yaml > gen/openapi/types.gen.go
oapi-codegen -config docs/config-server.yaml docs/openapi.yaml > gen/openapi/server.gen.go
cd ..
go mod tidy
```

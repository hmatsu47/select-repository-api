# select-repository-api

Go で内部管理用 API を作るテスト

## `.yaml`ファイルから API コードの枠組みを生成

- やり方を試行錯誤するため、一旦`go generate`を使わずに準備
- ディレクトリ構成は ↓ を参考に
  - https://github.com/deepmap/oapi-codegen/tree/master/examples/petstore-expanded/gin

```sh:install
go mod init github.com/hmatsu47/select-repository-api
mkdir internal
cd internal
（作成した`.yaml`ファイルを`internal`内にコピー）
cd ..
mkdir api
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0
oapi-codegen -output-config -old-config-style -package=api -generate=types -alias-types internal/select-repository.yaml > api/config-types.yaml
oapi-codegen -output-config -old-config-style -package=api -generate=gin,spec -alias-types internal/select-repository.yaml > api/config-server.yaml
oapi-codegen -config api/config-types.yaml internal/select-repository.yaml > api/types.gen.go
oapi-codegen -config api/config-server.yaml internal/select-repository.yaml > api/server.gen.go
go mod tidy
```

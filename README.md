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

## 起動方法

`go run main.go [-port=待機ポート番号（TCP）] [ワークディレクトリ [cron.d書き出しパス＋ファイル名プレフィックス [cron.dで実行するリリーススクリプト [ログの書き出し指定]]]]`

※ `cron.dで実行するリリーススクリプト` ・ `ログの書き出し指定` の中に `[SERVICE-NAME]` という文字列を入れるとサービス名で置換される

- 開発モードの例
  - `go build`後のバイナリでも同じパラメータの指定が可能
- 待機ポート番号を省略すると 8080 番で待機
- ワークディレクトリを省略すると`/var/select-repository`
- ワークディレクトリには以下のファイルを配置
  - `services` : サービス名一覧
    - 1 行 1 サービス名
  - `【サービス名】-repositories` : 対象サービスのリポジトリ URI 一覧
    - 1 行 1 URI
- 実行する中で以下のファイルが生成される想定
  - `【サービス名】-release-setting` : 次回リリース設定
    - 選択イメージ URI とリリース予定日時を保存
    - この API で生成
  - `【サービス名】-released` : 前回リリース設定
    - 外部のリリース処理スクリプト（`cron`で毎分実行）でリリース処理を終えた際に`【サービス名】-release-setting`をリネーム
  - `【サービス名】-release-processing` : リリース処理中を示すファイル
    - `flock`でロックファイルとして生成
    - このファイルがあると API では`【サービス名】-release-setting`を上書き保存せず 500 Error を返す
  - `【cron.d 書き出しパス＋ファイル名プレフィックス】-【サービス名】`
    - このファイルのみワークディレクトリ外に生成
      - `/etc/cron.d`に生成する想定
        - 例 : `5 4 22 9 * root flock /var/select-repository/test-release-processing sh /usr/local/sbin/release-test.sh 000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1:20220922-release >> /var/log/release-log-test && mv /var/select-repository/test-release-setting /var/select-repository/test-released && rm -f /etc/cron.d/test-release-setting && rm -f /var/select-repository/test-release-processing`

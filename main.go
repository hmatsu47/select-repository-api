package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	middleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
	"github.com/hmatsu47/select-repository-api/api"
)

func NewGinSelectRepositoryServer(selectRepository *api.SelectRepository, port int) *http.Server {
    swagger, err := api.GetSwagger()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Swagger specの読み取りに失敗しました\n: %s", err)
        os.Exit(1)
    }

    // Swagger Document 非公開
    swagger.Servers = nil

    // Gin Router 設定
    r := gin.Default()

    // HTTP Request の Validation 設定
    r.Use(middleware.OapiRequestValidator(swagger))

    // Handler 実装
    r = api.RegisterHandlers(r, selectRepository)

    s := &http.Server{
        Handler: r,
        Addr:    fmt.Sprintf("0.0.0.0:%d", port),
    }
    return s
}

func main() {
    port := flag.Int("port", 8080, "Port for API server")
    flag.Parse()
    // Work Directory はコマンドラインパラメータで取得→設定を取得
    workDir := flag.Arg(0)
    // cron.d への書き出し用の指定（コマンド部分とログ書き出し部分）
    cronDir := "/etc/cron.d"
    cronCmd := flag.Arg(1)
    cronLog := flag.Arg(2)
    // Server Instance 生成
    selectRepository := api.NewSelectRepository(workDir, cronDir, cronCmd, cronLog)
    s := NewGinSelectRepositoryServer(selectRepository, *port)
    // 停止まで HTTP Request を処理
    log.Fatal(s.ListenAndServe())
}

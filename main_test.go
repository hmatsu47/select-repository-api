package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
    "os"
    "path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
    "github.com/hmatsu47/select-repository-api/api"
)

func doGet(t *testing.T, handler http.Handler, url string) *httptest.ResponseRecorder {
	response := testutil.NewRequest().Get(url).WithAcceptJson().GoWithHTTPHandler(t, handler)
	return response.Recorder
}

// ファイルコピー
func fileCopy(srcPath string, dstPath string) (string, error) {
    src, err := os.Open(srcPath)
    if err != nil {
        return srcPath, err
    }
    defer src.Close()

    dst, err := os.Create(dstPath)
    if err != nil {
        return dstPath, err
    }
    defer dst.Close()

    _, err = io.Copy(dst, src)
    if  err != nil {
        return dstPath, err
    }
    return dstPath, err
}

// 設定をテンポラリディレクトリにコピー
func initConfig(templateConfigDir string) string {
    var err error
    tmpConfigDir, err := os.MkdirTemp("", "select-repository-test-config")
    if err != nil {
        panic(err)
    }
    files, err := ioutil.ReadDir(tmpConfigDir)
    if err != nil {
        panic(err)
    }
    for _, file := range files {
        _, err = fileCopy(filepath.Join(templateConfigDir, file.Name()), filepath.Join(tmpConfigDir, file.Name()))
        if  err != nil {
            panic(err)
        }
    }
    return tmpConfigDir
}

// 設定を削除
func clearConfig(tmpConfigDir string) {
    os.RemoveAll(tmpConfigDir)
}

func TestSelectRepository1(t *testing.T) {
	var err error
	templateConfigDir := "./test/config1-single-no-setting"
	workDir := initConfig(templateConfigDir)
	selectRepository := api.NewSelectRepository(workDir)
	ginSelectRepositoryServer := NewGinSelectRepositoryServer(selectRepository, 8080)
	r := ginSelectRepositoryServer.Handler
	
	defer clearConfig(workDir)
	
	t.Run("単一サービス・リリース未設定・サービス一覧取得", func(t *testing.T) {
	    rr := doGet(t, r, "/services")
	    
		var serviceNameList []api.ServiceName
		err = json.NewDecoder(rr.Body).Decode(&serviceNameList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 1, len(serviceNameList))
	})
}
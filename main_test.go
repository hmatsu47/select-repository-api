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
    files, err := ioutil.ReadDir(templateConfigDir)
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

// go test -v . で実行する
func TestSelectRepository1(t *testing.T) {
	var err error
	templateConfigDir := "./test/config1-single-no-setting"
	workDir := initConfig(templateConfigDir)
	selectRepository := api.NewSelectRepository(workDir)

    defer clearConfig(workDir)
 
    t.Run("単一サービス・リリース未設定・設定チェック", func(t *testing.T) {
        var serviceNameList []string = selectRepository.ServiceName
        assert.Equal(t, 1, len(serviceNameList))
        assert.Equal(t, "test1", serviceNameList[0])
        repositoryList := selectRepository.RepositoryMap["test1"]
        assert.Equal(t, 1, len(repositoryList))
        assert.Equal(t, "repository1", repositoryList[0].Name)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList[0].Uri)
        repositoryList2d := selectRepository.RepositoryMap2d[api.RepositoryKey{
            ServiceName: "test1",
            RepositoryName: "repository1"}]
        assert.Equal(t, "000000000000", repositoryList2d.RegistryId)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList2d.Uri)
    })

	ginSelectRepositoryServer := NewGinSelectRepositoryServer(selectRepository, 8080)
	r := ginSelectRepositoryServer.Handler
	
	t.Run("単一サービス・リリース未設定・サービス一覧取得", func(t *testing.T) {
	    rr := doGet(t, r, "/services")
	    
		var serviceList []api.Service
		err = json.NewDecoder(rr.Body).Decode(&serviceList)
		assert.NoError(t, err, "error getting response")
		assert.Equal(t, 1, len(serviceList))
	})
    	
	t.Run("単一サービス・リリース未設定・サービス一覧＆リポジトリ一覧取得", func(t *testing.T) {
	    rr := doGet(t, r, "/repositories/test1")

        var repositoryList []api.Repository
        err = json.NewDecoder(rr.Body).Decode(&repositoryList)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, 1, len(repositoryList))
        assert.Equal(t, "repository1", repositoryList[0].Name)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList[0].Uri)
	})
        	
	t.Run("単一サービス・リリース未設定・リリース設定（なし）取得", func(t *testing.T) {
	    rr := doGet(t, r, "/setting/test1")

        var setting api.Setting
        err = json.NewDecoder(rr.Body).Decode(&setting)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, false, setting.IsReleased)
        assert.Nil(t, setting.ImageUri)
        assert.Nil(t, setting.ReleaseAt)
	})
}

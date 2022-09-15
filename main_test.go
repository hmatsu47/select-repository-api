package main

import (
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"
    "time"

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
    fmt.Printf("テスト用のテンポラリディレクトリ（%s）を作成しました\n", tmpConfigDir)
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
    fmt.Printf("テスト用のテンポラリディレクトリ（%s）を削除しました\n", tmpConfigDir)
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
        assert.Equal(t, "test1", serviceList[0].Name)
    })
        
    t.Run("単一サービス・リリース未設定・リポジトリ一覧取得", func(t *testing.T) {
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

func TestSelectRepository2(t *testing.T) {
    var err error
    templateConfigDir := "./test/config2-single-released-setting-only"
    workDir := initConfig(templateConfigDir)
    selectRepository := api.NewSelectRepository(workDir)

    defer clearConfig(workDir)
 
    ginSelectRepositoryServer := NewGinSelectRepositoryServer(selectRepository, 8080)
    r := ginSelectRepositoryServer.Handler
    
    t.Run("単一サービス・過去リリースあり・リリース設定（過去のみ）取得", func(t *testing.T) {
        rr := doGet(t, r, "/setting/test1")

        var setting api.Setting
        err = json.NewDecoder(rr.Body).Decode(&setting)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, true, setting.IsReleased)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1:20220831-release", *setting.ImageUri)
        expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-08-31T23:50:00+09:00")
        assert.Equal(t, expectedTime, *setting.ReleaseAt)
    })
}

func TestSelectRepository3(t *testing.T) {
    var err error
    templateConfigDir := "./test/config3-single-new-setting-only"
    workDir := initConfig(templateConfigDir)
    selectRepository := api.NewSelectRepository(workDir)

    defer clearConfig(workDir)
 
    ginSelectRepositoryServer := NewGinSelectRepositoryServer(selectRepository, 8080)
    r := ginSelectRepositoryServer.Handler
    
    t.Run("単一サービス・過去リリースなし・リリース設定（あり）取得", func(t *testing.T) {
        rr := doGet(t, r, "/setting/test1")

        var setting api.Setting
        err = json.NewDecoder(rr.Body).Decode(&setting)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, false, setting.IsReleased)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1:20220911-release", *setting.ImageUri)
        expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-10T19:05:00Z")
        assert.Equal(t, expectedTime, *setting.ReleaseAt)
    })
}

func TestSelectRepository4(t *testing.T) {
    var err error
    templateConfigDir := "./test/config3-single-new-setting-only"
    workDir := initConfig(templateConfigDir)
    selectRepository := api.NewSelectRepository(workDir)

    defer clearConfig(workDir)
 
    ginSelectRepositoryServer := NewGinSelectRepositoryServer(selectRepository, 8080)
    r := ginSelectRepositoryServer.Handler
    
    t.Run("単一サービス・過去リリースあり・リリース設定（あり）取得", func(t *testing.T) {
        rr := doGet(t, r, "/setting/test1")

        var setting api.Setting
        err = json.NewDecoder(rr.Body).Decode(&setting)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, false, setting.IsReleased)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1:20220911-release", *setting.ImageUri)
        expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-10T19:05:00Z")
        assert.Equal(t, expectedTime, *setting.ReleaseAt)
    })
}

func TestSelectRepository5(t *testing.T) {
    var err error
    templateConfigDir := "./test/config5-double"
    workDir := initConfig(templateConfigDir)
    selectRepository := api.NewSelectRepository(workDir)

    defer clearConfig(workDir)
 
    t.Run("サービスx2・設定チェック", func(t *testing.T) {
        var serviceNameList []string = selectRepository.ServiceName
        assert.Equal(t, 2, len(serviceNameList))
        assert.Equal(t, "test1", serviceNameList[0])
        assert.Equal(t, "test2", serviceNameList[1])
        repositoryList := selectRepository.RepositoryMap["test1"]
        assert.Equal(t, 1, len(repositoryList))
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList[0].Uri)
        repositoryList2d := selectRepository.RepositoryMap2d[api.RepositoryKey{
            ServiceName: "test1",
            RepositoryName: "repository1"}]
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList2d.Uri)
        repositoryList2 := selectRepository.RepositoryMap["test2"]
        assert.Equal(t, 2, len(repositoryList2))
        assert.Equal(t, "repository21", repositoryList2[0].Name)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository21", repositoryList2[0].Uri)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository22", repositoryList2[1].Uri)
        repositoryList22d := selectRepository.RepositoryMap2d[api.RepositoryKey{
            ServiceName: "test2",
            RepositoryName: "repository22"}]
        assert.Equal(t, "000000000000", repositoryList22d.RegistryId)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository22", repositoryList22d.Uri)
    })

    ginSelectRepositoryServer := NewGinSelectRepositoryServer(selectRepository, 8080)
    r := ginSelectRepositoryServer.Handler
    
    t.Run("サービスx2・サービス一覧取得", func(t *testing.T) {
        rr := doGet(t, r, "/services")
        
        var serviceList []api.Service
        err = json.NewDecoder(rr.Body).Decode(&serviceList)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, 2, len(serviceList))
        assert.Equal(t, "test2", serviceList[1].Name)
    })

    t.Run("サービスx2・リポジトリ一覧取得", func(t *testing.T) {
        rr := doGet(t, r, "/repositories/test2")

        var repositoryList []api.Repository
        err = json.NewDecoder(rr.Body).Decode(&repositoryList)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, 2, len(repositoryList))
        assert.Equal(t, "repository21", repositoryList[0].Name)
        assert.Equal(t, "repository22", repositoryList[1].Name)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository21", repositoryList[0].Uri)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository22", repositoryList[1].Uri)
    })
    
    t.Run("サービスx2・過去リリースなし・リリース設定（あり）取得", func(t *testing.T) {
        rr := doGet(t, r, "/setting/test2")

        var setting api.Setting
        err = json.NewDecoder(rr.Body).Decode(&setting)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, false, setting.IsReleased)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository22:20220912-release", *setting.ImageUri)
        expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-11T19:05:00Z")
        assert.Equal(t, expectedTime, *setting.ReleaseAt)
    })
    
    t.Run("サービスx2・リリース設定更新（失敗）", func(t *testing.T) {
        testImageUri := "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository22:20220921-release"
        testReleaseAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-20T19:05:00Z")
        setting := api.Setting{
            ImageUri: &testImageUri,
            IsReleased: false,
            ReleaseAt: &testReleaseAt,
        }
        rr := testutil.NewRequest().Post("/setting/test2").WithJsonBody(setting).GoWithHTTPHandler(t, r).Recorder
        // レスポンスを確認（リリース処理中なのでエラーメッセージが返る想定）
        var resultErrorMessage api.Error
        err = json.NewDecoder(rr.Body).Decode(&resultErrorMessage)
        assert.NoError(t, err, "error getting response")
        expectMessage := "現在リリース処理中です。完了までしばらくお待ちください"
        message := resultErrorMessage.Message
        assert.Equal(t, expectMessage, message)
        // 実際の設定を確認（ファイルに上書き保存されていないか？）
        settingFile := fmt.Sprintf("%s/test2-release-setting", workDir)
        settingItems, err := api.ReadSettingFromFile(settingFile)
        assert.NoError(t, err, "変更不可のリリース設定が上書きされています")
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository22:20220912-release", settingItems.ImageUri)
        expectReleaseAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-11T19:05:00Z")
        assert.Equal(t, expectReleaseAt, settingItems.ReleaseAt)
    })
}

func TestSelectRepository6(t *testing.T) {
    var err error
    templateConfigDir := "./test/config6-triple"
    workDir := initConfig(templateConfigDir)
    selectRepository := api.NewSelectRepository(workDir)

    defer clearConfig(workDir)
 
    t.Run("サービスx3・設定チェック", func(t *testing.T) {
        var serviceNameList []string = selectRepository.ServiceName
        assert.Equal(t, 3, len(serviceNameList))
        assert.Equal(t, "test1", serviceNameList[0])
        assert.Equal(t, "test2", serviceNameList[1])
        assert.Equal(t, "test3", serviceNameList[2])
        repositoryList := selectRepository.RepositoryMap["test1"]
        assert.Equal(t, 1, len(repositoryList))
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList[0].Uri)
        repositoryList2d := selectRepository.RepositoryMap2d[api.RepositoryKey{
            ServiceName: "test1",
            RepositoryName: "repository1"}]
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList2d.Uri)
        repositoryList3 := selectRepository.RepositoryMap["test3"]
        assert.Equal(t, 3, len(repositoryList3))
        assert.Equal(t, "repository31", repositoryList3[0].Name)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository31", repositoryList3[0].Uri)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33", repositoryList3[2].Uri)
        repositoryList32d := selectRepository.RepositoryMap2d[api.RepositoryKey{
            ServiceName: "test3",
            RepositoryName: "repository32"}]
        assert.Equal(t, "000000000000", repositoryList32d.RegistryId)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository32", repositoryList32d.Uri)
    })

    ginSelectRepositoryServer := NewGinSelectRepositoryServer(selectRepository, 8080)
    r := ginSelectRepositoryServer.Handler
    
    t.Run("サービスx3・サービス一覧取得", func(t *testing.T) {
        rr := doGet(t, r, "/services")
        
        var serviceList []api.Service
        err = json.NewDecoder(rr.Body).Decode(&serviceList)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, 3, len(serviceList))
        assert.Equal(t, "test3", serviceList[2].Name)
    })

    t.Run("サービスx3・リポジトリ一覧取得", func(t *testing.T) {
        rr := doGet(t, r, "/repositories/test3")

        var repositoryList []api.Repository
        err = json.NewDecoder(rr.Body).Decode(&repositoryList)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, 3, len(repositoryList))
        assert.Equal(t, "repository31", repositoryList[0].Name)
        assert.Equal(t, "repository32", repositoryList[1].Name)
        assert.Equal(t, "repository33", repositoryList[2].Name)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository31", repositoryList[0].Uri)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository32", repositoryList[1].Uri)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33", repositoryList[2].Uri)
    })
    
    t.Run("サービスx3・過去リリースあり・リリース設定（過去のみ）取得", func(t *testing.T) {
        rr := doGet(t, r, "/setting/test3")

        var setting api.Setting
        err = json.NewDecoder(rr.Body).Decode(&setting)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, true, setting.IsReleased)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220915-release", *setting.ImageUri)
        expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-15T23:30:00+09:00")
        assert.Equal(t, expectedTime, *setting.ReleaseAt)
    })
    
    t.Run("サービスx3・リリース設定保存（成功）", func(t *testing.T) {
        testImageUri := "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release"
        testReleaseAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-22T22:30:00+09:00")
        setting := api.Setting{
            ImageUri: &testImageUri,
            IsReleased: false,
            ReleaseAt: &testReleaseAt,
        }
        rr := testutil.NewRequest().Post("/setting/test3").WithJsonBody(setting).GoWithHTTPHandler(t, r).Recorder
        // レスポンスを確認
        var resultSetting api.Setting
        err = json.NewDecoder(rr.Body).Decode(&resultSetting)
        assert.NoError(t, err, "error getting response")
        assert.Equal(t, false, resultSetting.IsReleased)
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release", *resultSetting.ImageUri)
        assert.Equal(t, testReleaseAt, *resultSetting.ReleaseAt)
        // 実際の設定を確認（ファイルに正しく保存されたか？）
        settingFile := fmt.Sprintf("%s/test3-release-setting", workDir)
        settingItems, err := api.ReadSettingFromFile(settingFile)
        assert.NoError(t, err, "リリース設定が保存されていないか、設定内容が不正です")
        assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release", settingItems.ImageUri)
        assert.Equal(t, testReleaseAt, settingItems.ReleaseAt)
    })
}

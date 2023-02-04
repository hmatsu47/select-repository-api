package main

import (
	"bufio"
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

	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
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
	if err != nil {
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
		if err != nil {
			panic(err)
		}
	}
	return tmpConfigDir
}

// cron.d 用のテンポラリディレクトリを作成
func initCronPath() string {
	tmpCronPath, err := os.MkdirTemp("", "cron.d")
	if err != nil {
		panic(err)
	}
	fmt.Printf("テスト用のテンポラリディレクトリ（%s）を作成しました\n", tmpCronPath)

	return fmt.Sprintf("%s/", tmpCronPath)
}

// cron.d 内のファイルを読み取り
func readCron(cronPath string, serviceName string) string {
	cronFile := fmt.Sprintf("%s/test3-release", cronPath)
	f, err := os.Open(cronFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	cron := scanner.Text()

	return cron
}

// テンポラリディレクトリを削除
func clearTempDir(tmpDir string) {
	os.RemoveAll(tmpDir)
	fmt.Printf("テスト用のテンポラリディレクトリ（%s）を削除しました\n", tmpDir)
}

// go test -v で実行する
func TestSelectRepository1(t *testing.T) {
	var err error
	templateConfigDir := "./test/config1-single-no-setting"
	workDir := initConfig(templateConfigDir)
	cronPath := initCronPath()
	cronCmd := "echo [SERVICE-NAME]"
	cronLog := ">> /dev/null [SERVICE-NAME]"
	selectRepository := api.NewSelectRepository(workDir, cronPath, cronCmd, cronLog)

	t.Cleanup(func() {
		clearTempDir(workDir)
		clearTempDir(cronPath)
	})

	t.Run("単一サービス・リリース未設定・設定チェック", func(t *testing.T) {
		var serviceNameList []string = selectRepository.ServiceName
		assert.Equal(t, 1, len(serviceNameList))
		assert.Equal(t, "test1", serviceNameList[0])
		repositoryList := selectRepository.RepositoryMap["test1"]
		assert.Equal(t, 1, len(repositoryList))
		assert.Equal(t, "repository1", repositoryList[0].Name)
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList[0].Uri)
		repositoryList2d := selectRepository.RepositoryMap2d[api.RepositoryKey{
			ServiceName:    "test1",
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

	t.Run("単一サービス・リリース未設定・リリース設定（なし）削除（失敗）", func(t *testing.T) {
		rr := testutil.NewRequest().Delete("/setting/test1").GoWithHTTPHandler(t, r).Recorder
		// レスポンスを確認（リリース設定がないのでエラーメッセージが返る想定）
		var resultErrorMessage api.Error
		err = json.NewDecoder(rr.Body).Decode(&resultErrorMessage)
		assert.NoError(t, err, "error getting response")
		expectMessage := fmt.Sprintf("設定の削除が失敗しました : remove %stest1-release: no such file or directory", cronPath)
		message := resultErrorMessage.Message
		assert.Equal(t, expectMessage, message)
	})
}

func TestSelectRepository2(t *testing.T) {
	var err error
	templateConfigDir := "./test/config2-single-released-setting-only"
	workDir := initConfig(templateConfigDir)
	cronPath := initCronPath()
	cronCmd := "echo [SERVICE-NAME]"
	cronLog := ">> /dev/null [SERVICE-NAME]"
	selectRepository := api.NewSelectRepository(workDir, cronPath, cronCmd, cronLog)

	t.Cleanup(func() {
		clearTempDir(workDir)
		clearTempDir(cronPath)
	})

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

	t.Run("単一サービス・リリース未設定・リリース設定（過去のみ）削除（失敗）", func(t *testing.T) {
		rr := testutil.NewRequest().Delete("/setting/test1").GoWithHTTPHandler(t, r).Recorder
		// レスポンスを確認（リリース設定がないのでエラーメッセージが返る想定）
		var resultErrorMessage api.Error
		err = json.NewDecoder(rr.Body).Decode(&resultErrorMessage)
		assert.NoError(t, err, "error getting response")
		expectMessage := fmt.Sprintf("設定の削除が失敗しました : remove %stest1-release: no such file or directory", cronPath)
		message := resultErrorMessage.Message
		assert.Equal(t, expectMessage, message)
	})
}

func TestSelectRepository3(t *testing.T) {
	var err error
	templateConfigDir := "./test/config3-single-new-setting-only"
	workDir := initConfig(templateConfigDir)
	cronPath := initCronPath()
	cronCmd := "echo [SERVICE-NAME]"
	cronLog := ">> /dev/null [SERVICE-NAME]"
	selectRepository := api.NewSelectRepository(workDir, cronPath, cronCmd, cronLog)

	t.Cleanup(func() {
		clearTempDir(workDir)
		clearTempDir(cronPath)
	})

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
	cronPath := initCronPath()
	cronCmd := "echo [SERVICE-NAME]"
	cronLog := ">> /dev/null [SERVICE-NAME]"
	selectRepository := api.NewSelectRepository(workDir, cronPath, cronCmd, cronLog)

	t.Cleanup(func() {
		clearTempDir(workDir)
		clearTempDir(cronPath)
	})

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
	cronPath := initCronPath()
	cronCmd := "echo [SERVICE-NAME]"
	cronLog := ">> /dev/null [SERVICE-NAME]"
	selectRepository := api.NewSelectRepository(workDir, cronPath, cronCmd, cronLog)

	t.Cleanup(func() {
		clearTempDir(workDir)
		clearTempDir(cronPath)
	})

	t.Run("サービスx2・設定チェック", func(t *testing.T) {
		var serviceNameList []string = selectRepository.ServiceName
		assert.Equal(t, 2, len(serviceNameList))
		assert.Equal(t, "test1", serviceNameList[0])
		assert.Equal(t, "test2", serviceNameList[1])
		repositoryList := selectRepository.RepositoryMap["test1"]
		assert.Equal(t, 1, len(repositoryList))
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList[0].Uri)
		repositoryList2d := selectRepository.RepositoryMap2d[api.RepositoryKey{
			ServiceName:    "test1",
			RepositoryName: "repository1"}]
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList2d.Uri)
		repositoryList2 := selectRepository.RepositoryMap["test2"]
		assert.Equal(t, 2, len(repositoryList2))
		assert.Equal(t, "repository21", repositoryList2[0].Name)
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository21", repositoryList2[0].Uri)
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository22", repositoryList2[1].Uri)
		repositoryList22d := selectRepository.RepositoryMap2d[api.RepositoryKey{
			ServiceName:    "test2",
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
			ImageUri:   &testImageUri,
			IsReleased: false,
			ReleaseAt:  &testReleaseAt,
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

	t.Run("サービスx2・過去リリースなし・リリース設定削除（失敗）", func(t *testing.T) {
		rr := testutil.NewRequest().Delete("/setting/test2").GoWithHTTPHandler(t, r).Recorder
		// レスポンスを確認（リリース処理中なのでエラーメッセージが返る想定）
		var resultErrorMessage api.Error
		err = json.NewDecoder(rr.Body).Decode(&resultErrorMessage)
		assert.NoError(t, err, "error getting response")
		expectMessage := "すでにリリース処理が開始されています。取り消しできません"
		message := resultErrorMessage.Message
		assert.Equal(t, expectMessage, message)
	})
}

func TestSelectRepository6(t *testing.T) {
	var err error
	templateConfigDir := "./test/config6-triple"
	workDir := initConfig(templateConfigDir)
	cronPath := initCronPath()
	cronCmd := "echo [SERVICE-NAME]"
	cronLog := ">> /dev/null [SERVICE-NAME]"
	selectRepository := api.NewSelectRepository(workDir, cronPath, cronCmd, cronLog)

	t.Cleanup(func() {
		clearTempDir(workDir)
		clearTempDir(cronPath)
	})

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
			ServiceName:    "test1",
			RepositoryName: "repository1"}]
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository1", repositoryList2d.Uri)
		repositoryList3 := selectRepository.RepositoryMap["test3"]
		assert.Equal(t, 3, len(repositoryList3))
		assert.Equal(t, "repository31", repositoryList3[0].Name)
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository31", repositoryList3[0].Uri)
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33", repositoryList3[2].Uri)
		repositoryList32d := selectRepository.RepositoryMap2d[api.RepositoryKey{
			ServiceName:    "test3",
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

	t.Run("サービスx3・リリース設定（指定日時）保存（成功）→削除（成功）", func(t *testing.T) {
		// 保存
		testImageUri := "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release"
		testReleaseAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2122-09-02T22:30:00+09:00")
		setting := api.Setting{
			ImageUri:   &testImageUri,
			IsReleased: false,
			ReleaseAt:  &testReleaseAt,
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
		// cron.d に出力されたファイルの内容を確認
		cron := readCron(cronPath, "test3")
		expected1 := fmt.Sprintf("30 22 2 9 * root flock %s/test3-release-processing ", workDir)
		expected2 := "echo test3 000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release >> /dev/null test3 && "
		expected3 := fmt.Sprintf("mv -f %s/test3-release-setting %s/test3-released && rm -f %stest3-release && ", workDir, workDir, cronPath)
		expected4 := fmt.Sprintf("rm -f %s/test3-release-processing", workDir)
		assert.Equal(t, expected1+expected2+expected3+expected4, cron)
		// 削除
		rr = testutil.NewRequest().Delete("/setting/test3").GoWithHTTPHandler(t, r).Recorder
		// レスポンスを確認
		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		// 実際のcron.d に出力されたファイルを確認
		_, err = os.Stat(fmt.Sprintf("%stest3-release", cronPath))
		assert.NotEqual(t, err, nil)
		// 実際の設定を確認
		_, err = os.Stat(fmt.Sprintf("%s/test3-release-setting", workDir))
		assert.NotEqual(t, err, nil)
	})

	t.Run("サービスx3・リリース設定（即時リリース）保存（成功）", func(t *testing.T) {
		testImageUri := "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release"
		testReleaseAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-02T22:30:00+09:00")
		tmpNow := time.Now().In(time.Local).Add(1 * time.Minute)
		now := time.Date(tmpNow.Year(), tmpNow.Month(), tmpNow.Day(), tmpNow.Hour(), tmpNow.Minute(), 0, 0, time.Local)
		setting := api.Setting{
			ImageUri:   &testImageUri,
			IsReleased: false,
			ReleaseAt:  &testReleaseAt,
		}
		rr := testutil.NewRequest().Post("/setting/test3").WithJsonBody(setting).GoWithHTTPHandler(t, r).Recorder
		// レスポンスを確認
		var resultSetting api.Setting
		err = json.NewDecoder(rr.Body).Decode(&resultSetting)
		assert.NoError(t, err, "error getting response")
		assert.Equal(t, false, resultSetting.IsReleased)
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release", *resultSetting.ImageUri)
		assert.Equal(t, now, *resultSetting.ReleaseAt)
		// 実際の設定を確認（ファイルに正しく保存されたか？）
		settingFile := fmt.Sprintf("%s/test3-release-setting", workDir)
		settingItems, err := api.ReadSettingFromFile(settingFile)
		assert.NoError(t, err, "リリース設定が保存されていないか、設定内容が不正です")
		assert.Equal(t, "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33:20220922-release", settingItems.ImageUri)
		assert.Equal(t, now, settingItems.ReleaseAt)
		// cron.d に出力されたファイルの内容を確認
		cron := readCron(cronPath, "test3")
		expected1 := fmt.Sprintf("%d %d %d %d * root flock %s/test3-release-processing echo test3 %s >> /dev/null test3 && ", now.Minute(), now.Hour(), now.Day(), int(now.Month()), workDir, testImageUri)
		expected2 := fmt.Sprintf("mv -f %s/test3-release-setting %s/test3-released && rm -f %stest3-release && ", workDir, workDir, cronPath)
		expected3 := fmt.Sprintf("rm -f %s/test3-release-processing", workDir)
		assert.Equal(t, expected1+expected2+expected3, cron)
	})

	t.Run("サービス×3・イメージ取得（GetImageListのみ）", func(t *testing.T) {
		// テスト用の ListImages の結果を生成
		digest1 := "sha256:4d2653f861f1c4cb187f1a61f97b9af7adec9ec1986d8e253052cfa60fd7372f"
		tag1 := "latest"
		imageId1 :=
			types.ImageIdentifier{
				ImageDigest: &digest1,
				ImageTag:    &tag1,
			}
		digest2 := "sha256:20b39162cb057eab7168652ab012ae3712f164bf2b4ef09e6541fca4ead3df62"
		imageId2 :=
			types.ImageIdentifier{
				ImageDigest: &digest2,
			}
		var imageIds []types.ImageIdentifier
		imageIds = append(imageIds, imageId1)
		imageIds = append(imageIds, imageId2)

		// テスト用の DescribeImages の結果を生成
		expectedTime1, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-02T05:27:02Z")
		expectedTime2, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2022-09-02T05:07:10Z")
		registryId := "000000000000"
		repositoryName := "repository33"
		size1 := float32(10017365)
		size1Int64 := int64(10017365)
		var tags1 []string
		tags1 = append(tags1, tag1)
		imageDetail1 :=
			types.ImageDetail{
				ImageDigest:      &digest1,
				ImagePushedAt:    &expectedTime1,
				ImageSizeInBytes: &size1Int64,
				ImageTags:        tags1,
				RegistryId:       &registryId,
				RepositoryName:   &repositoryName,
			}
		size2 := float32(10017367)
		size2Int64 := int64(10017367)
		imageDetail2 :=
			types.ImageDetail{
				ImageDigest:      &digest2,
				ImagePushedAt:    &expectedTime2,
				ImageSizeInBytes: &size2Int64,
				RegistryId:       &registryId,
				RepositoryName:   &repositoryName,
			}
		var imageDetails []types.ImageDetail
		imageDetails = append(imageDetails, imageDetail1)
		imageDetails = append(imageDetails, imageDetail2)

		repositoryUri := "000000000000.dkr.ecr.ap-northeast-1.amazonaws.com/repository33"
		imageList := api.GetImageList(imageIds, imageDetails, repositoryName, repositoryUri)
		assert.Equal(t, 2, len(imageList))
		assert.Equal(t, digest1, imageList[0].Digest)
		assert.Equal(t, expectedTime1, imageList[0].PushedAt)
		assert.Equal(t, repositoryName, imageList[0].RepositoryName)
		assert.Equal(t, size1, imageList[0].Size)
		assert.Equal(t, 1, len(imageList[0].Tags))
		assert.Equal(t, tag1, imageList[0].Tags[0])
		assert.Equal(t, fmt.Sprintf("%s:%s", repositoryUri, tag1), imageList[0].Uri)
		assert.Equal(t, digest2, imageList[1].Digest)
		assert.Equal(t, expectedTime2, imageList[1].PushedAt)
		assert.Equal(t, size2, imageList[1].Size)
		assert.Nil(t, imageList[1].Tags)
		assert.Equal(t, fmt.Sprintf("%s@%s", repositoryUri, digest2), imageList[1].Uri)
	})
}

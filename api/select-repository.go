package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SelectRepository struct {
    ServiceName         []string
    RepositoryMap       map[string][]Repository
    RepositoryMap2d     map[RepositoryKey]RepositoryItem
    ServiceSettingPath  string
    CronPath            string
    CronCmd             string
    CronLog             string
}

func NewSelectRepository(workDir string, cronPath string, cronCmd string, cronLog string) *SelectRepository {
    selectRepository := ReadConfig(workDir, cronPath, cronCmd, cronLog)
    return selectRepository
}

// エラーメッセージ返却用
func sendError(c *gin.Context, code int, message string) {
    selectErr := Error{
        Message: message,
    }
    c.JSON(code, selectErr)
}

// コンテナイメージ一覧の取得
func (s *SelectRepository) GetImages(c *gin.Context, serviceName ServiceName, repositoryName RepositoryName) {
    repository := s.RepositoryMap[serviceName]
    if repository == nil {
        sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
        return
    }
    repository2d := s.RepositoryMap2d[RepositoryKey{
        ServiceName: serviceName,
        RepositoryName: repositoryName,
    }]
    result, err := ImageList(repositoryName, repository2d.RegistryId, repository2d.Uri)
    if err != nil {
        sendError(c, http.StatusInternalServerError, fmt.Sprintf("%s", err))
        return
    }
    c.JSON(http.StatusOK, result)
}

// リポジトリ一覧の取得
func (s *SelectRepository) GetRepositories(c *gin.Context, serviceName ServiceName) {
    if s.RepositoryMap[serviceName] == nil {
        sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
        return
    }
    result := s.RepositoryMap[serviceName]
    c.JSON(http.StatusOK, result)
}

// コンテナサービス一覧の取得
func (s *SelectRepository) GetServices(c *gin.Context) {
    var result []Service
    for _, v := range s.ServiceName {
        service := Service{Name: v}
        result = append(result, service)
    }
    c.JSON(http.StatusOK, result)
}

// リリース設定の削除（リリース取り消し）
func (s *SelectRepository) DeleteSetting(c *gin.Context, serviceName ServiceName) {
    var err error
    if s.RepositoryMap[serviceName] == nil {
        sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
        return
    }

    // リリース処理中なら 500 Error を返す
    if CheckNowReleaseProcessing(s.ServiceSettingPath, serviceName) {
        sendError(c, http.StatusInternalServerError, "すでにリリース処理が開始されています。取り消しできません")
        return
    }

    // 設定を削除
    err = RemoveSetting(s.ServiceSettingPath, s.CronPath, serviceName)
    if err != nil {
        sendError(c, http.StatusInternalServerError, fmt.Sprintf("設定の削除が失敗しました : %s", err))
        return
    }
    c.Status(http.StatusNoContent)
}

// リリース設定の取得
func (s *SelectRepository) GetSetting(c *gin.Context, serviceName ServiceName) {
    if s.RepositoryMap[serviceName] == nil {
        sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
        return
    }
    result := ReadSetting(s.ServiceSettingPath, serviceName)

    c.JSON(http.StatusOK, result)
}

// リリース設定の生成・更新
func (s *SelectRepository) PostSetting(c *gin.Context, serviceName ServiceName) {
    var err error
    if s.RepositoryMap[serviceName] == nil {
        sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
        return
    }
    var setting Setting
    err = c.Bind(&setting)
    if err != nil {
        sendError(c, http.StatusBadRequest, fmt.Sprintf("設定項目の形式が誤っています : %s", err))
        return
    }

    // リリース処理中なら 500 Error を返す
    if CheckNowReleaseProcessing(s.ServiceSettingPath, serviceName) {
        sendError(c, http.StatusInternalServerError, "現在リリース処理中です。完了までしばらくお待ちください")
        return
    }

    // 設定を保存 or 更新
    err = UpdateSetting(s.ServiceSettingPath, s.CronPath, s.CronCmd, s.CronLog, serviceName, &setting)
    if err != nil {
        sendError(c, http.StatusInternalServerError, fmt.Sprintf("設定の保存・更新が失敗しました : %s", err))
        return
    }
    c.JSON(http.StatusOK, setting)
}

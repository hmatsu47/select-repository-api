package api

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type SelectRepository struct {
    ServiceName         []string
    RepositoryMap       map[string][]Repository
    RepositoryMap2d     map[RepositoryKey]RepositoryItem
    ServiceSettingPath  string
}

func NewSelectRepository(workDir string) *SelectRepository {
    selectRepository := ReadConfig(workDir)
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
    sort.Slice(result, func(i, j int) bool {
        return result[i].PushedAt.After(result[j].PushedAt)
    })
    c.JSON(http.StatusOK, result)
}

// リポジトリ一覧の取得
func (s *SelectRepository) GetRepositories(c *gin.Context, serviceName ServiceName) {
    if s.RepositoryMap[serviceName] == nil {
        sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
        return
    }
    result := s.RepositoryMap[serviceName]
    sort.Slice(result, func(i, j int) bool {
        return result[i].Name < result[j].Name
    })
    c.JSON(http.StatusOK, result)
}

// コンテナサービス一覧の取得
func (s *SelectRepository) GetServices(c *gin.Context) {
    var result []Service
    for _, v := range s.ServiceName {
        service := Service{Name: v}
        result = append(result, service)
    }
    sort.Slice(result, func(i, j int) bool {
        return result[i].Name < result[j].Name
    })
    c.JSON(http.StatusOK, result)
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
    if s.RepositoryMap[serviceName] == nil {
        sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
        return
    }
    var setting Setting
    err := c.Bind(&setting)
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
    uerr := UpdateSetting(s.ServiceSettingPath, serviceName, &setting)
    if uerr != nil {
        sendError(c, http.StatusInternalServerError, fmt.Sprintf("設定の保存・更新が失敗しました : %s", uerr))
        return
    }
    c.JSON(http.StatusOK, setting)
}

package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RepositoryMap struct {
	Name		string
	Uri			string
	RegistryId	string
}

type SelectRepository struct {
	ServiceName			[]string
	RepositoryMap		map[string][]RepositoryMap
	ServiceSettingPath	string
}

func NewSelect(config *Config) *SelectRepository {
	return &SelectRepository {
		ServiceName: config.ServiceName,
		RepositoryMap: config.RepositoryMap,
		ServiceSettingPath: config.ServiceSettingPath,
	}
}

// エラーメッセージ返却用
func sendError(c *gin.Context, code int, message string) {
	selectErr := Error{
		Message: message,
	}
	c.JSON(code, selectErr)
}

// コンテナイメージ一覧の取得
func (s *SelectRepository) GetImages(c *gin.Context, repositoryName RepositoryName) {
	var result []Image
	c.JSON(http.StatusOK, result)
}

// リポジトリ一覧の取得
func (s *SelectRepository) GetRepositories(c *gin.Context, serviceName ServiceName) {
	if (s.RepositoryMap[serviceName] == nil) {
		sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
		return
	}
	var result []Repository
	for _ ,v := range s.RepositoryMap[serviceName] {
		repository := Repository{
			Name: v.Name,
			Uri: v.Uri,
		}
        result = append(result, repository)
    }
	c.JSON(http.StatusOK, result)
}

// コンテナサービス一覧の取得
func (s *SelectRepository) GetServices(c *gin.Context) {
	var result []Service
    for _ ,v := range s.ServiceName {
		service := Service{Name: v}
        result = append(result, service)
    }
	c.JSON(http.StatusOK, result)
}

// リリース設定の取得
func (s *SelectRepository) GetSetting(c *gin.Context, serviceName ServiceName) {
	if (s.RepositoryMap[serviceName] == nil) {
		sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
		return
	}
	setting := NewSetting(s.ServiceSettingPath, serviceName)

	c.JSON(http.StatusOK, setting)
}

// リリース設定の生成・更新
func (s *SelectRepository) PostSetting(c *gin.Context, serviceName ServiceName) {
	if (s.RepositoryMap[serviceName] == nil) {
		sendError(c, http.StatusNotFound, fmt.Sprintf("指定されたサービスが存在しません : %s", serviceName))
		return
	}
	var setting Setting
	err := c.Bind(&setting)
	if (err != nil) {
		sendError(c, http.StatusBadRequest, fmt.Sprintf("設定項目の形式が誤っています : %s", err))
		return
	}

	// リリース処理中なら 500 Error を返す
	if (CheckNowReleaseProcessing(s.ServiceSettingPath, serviceName)) {
		sendError(c, http.StatusInternalServerError, "現在リリース処理中です。完了までしばらくお待ちください")
		return
	}

	// 設定を保存 or 更新
	uerr := UpdateSetting(s.ServiceSettingPath, serviceName, &setting)
	if (uerr != nil) {
		sendError(c, http.StatusInternalServerError, fmt.Sprintf("設定の保存・更新が失敗しました : %s", uerr))
		return
	}
	c.JSON(http.StatusOK, setting)
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SelectRepository struct {
	ServiceName			[]string
	RepositoryName		map[string][]string
	ServiceSettingPath	string
}

func NewSelect(config Config) *SelectRepository {
	return &SelectRepository {
		ServiceName: config.ServiceName,
		RepositoryName: config.RepositoryName,
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
	var result []Repository
	c.JSON(http.StatusOK, result)
}

// コンテナサービス一覧の取得
// (GET /services)
func (s *SelectRepository) GetServices(c *gin.Context) {
	var result []Service
    for _ ,v := range s.ServiceName {
		service := Service{Name: v}
        result = append(result, service)
    }
	c.JSON(http.StatusOK, result)
}

// リリース設定の取得
func (s *SelectRepository) GetSetting(c *gin.Context) {
	var result Setting
	c.JSON(http.StatusOK, result)
}

// リリース設定の生成・更新
func (s *SelectRepository) PostSetting(c *gin.Context) {
	var result Setting
	c.JSON(http.StatusOK, result)
}

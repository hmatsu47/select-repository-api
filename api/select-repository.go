package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SelectRepository struct {}

func NewSelect() *SelectRepository {
	return &SelectRepository{}
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
func (s *SelectRepository) GetRepositories(c *gin.Context) {
	var result []Repository
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

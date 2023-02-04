// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// コンテナイメージ一覧の取得
	// (GET /images/{service_name}/{repository_name})
	GetImages(c *gin.Context, serviceName ServiceName, repositoryName RepositoryName)
	// リポジトリ一覧の取得
	// (GET /repositories/{service_name})
	GetRepositories(c *gin.Context, serviceName ServiceName)
	// コンテナサービス一覧の取得
	// (GET /services)
	GetServices(c *gin.Context)

	// (DELETE /setting/{service_name})
	DeleteSetting(c *gin.Context, serviceName ServiceName)
	// リリース設定の取得
	// (GET /setting/{service_name})
	GetSetting(c *gin.Context, serviceName ServiceName)
	// リリース設定の生成・更新
	// (POST /setting/{service_name})
	PostSetting(c *gin.Context, serviceName ServiceName)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
}

type MiddlewareFunc func(c *gin.Context)

// GetImages operation middleware
func (siw *ServerInterfaceWrapper) GetImages(c *gin.Context) {

	var err error

	// ------------- Path parameter "service_name" -------------
	var serviceName ServiceName

	err = runtime.BindStyledParameter("simple", false, "service_name", c.Param("service_name"), &serviceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter service_name: %s", err)})
		return
	}

	// ------------- Path parameter "repository_name" -------------
	var repositoryName RepositoryName

	err = runtime.BindStyledParameter("simple", false, "repository_name", c.Param("repository_name"), &repositoryName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter repository_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetImages(c, serviceName, repositoryName)
}

// GetRepositories operation middleware
func (siw *ServerInterfaceWrapper) GetRepositories(c *gin.Context) {

	var err error

	// ------------- Path parameter "service_name" -------------
	var serviceName ServiceName

	err = runtime.BindStyledParameter("simple", false, "service_name", c.Param("service_name"), &serviceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter service_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetRepositories(c, serviceName)
}

// GetServices operation middleware
func (siw *ServerInterfaceWrapper) GetServices(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetServices(c)
}

// DeleteSetting operation middleware
func (siw *ServerInterfaceWrapper) DeleteSetting(c *gin.Context) {

	var err error

	// ------------- Path parameter "service_name" -------------
	var serviceName ServiceName

	err = runtime.BindStyledParameter("simple", false, "service_name", c.Param("service_name"), &serviceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter service_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.DeleteSetting(c, serviceName)
}

// GetSetting operation middleware
func (siw *ServerInterfaceWrapper) GetSetting(c *gin.Context) {

	var err error

	// ------------- Path parameter "service_name" -------------
	var serviceName ServiceName

	err = runtime.BindStyledParameter("simple", false, "service_name", c.Param("service_name"), &serviceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter service_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetSetting(c, serviceName)
}

// PostSetting operation middleware
func (siw *ServerInterfaceWrapper) PostSetting(c *gin.Context) {

	var err error

	// ------------- Path parameter "service_name" -------------
	var serviceName ServiceName

	err = runtime.BindStyledParameter("simple", false, "service_name", c.Param("service_name"), &serviceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter service_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.PostSetting(c, serviceName)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL     string
	Middlewares []MiddlewareFunc
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router *gin.Engine, si ServerInterface) *gin.Engine {
	return RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router *gin.Engine, si ServerInterface, options GinServerOptions) *gin.Engine {
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
	}

	router.GET(options.BaseURL+"/images/:service_name/:repository_name", wrapper.GetImages)

	router.GET(options.BaseURL+"/repositories/:service_name", wrapper.GetRepositories)

	router.GET(options.BaseURL+"/services", wrapper.GetServices)

	router.DELETE(options.BaseURL+"/setting/:service_name", wrapper.DeleteSetting)

	router.GET(options.BaseURL+"/setting/:service_name", wrapper.GetSetting)

	router.POST(options.BaseURL+"/setting/:service_name", wrapper.PostSetting)

	return router
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8xY3W7cRBR+lWrg0om3TaOmewWECCLRqkq4C1E06z27ntb2mJnZJEtkidgSpJFQK1Gl",
	"4qcgoIIQICAVUKJGvIzZtL3qK6AZ27ve9TjrpInE3a49c873fWfOj2cDWdT1qQee4Ki+gXzMsAsCmPrH",
	"wKecCMq6Kx52QT5qArcY8QWhHqqjONqLo0dxeBBHW3G017v/GTIQkW98LGxkoGRbwZCBGHzYIQyaqC5Y",
	"BwzELRtcLD2Iri+3cMGI10ZBYCAObJVYUIYh/CuOjuLo8zg8LAUwZOI03oNkMXDxFm0SULJwEIJ47YXk",
	"uXxiUU+Ap35i33eIhSU48zaXCDdy5l9n0EJ19Jo5kN1M3nJzMTGbONXovCdphofPd3/r7X8p/4a/x+Fu",
	"HB4q8b+Oo0/j8AeUIOY+9XiCFhijbCF9cm5g56RVLdRwN45+VhH5Lo6iOHyqYB/E0a8K6qM4eqJ+9AEb",
	"iLi4DfxMGIkAl48DOy/tS0dpdDFjuKsH/0TCiz6Jo+04fKwoSPD/Hnz8/MefTqLQP+HkIoks9POoEpvh",
	"7BxPIs2SCySwmHioFotBXleBnubkWOTrrnNhOVmGL8hKjBIoyR1NISvLnO+lnegXZCCfUR+YSAuRC5zL",
	"k62tmoMat9RfuGwgQYQjVyYg+mGgjdtgCWSg9QkuqO+Qtq2UI01UR9c7vjMztVabvmbZWBlPMkpDQZ8/",
	"5RSapJ3W0BEGBvI73IbmClZvW5S58hdqYgETgqgyXtiiaViFNZx8lH/hddwGMHUgcZsPHeTC1uEja6AO",
	"I+O1l4tS46lzQ9MPUx3yrHPBSuTWBYt4ApiHHVRvYYeDPn4zNdey27zbvu7NrCmAuUIyrqmXx65U4kq6",
	"pMTl2hzTHLAz08XTtus2plmTrjVmlN+s7Jw4PJyaqI5QjkrmtFqO3bkCU+ur5E7ryuUWS0EntUcfoJHK",
	"UwZdtdYVfUAMRPgKAwcwlwz67xuUOoC9JKHU21Mk4YgoeQ9D2iTcdNrAOnZ9R8EPNDFXpZh4LZrVeGwp",
	"cOBi4qA6sl0seOfqtTfa8sGkRd3BGPguYbSLeefSDbnGJhyrE6i2CeHzumm2ibA7DbnNzCyhQu2fm10Y",
	"yZIXmwfH2988e7D75q15ZCCHWJC2odT1jfn3q/gyOThgiYlBhZjAvm82HNowXcwFMPO9+dm5m4tzqiCl",
	"chY2IQOtAuMJ2suTNbmY+uBhn6A6mpqsTdbkUcHCVqfETCYwcyM/JwfmxkidCuTaNojqhT9t3Jv7vXs7",
	"vX8eIoWCqUY8L0/9OyBUbeNoZGq9UquV9eb+OnNkblRhauGOI8ZvHZ6KVYPuuC6W9fAUZJKWsZRkGVqW",
	"TSv3AbWkBzFYYo72gcAYu2XoSyaQLs387DkSwvKAaWfDk+K0kPNypmhpR+Rzi9k4Plmocjly+nhpxM9m",
	"5oqpUZhpM4gfeDrRFzPrZxG8MM5fTIKUUsqpnkJBy0Gimar+mrPaBAcEVGp5m/u9u9svvnj88mgr/7Z3",
	"bycOt4//3oo3H748ulsQ9W3lYdB/RmS9WvR9k16aTT8lXkXAwDghEwvUStOwFHmlAzH8kXSuyVdKYXAG",
	"EuTLuqb+iploIJ/yquI+e/Dt8db9OHp6/NWfxzt/FFS+RfmQzNn1T7dcpdwNkTlyPRT87yNV1KMQsiC7",
	"h8vCMxhk6qbpUAs7NuWiPlWr1VQ4UgtV54TBiEbSu5oTdw4Xe9094zgLuapVuCXU7S3olt+V3hEsB/8F",
	"AAD//3W38OZPFQAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}

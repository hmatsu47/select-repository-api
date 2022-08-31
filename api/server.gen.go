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
	// (GET /images/{repository_name})
	GetImages(c *gin.Context, repositoryName RepositoryName)
	// リポジトリ一覧の取得
	// (GET /repositories/{service_name})
	GetRepositories(c *gin.Context, serviceName ServiceName)
	// コンテナサービス一覧の取得
	// (GET /services)
	GetServices(c *gin.Context)
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

	siw.Handler.GetImages(c, repositoryName)
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

	router.GET(options.BaseURL+"/images/:repository_name", wrapper.GetImages)

	router.GET(options.BaseURL+"/repositories/:service_name", wrapper.GetRepositories)

	router.GET(options.BaseURL+"/services", wrapper.GetServices)

	router.GET(options.BaseURL+"/setting/:service_name", wrapper.GetSetting)

	router.POST(options.BaseURL+"/setting/:service_name", wrapper.PostSetting)

	return router
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8xY3W7cRBR+FXTg0om3CVHTvQKqCCJRVCXchSia9c6up7U9ZmacZIksUVuCqhKiElWR",
	"gIL4EYQABakgJSriZUzS8hZoZuxdez3en5BI3O16Zs75znd+Zw7AoX5IAxwIDu0DCBFDPhaYqX8Mh5QT",
	"QdlgJ0A+lp+6mDuMhILQANqQpUdZ+ihLjrP0bpYend7/CCwgciVEwgUL9LGaIAsYfjciDHehLViELeCO",
	"i30kNYhBKI9wwUjQhzi2gGO2SxzchCH5I0v/zNJPsuSkEUBFxDzaY70Zc/Ea7RKsaOFYCBL0N/R3+cWh",
	"gcCB+onC0CMOkuDsW1wiPCiJf4nhHrThRXtEu61Xub2pxWqlBp6PpJnJyfPDX04ffyb/Jr9myWGWnCjy",
	"v8jSD7PkW9CIeUgDrtFixijbyL9cGNg1KdUINTnM0h+VR77O0jRLnirYx1n6s4L6KEufqB9DwBYQH/Ux",
	"PxdGIrDPp4Fdl/Kloty7iDE0MIN/IuGlH2TpvSz5Tpkgwf99/P7z73+YZMIwwsllGrIxzKOZrKlm53Qj",
	"8iy5RAM2tYbZfDHK61mg5zk5Ffm+711aTjbhi4sSowjSuWMoZE2Z842Uk/4EFoSMhpiJvBD5mHMZ2caq",
	"OapxW8ON2xYIIjy5U4MYuoF2bmFHgAX7C1zQ0CN9VzFHutCGa1HorS7vtVauOi5SwnVGGUww50+zCV3S",
	"z2vomAUWhBF3cXcHqdUeZb78BV0k8IIgqozXjhgaVm0PJ++VF4LI72CmAhL1eSWQa0erIWtBxMh07uWm",
	"XHiu3DL0w5yHstUlZ2m6Tc4igcAsQB60e8jj2Oy/1ZbvuH0+6F8LVvcUwFIhmdbUm33XSPFMvOSGy70l",
	"S0vAzm0uWnF9v7PCunSvs6r0FmVn4vAwt6Emg0qmFEpny7HbS3h5f5fc7i1d6bEctK49ZgeNVZ4m6Kq1",
	"7pgdYgHhOwx7GHFpwXC9Q6mHUaATSq3OkYRjpJQ1VLjRtpm4wfvIDz0FPzb4XJViEvRoUeORo8BhHxEP",
	"2uD6SPDo5auv9OWHRYf6ozHwDcLoAPHohRtyj0s4UhGojgkR8rZt94lwo448ZheSoFb7165vjGXJP3eO",
	"z+59+ezB4as318ECjzg4b0O56hvrb8+iy+bYw45YGFWIBRSGdsejHdtHXGBmv7l+fe2tzTVVkHI6a4fA",
	"gl3MuEZ7ZbElN9MQBygk0IblxdZiS4YKEq6KEltPYPbBWGGK5WIfi9krfd6p7zw+/fjh6V+fglLLVOdd",
	"l2H+OhaqmHEYG1OXWq2mZjzcZ48NisovPRR5YvrR6hisOnLk+0gWwDmM0T1iS6cVbMsuVboxbZlBjLbY",
	"44U/liLs8vBoH5TvKhMcYBzuJvG+UdJyLvaNM+6F+WCaPQX1pSCfn//KPVCTXwy9M4Z6bSgtIL4TmEjf",
	"LKSfh/DaPH45Ad9oUon1HApsx5ozVb7nitVqw5oUqKPmcB7KqveACw3PRhNGLGnk26a+9R9j1YKQ8lnJ",
	"ffbgq7O797P06dnnv589/K3G8k3KKzQXLxyDZpZKjyD22AtI/L/3VJ2Pmsvi4qmpcM+oV7dt26MO8lzK",
	"RXu51Wopd+QSZu2MoymE5M8RE09Wy6HpKW2ahFJe1x7CTGdrvJVP5dfg7fjfAAAA////0xBlMhQAAA==",
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

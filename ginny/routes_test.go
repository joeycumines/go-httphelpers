package ginny

import (
	"testing"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"github.com/go-test/deep"
	"strings"
)

func TestRouteDefinition_String(t *testing.T) {
	d := RouteDefinition{
		Method:       "some_method",
		RelativePath: "some_path",
		Handlers:     []gin.HandlerFunc{nil, nil, nil},
	}
	if s := d.String(); s != "Route<method, relativePath, handlerCount> = (some_method, some_path, 3)" {
		t.Error("unexpected value:", s)
	}
}

func TestRouterDefinition_String(t *testing.T) {
	d := RouterDefinition{
		RelativePath: "some_path",
		Handlers:     []gin.HandlerFunc{nil, nil, nil},
		Routes:       []RouteDefinition{{}, {}, {}, {}},
		Routers:      []RouterDefinition{{}, {}, {}, {}, {}},
	}
	if s := d.String(); s != "Router<relativePath, handlerCount, routeCount, routerCount> = (some_path, 3, 4, 5)" {
		t.Error("unexpected value:", s)
	}
}

func TestRouter_Apply_http(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	var (
		out    []int
		engine        = gin.New()
		router Router = func() (relativePath string, handlers []gin.HandlerFunc, routes []Route, routers []Router, err error) {
			relativePath = "/1.1"

			handlers = append(handlers, func(context *gin.Context) {
				out = append(out, 1)
			})
			handlers = append(handlers, func(context *gin.Context) {
				out = append(out, 2)
			})

			routes = append(routes, func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
				method = "GET"
				relativePath = "/2.1"
				handlers = append(handlers, func(context *gin.Context) {
					out = append(out, 3)
				})
				return
			})

			routes = append(routes, func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
				method = "GET"
				relativePath = "/2.2"
				handlers = append(handlers, func(context *gin.Context) {
					out = append(out, -4)
				})
				return
			})

			routers = append(routers, func() (relativePath string, handlers []gin.HandlerFunc, routes []Route, routers []Router, err error) {
				relativePath = "/2.2"
				handlers = append(handlers, func(context *gin.Context) {
					out = append(out, -5)
				})
				routes = append(routes, func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
					method = "GET"
					relativePath = "/3.1"
					handlers = append(handlers, func(context *gin.Context) {
						out = append(out, 4)
					})
					return
				})
				routers = append(routers, func() (relativePath string, handlers []gin.HandlerFunc, routes []Route, routers []Router, err error) {
					relativePath = "/3.2"
					routes = append(routes, func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
						method = "POST"
						relativePath = "/4.1"
						handlers = append(handlers, func(context *gin.Context) {
							out = append(out, 5)
						})
						return
					})
					return
				})
				return
			})

			routes = append(routes, func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
				method = "HEAD"
				relativePath = "/2.3"
				handlers = append(handlers, func(context *gin.Context) {
					out = append(out, 6)
				})
				handlers = append(handlers, func(context *gin.Context) {
					out = append(out, 7)
				})
				return
			})

			return
		}
	)

	if routerGroup, err := router.Apply(engine); err != nil || routerGroup == nil {
		t.Fatalf("unexpected values (%+v, %+v)", routerGroup, err)
	}

	runTest := func(method, url string, expected []int) {
		out = make([]int, 0)
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
		if diff := deep.Equal(out, expected); diff != nil {
			t.Errorf("%s %s unexpected out %v had diff:\n  - %s", method, url, out, strings.Join(diff, "\n  - "))
		}
	}

	runTest("GET", "/1.1", []int{})
	runTest("GET", "/1.1/2.1", []int{1, 2, 3})
	runTest("HEAD", "/1.1/2.1", []int{})
	runTest("GET", "/1.1/2.2", []int{1, 2, -4})
	runTest("GET", "/1.1/2.2/3.1", []int{1, 2, -5, 4})
	runTest("POST", "/1.1/2.2/3.1", []int{})
	runTest("POST", "/1.1/2.2/3.2/4.1", []int{1, 2, -5, 5})
	runTest("HEAD", "/1.1/2.3", []int{1, 2, 6, 7})
}

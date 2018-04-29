package server

import (
	"github.com/joeycumines/httphelpers.go/ginny"
	"github.com/gin-gonic/gin"
)

type Root struct {
	MovieStore MovieStore
}

func (r *Root) Router() ginny.Router {
	return func() (relativePath string, handlers []gin.HandlerFunc, routes []ginny.Route, routerGroups []ginny.Router, err error) {
		movies := Movies{
			MovieStore: r.MovieStore,
		}
		routerGroups = append(routerGroups, movies.RouterGroup())

		return
	}
}

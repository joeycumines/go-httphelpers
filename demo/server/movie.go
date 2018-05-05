package server

import (
	"github.com/joeycumines/go-httphelpers/ginny"
	"github.com/gin-gonic/gin"
	"net/http"
	"errors"
	"strconv"
)

const (
	paramMovieID = `movie_id`
)

type Movie struct {
	MovieStore MovieStore
}

func (m *Movie) RouterGroup() ginny.Router {
	return func() (relativePath string, handlers []gin.HandlerFunc, routes []ginny.Route, routerGroups []ginny.Router, err error) {
		relativePath = "/:" + paramMovieID

		routes = append(routes, m.Get())

		return
	}
}

func (m *Movie) Get() ginny.Route {
	return func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
		if m == nil || m.MovieStore == nil {
			err = errors.New("server.Movie.Get movie store must be set")
			return
		}

		method = "GET"
		handlers = append(handlers, func(c *gin.Context) {
			if movieID, ok := c.Params.Get(paramMovieID); !ok {
				panic("expected a movie id")
			} else if ID, err := strconv.Atoi(movieID); err != nil {
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"message": "movie id must be an int",
					"error":   err.Error(),
					"code":    http.StatusBadRequest,
				})
			} else if movie, ok, err := m.MovieStore.Load(ID); err != nil {
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"message": "failed to get movies",
					"error":   err.Error(),
					"code":    http.StatusInternalServerError,
				})
				return
			} else if !ok || movie == nil {
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"message": "movie not found",
					"error":   "movie does not exist",
					"code":    http.StatusNotFound,
				})
				return
			} else {
				c.JSON(http.StatusOK, movie)
			}
		})

		return
	}
}

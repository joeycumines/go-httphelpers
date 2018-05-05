package server

import (
	"github.com/joeycumines/go-httphelpers/ginny"
	"github.com/gin-gonic/gin"
	"errors"
	"net/http"
	"github.com/joeycumines/go-httphelpers/demo/server/models"
	"encoding/json"
)

type Movies struct {
	MovieStore MovieStore
}

func (m *Movies) RouterGroup() ginny.Router {
	return func() (relativePath string, handlers []gin.HandlerFunc, routes []ginny.Route, routerGroups []ginny.Router, err error) {
		relativePath = "/movies"

		routes = append(routes, m.Get())
		routes = append(routes, m.Post())

		movie := Movie{
			MovieStore: m.MovieStore,
		}
		routerGroups = append(routerGroups, movie.RouterGroup())

		return
	}
}

func (m *Movies) Get() ginny.Route {
	return func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
		if m == nil || m.MovieStore == nil {
			err = errors.New("server.Movies.Get movie store must be set")
			return
		}

		method = "GET"
		handlers = append(handlers, func(c *gin.Context) {
			movies := make([]*models.Movie, 0)

			if err := m.MovieStore.Range(func(ID int, movie *models.Movie) bool {
				movies = append(movies, movie)
				return true
			}); err != nil {
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"message": "failed to get movies",
					"error":   err.Error(),
					"code":    http.StatusInternalServerError,
				})
				return
			}

			c.JSON(http.StatusOK, movies)
		})

		return
	}
}

func (m *Movies) Post() ginny.Route {
	return func() (method string, relativePath string, handlers []gin.HandlerFunc, err error) {
		if m == nil || m.MovieStore == nil {
			err = errors.New("server.Movies.Post movie store must be set")
			return
		}

		method = "POST"
		handlers = append(handlers, func(c *gin.Context) {
			movie := new(models.Movie)

			if err := json.NewDecoder(c.Request.Body).Decode(movie); err != nil {
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"message": "failed to parse request",
					"error":   err.Error(),
					"code":    http.StatusBadRequest,
				})
				return
			}

			if stored, err := m.MovieStore.Create(movie); err != nil {
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"message": "failed to store movie",
					"error":   err.Error(),
					"code":    http.StatusBadRequest,
				})
				return
			} else {
				c.JSON(http.StatusOK, stored)
			}
		})

		return
	}
}

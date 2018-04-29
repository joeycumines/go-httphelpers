package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joeycumines/httphelpers.go/demo/server"
	"fmt"
	"os"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	root := server.Root{
		MovieStore: server.NewMemMovieStore(),
	}

	if _, err := root.Router().Apply(r); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	r.Run()
}

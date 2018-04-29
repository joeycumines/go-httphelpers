/*
   Copyright 2018 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */

// Package ginny holds common boilerplate for working with the gin-gonic/gin http server.
package ginny

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"errors"
)

type (
	// Route models a closure or method that can be applied to something implementing gin.IRoutes, which may also
	// perform validation of any configuration or input, and should not return without an error unless the route
	// is ready to be served.
	// For informative debugging, an error case should still return a method and relative path.
	Route func() (method string, relativePath string, handlers []gin.HandlerFunc, err error)

	// Router models a closure or method that can be applied to something implementing gin.IRouter, which
	// may also perform validation of any configuration or input, and should not return without an error unless the
	// group of routes are ready to be served.
	// For informative debugging, an error case should still return a relative path.
	Router func() (relativePath string, handlers []gin.HandlerFunc, routes []Route, routers []Router, err error)

	// RouteDefinition is a resolved Route that is ready to be applied.
	RouteDefinition struct {
		Method       string
		RelativePath string
		Handlers     []gin.HandlerFunc
	}

	// RouterDefinition is a resolved Router that is ready to be applied.
	RouterDefinition struct {
		RelativePath string
		Handlers     []gin.HandlerFunc
		Routes       []RouteDefinition
		Routers      []RouterDefinition
	}
)

func (d RouteDefinition) String() string {
	return fmt.Sprintf(
		`Route<method, relativePath, handlerCount> = (%s, %s, %d)`,
		d.Method,
		d.RelativePath,
		len(d.Handlers),
	)
}

func (d RouterDefinition) String() string {
	return fmt.Sprintf(
		`Router<relativePath, handlerCount, routeCount, routerCount> = (%s, %d, %d, %d)`,
		d.RelativePath,
		len(d.Handlers),
		len(d.Routes),
		len(d.Routers),
	)
}

// Resolve calls a Route, validates the output, and returns a RouteDefinition, possibly with an error.
// Any nil handlers will cause an error.
func (r Route) Resolve() (definition RouteDefinition, err error) {
	if r == nil {
		err = errors.New("ginny.Route.Resolve nil route")
		return
	}
	definition.Method, definition.RelativePath, definition.Handlers, err = r()
	if err != nil {
		err = fmt.Errorf("ginny.Route.Resolve route error (%s): %s",
			definition.String(),
			err.Error(),
		)
		return
	}
	// validate handlers
	for i, handler := range definition.Handlers {
		if handler == nil {
			err = fmt.Errorf("ginny.Route.Resolve nil handler at index %d (%s)",
				i,
				definition.String(),
			)
			return
		}
	}
	// success
	return
}

// Resolve calls a Router, validates the output, and returns a RouterDefinition, possibly with an error.
// Any nil handlers, routes, or routers will also cause an error, as will any errors resolving these (recursive).
func (r Router) Resolve() (definition RouterDefinition, err error) {
	if r == nil {
		err = errors.New("ginny.Router.Resolve nil router")
		return
	}
	// these ones need to be resolved themselves
	var (
		routes  []Route
		routers []Router
	)
	// load what we can into the definition
	definition.RelativePath, definition.Handlers, routes, routers, err = r()
	// initialise the route and group slices in the definition so the count is correct
	definition.Routes = make([]RouteDefinition, len(routes))
	definition.Routers = make([]RouterDefinition, len(routers))
	// handle router error case
	if err != nil {
		err = fmt.Errorf("ginny.Router.Resolve router error (%s): %s",
			definition.String(),
			err.Error(),
		)
		return
	}
	// validate handlers
	for i, handler := range definition.Handlers {
		if handler == nil {
			err = fmt.Errorf("ginny.Router.Resolve nil handler at index %d (%s)",
				i,
				definition.String(),
			)
			return
		}
	}
	// resolve all the routes (nil values get handled by Route.Resolve)
	for i, route := range routes {
		definition.Routes[i], err = route.Resolve()
		if err != nil {
			err = fmt.Errorf("ginny.Router.Resolve route error at index %d (%s): %s",
				i,
				definition.String(),
				err.Error(),
			)
			return
		}
	}
	// resolve all the routers (nil values get handled by Router.Resolve)
	for i, router := range routers {
		definition.Routers[i], err = router.Resolve()
		if err != nil {
			err = fmt.Errorf("ginny.Router.Resolve router error at index %d (%s): %s",
				i,
				definition.String(),
				err.Error(),
			)
			return
		}
	}
	// success
	return
}

// Apply will register a new route using gin.IRoutes.Handle and return the result, or return an error, note that
// the definition itself is not validated.
func (d RouteDefinition) Apply(target gin.IRoutes) (gin.IRoutes, error) {
	if target == nil {
		return nil, fmt.Errorf(
			"ginny.RouteDefinition.Apply nil target (%s)",
			d.String(),
		)
	}
	// add the handlers
	routes := target.Handle(d.Method, d.RelativePath, d.Handlers...)
	// success
	return routes, nil
}

// Apply will register a new router group, using gin.IRouter.Group, and return the result, after applying all nested
// Router and Router instances, note that the definition itself is not validated.
func (d RouterDefinition) Apply(target gin.IRouter) (*gin.RouterGroup, error) {
	if target == nil {
		return nil, fmt.Errorf(
			"ginny.RouterDefinition.Apply nil target (%s)",
			d.String(),
		)
	}
	// instance a new group, adding the handlers
	group := target.Group(d.RelativePath, d.Handlers...)
	// apply all routes
	for i, route := range d.Routes {
		if _, err := route.Apply(group); err != nil {
			return group, fmt.Errorf(
				"ginny.RouterDefinition.Apply error applying route at index %d (%s): %s",
				i,
				d.String(),
				err.Error(),
			)
		}
	}
	// apply all nested routers
	for i, router := range d.Routers {
		if _, err := router.Apply(group); err != nil {
			return group, fmt.Errorf(
				"ginny.RouterDefinition.Apply error applying router at index %d (%s): %s",
				i,
				d.String(),
				err.Error(),
			)
		}
	}
	// success
	return group, nil
}

// Apply is a convenience method that resolves the definition of the route, before applying it.
func (r Route) Apply(target gin.IRoutes) (gin.IRoutes, error) {
	if definition, err := r.Resolve(); err != nil {
		return nil, fmt.Errorf("ginny.Route.Apply resolve error: %s", err.Error())
	} else if routes, err := definition.Apply(target); err != nil {
		return routes, fmt.Errorf("ginny.Route.Apply apply error: %s", err.Error())
	} else {
		return routes, nil
	}
}

// Apply is a convenience method that resolves the definition of the router, before applying it.
func (r Router) Apply(target gin.IRouter) (*gin.RouterGroup, error) {
	if definition, err := r.Resolve(); err != nil {
		return nil, fmt.Errorf("ginny.Router.Apply resolve error: %s", err.Error())
	} else if routes, err := definition.Apply(target); err != nil {
		return routes, fmt.Errorf("ginny.Router.Apply apply error: %s", err.Error())
	} else {
		return routes, nil
	}
}

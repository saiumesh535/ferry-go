package ferry

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Middleware/Route Handler
type handler func(ctx *Ctx) error

type router struct {
	routerPath string
	regexPath  string
	handler    handler
}

type group struct {
	path                   string
	ferry                  *Ferry
	middleware             []handler
	middlewareCurrentIndex int
}

var (
	get    = "GET"
	post   = "POST"
	put    = "PUT"
	delete = "DELETE"
)

var routerRegexReplace = "[a-zA-Z0-9_-]*"

func (g *group) addRoute(method, path string, h handler) {
	groupPath := fmt.Sprintf("%s%s", g.path, path)
	pathWithRegex := findAndReplace(groupPath)
	g.ferry.routerMap[method] = append(g.ferry.routerMap[method], router{
		routerPath: groupPath,
		regexPath:  pathWithRegex,
		handler:    h,
	})
}

// Get method of ferry
func (g *group) Get(path string, h handler) {
	g.addRoute(get, path, h)
}

// Post method of ferry
func (g *group) Post(path string, h handler) {
	g.addRoute(post, path, h)
}

// Put method of ferry
func (g *group) Put(path string, h handler) {
	g.addRoute(put, path, h)
}

// Delete method of ferry
func (g *group) Delete(path string, h handler) {
	g.addRoute(delete, path, h)
}

func (g *group) Use(h handler) {
	g.ferry.groupMiddlewareMap[g.path] = append(g.ferry.groupMiddlewareMap[g.path], h)
}

// Group method
func (g *group) Group(path string) *group {
	return &group{
		path:       fmt.Sprintf("%s%s", g.path, path),
		ferry:      g.ferry,
		middleware: []handler{},
	}
}

// handler 404
func handle404(ferry *Ferry, ctx *Ctx) {
	if ferry.urlNotFoundHandler != nil {
		if err := ferry.urlNotFoundHandler(ctx); err != nil {
			handlerRouterError(err, ctx.Writer)
		}
		return
	}
	ctx.Writer.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprint(ctx.Writer, "Not found")
}

func handlerRouterError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}

func handleRouting(ferry *Ferry, ctx *Ctx) {
	// first get handler by method
	routesByMethod := ferry.routerMap[ctx.Request.Method]
	if routesByMethod != nil {
		groupLevelMiddleware(ctx, ferry, routesByMethod)
	} else {
		// run 404
		handle404(ferry, ctx)
	}

}

// Finds wild card in URL and replace them with a regex for,
// ex if path is /auth/:name -> /auth/[a-zA-Z0-9]*
// ex if path is /auth/name -> /auth/name
func findAndReplace(path string) string {
	if !strings.Contains(path, ":") {
		return fmt.Sprintf("%s%s%s", "^", path, "$")
	}
	result := ""
	slitted := strings.Split(path, "/")
	for _, v := range slitted {
		if v == "" {
			continue
		}
		if strings.Contains(v, ":") {
			result = fmt.Sprintf("%s/%s", result, routerRegexReplace)
			continue
		}
		result = fmt.Sprintf("%s/%s", result, v)
	}
	// replace slashes
	result = strings.ReplaceAll(result, "/", "\\/")
	result = fmt.Sprintf("%s%s%s", "^", result, "$")
	return result
}

// calls actual handler
func handleRouter(ctx *Ctx, ferry *Ferry, routers []router) {
	urlPath := ctx.Request.URL.Path
	queryPath := ctx.Request.URL.RawQuery
	for _, route := range routers {
		match, _ := regexp.MatchString(route.regexPath, urlPath)
		if match {
			ctx.routerPath = route.routerPath
			ctx.queryPath = queryPath
			if err := route.handler(ctx); err != nil {
				handlerRouterError(err, ctx.Writer)
			}
			return
		}
	}
	handle404(ferry, ctx)
}

// routerPath /auth/:name
// requestPath /auth/madhuri
// paramName name
// returns madhuri
func extractParamFromPath(routerPath, requestPath, paramName string) string {
	routerSplit := strings.Split(routerPath, "/")
	requestSplit := strings.Split(requestPath, "/")
	if len(routerSplit) != len(requestSplit) {
		return ""
	}
	paramWithWildCard := fmt.Sprintf(":%s", paramName)
	for k, v := range routerSplit {
		if v == paramWithWildCard {
			return requestSplit[k]
		}
	}
	return ""
}

// routerPath /auth/:name/:age
// requestPath /auth/madhuri/32
// returns { name: madhuri, age: 32 }
func getParamsFromPath(routerPath, requestPath string) map[string]string {
	paramsMap := map[string]string{}
	routerSplit := strings.Split(routerPath, "/")
	requestSplit := strings.Split(requestPath, "/")
	for k, v := range routerSplit {
		if strings.Contains(v, ":") {
			key := strings.ReplaceAll(v, ":", "")
			paramsMap[key] = requestSplit[k]
		}
	}
	return paramsMap
}

func getAllQueryParams(querypath string) map[string]string {
	queryParamsMap := map[string]string{}
	params := strings.Split(querypath, "&")
	for _, v := range params {
		if strings.Contains(v, "=") {
			pair := strings.Split(v, "=")
			queryParamsMap[pair[0]] = pair[1]
		}
	}
	return queryParamsMap
}

func getQueryParam(querypath string, name string) string {
	params := strings.Split(querypath, "&")
	for _, v := range params {
		if strings.Contains(v, "=") {
			pair := strings.Split(v, "=")
			if pair[0] == name {
				return pair[1]
			}
		}
	}
	return ""
}

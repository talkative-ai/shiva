package muxlite

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Method string

const (
	MethodGet   Method = "GET"
	MethodPost         = "POST"
	MethodPatch        = "PATCH"
	MethodAll          = "ALL"
)

type PathNode struct {
	Value    string
	Children map[string]*PathNode
	Routes   map[Method]*Route
}

type Router struct {
	routes map[string]*PathNode
}

type Route struct {
	path       string
	handler    *http.HandlerFunc
	methods    []Method
	parentNode *PathNode
}

func trim(s string, char byte) string {
	for s[0] == char {
		s = s[1:]
	}
	for s[len(s)-1] == char {
		s = s[:len(s)-1]
	}

	return s
}

var reqCount int
var varStore map[int]map[string]string

func prune(s string, char byte) string {
	var prev byte
	for i := 0; i < len(s); i++ {
		if s[i] == prev && s[i] == char {
			s = s[:i] + s[i+1:]
			i--
		}
		prev = s[i]
	}
	return s
}

func (r *Route) Methods(methods ...Method) {
	for routeMethod, route := range r.parentNode.Routes {
		if route == r {
			r.parentNode.Routes[routeMethod] = nil
		}
	}
	for _, method := range methods {
		r.parentNode.Routes[method] = r
	}
}

func parseExpr(s string) (v string, expr string, err error) {
	if len(s) <= 2 {
		return "", "", fmt.Errorf("String too small")
	}

	if s[0] == '{' && s[len(s)-1] == '}' {
		s = s[1 : len(s)-1]
	} else {
		return "", "", fmt.Errorf("Not an expression")
	}

	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Not an expression")
	}

	v = parts[0]
	expr = parts[1]
	err = nil

	return
}

func (r *Router) Handle(path string, handler http.HandlerFunc) *Route {
	newRoute := &Route{
		path:    path,
		handler: &handler,
		methods: []Method{},
	}

	path = trim(prune(path, '/'), '/')
	slugs := strings.Split(path, "/")

	var currentNode *PathNode
	currentNodeMap := r.routes
	for _, slug := range slugs {
		if _, exist := currentNodeMap[slug]; !exist {
			newNode := &PathNode{
				Value:    slug,
				Children: map[string]*PathNode{},
			}
			currentNodeMap[slug] = newNode
		}

		currentNode = currentNodeMap[slug]
		currentNodeMap = currentNodeMap[slug].Children
	}

	newRoute.parentNode = currentNode

	if currentNode.Routes == nil {
		currentNode.Routes = map[Method]*Route{}
	}

	currentNode.Routes[MethodAll] = newRoute

	return newRoute
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqCount++
	path := trim(prune(req.URL.Path, '/'), '/')
	slugs := strings.Split(path, "/")
	var currentNode *PathNode
	currentNodeMap := r.routes
	success := false

	vars := map[string]string{}
	for _, slug := range slugs {
		success = false
		nextKey := slug
		for k := range currentNodeMap {

			v, expr, err := parseExpr(k)

			if err == nil {
				r, _ := regexp.Compile(expr)
				str := r.FindString(slug)
				if str != "" {
					vars[v] = str
					nextKey = k
					success = true
					break
				}
			}

			if err != nil && slug == k {
				success = true
				break
			}
		}
		if success {
			currentNode = currentNodeMap[nextKey]
			currentNodeMap = currentNode.Children
		}
		if !success {
			break
		}
	}

	if success {
		for method, route := range currentNode.Routes {
			if (method == Method(req.Method) || method == "ALL") && route != nil {
				varStore[reqCount] = vars
				defer func() {
					delete(varStore, reqCount)
				}()
				req.Header.Set("-muxlite-req", strconv.Itoa(reqCount))
				(*route.handler)(w, req)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
	return
}

func Vars(r *http.Request) map[string]string {
	reqNum, _ := strconv.Atoi(r.Header.Get("-muxlite-req"))
	return varStore[reqNum]
}

func NewRouter() *Router {
	reqCount = 0
	varStore = map[int]map[string]string{}
	return &Router{
		routes: map[string]*PathNode{},
	}
}

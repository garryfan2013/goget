package route

import (
	"github.com/garryfan2013/goget/rest/handler"
)

func init() {
	register(&Route{
		Name:    "staticIndex",
		Method:  "GET",
		Pattern: "/",
		Handler: handler.StaticFileHandler,
	})

	register(&Route{
		Name:    "staticJs",
		Method:  "GET",
		Pattern: "/static/js/{file}",
		Handler: handler.StaticFileHandler,
	})

	register(&Route{
		Name:    "staticCss",
		Method:  "GET",
		Pattern: "/static/css/{file}",
		Handler: handler.StaticFileHandler,
	})

	register(&Route{
		Name:    "staticFonts",
		Method:  "GET",
		Pattern: "/static/fonts/{file}",
		Handler: handler.StaticFileHandler,
	})

	register(&Route{
		Name:    "staticImg",
		Method:  "GET",
		Pattern: "/static/img/{file}",
		Handler: handler.StaticFileHandler,
	})
}

package route

import (
	"github.com/garryfan2013/goget/rest/handler"
)

func init() {
	registerJobRoute()
}

func registerJobRoute() {
	register(&Route{
		Name:    "GetJob",
		Method:  "GET",
		Pattern: "/job/{id}",
		Handler: handler.GetJobHandler,
	})

	register(&Route{
		Name:    "GetJobs",
		Method:  "GET",
		Pattern: "/jobs",
		Handler: handler.GetJobsHandler,
	})

	register(&Route{
		Name:    "PostJob",
		Method:  "POST",
		Pattern: "/job",
		Handler: handler.PostJobHandler,
	})

	register(&Route{
		Name:    "PostJobAction",
		Method:  "POST",
		Pattern: "/job/{id}/action",
		Handler: handler.PostJobActionHandler,
	})

	register(&Route{
		Name:    "GetJobProgress",
		Method:  "Get",
		Pattern: "/job/{id}/progress",
		Handler: handler.GetJobProgressHandler,
	})

	register(&Route{
		Name:    "DeleteJob",
		Method:  "DELETE",
		Pattern: "/job/{id}",
		Handler: handler.DeleteJobHandler,
	})
}

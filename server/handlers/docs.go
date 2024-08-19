package handlers

import (
	"HTTP-download-server/server/models"
	"net/http"

	"github.com/restsend/carrot/apidocs"
)

func (h *Handlers) GetDocs() []apidocs.UriDoc {
	uriDocs := []apidocs.UriDoc{
		{
			Group:   "httpDownloadServer",
			Path:    "/api/task/submit",
			Method:  http.MethodPost,
			Desc:    `提交下载任务`,
			Request: apidocs.GetDocDefine(&DownloadRequest{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "key", Type: apidocs.TYPE_STRING, Desc: "SSE接口需要的参数"},
				},
			},
		},
		{
			Group:   "httpDownloadServer",
			Path:    "/api/task/list",
			Method:  http.MethodPost,
			Desc:    ``,
			Request: apidocs.GetDocDefine(&FilterRequest{}),
			Response: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "totalCount", Type: apidocs.TYPE_STRING, Desc: "总数"},
					{Name: "data", Type: apidocs.TYPE_OBJECT, Desc: "任务列表"},
				},
				Desc: "根据任务类型获取list",
			},
		},
		{
			Group:  "httpDownloadServer",
			Path:   "/api/task/pause",
			Method: http.MethodPost,
			Desc:   `暂停下载`,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ids", Type: apidocs.TYPE_STRING, Desc: "任务id列表"},
				},
			},
			Response: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ok", Type: apidocs.TYPE_STRING},
				},
			},
		},
		{
			Group:  "httpDownloadServer",
			Path:   "/api/task/resume",
			Method: http.MethodPost,
			Desc:   `恢复下载`,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ids", Type: apidocs.TYPE_STRING, Desc: "任务id列表"},
				},
			},
		},
		{
			Group:  "httpDownloadServer",
			Path:   "/api/task/resume",
			Method: http.MethodPost,
			Desc:   `重新下载`,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ids", Type: apidocs.TYPE_STRING, Desc: "任务id列表"},
				},
			},
		},
		{
			Group:  "httpDownloadServer",
			Path:   "/api/task/delete",
			Method: http.MethodPost,
			Desc:   ``,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ids", Type: apidocs.TYPE_STRING, Desc: "任务id列表"},
				},
			},
			Response: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ok", Type: apidocs.TYPE_STRING},
				},
			},
		},
		{
			Group:  "httpDownloadServer",
			Path:   "/api/event/:key",
			Method: http.MethodGet,
			Desc:   ``,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "key", Type: apidocs.TYPE_STRING, Desc: "key, 这个是URL参数"},
				},
			},
			Response: &apidocs.DocField{
				Type: "object",
				Desc: "SSE接口为通用长链接接口，返回数据类型不同",
			},
		},
		{
			Group:  "httpDownloadServer",
			Path:   "/api/settings/:userId",
			Method: http.MethodPost,
			Desc:   `保存设置`,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "userId", Type: apidocs.TYPE_INT, Desc: "userId, 这个是URL参数", Required: true, CanNull: false},
				},
			},
			Response: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ok", Type: apidocs.TYPE_STRING},
				},
			},
		},
		{
			Group:  "httpDownloadServer",
			Path:   "/api/settings/:userId",
			Method: http.MethodGet,
			Desc:   `获取设置`,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "userId", Type: apidocs.TYPE_INT, Desc: "userId, 这个是URL参数", Required: true, CanNull: false},
				},
			},
			Response: apidocs.GetDocDefine(&models.Settings{}),
		},
	}
	return uriDocs
}

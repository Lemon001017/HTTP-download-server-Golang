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
			Group:        "httpDownloadServer",
			Path:         "/api/event/:key",
			Method:       http.MethodGet,
			AuthRequired: true,
			Desc:         ``,
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
			Group:        "httpDownloadServer",
			Path:         "/api/settings/:userId",
			Method:       http.MethodPost,
			AuthRequired: true,
			Desc:         `保存设置`,
			Request: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "userId", Type: apidocs.TYPE_INT, Desc: "userId, 这个是URL参数", Required: true, CanNull: false},
				},
			},
			Response: &apidocs.DocField{
				Fields: []apidocs.DocField{
					{Name: "ok", Type: apidocs.TYPE_INT},
				},
			},
		},
		{
			Group:        "httpDownloadServer",
			Path:         "/api/settings/:userId",
			Method:       http.MethodGet,
			AuthRequired: true,
			Desc:         `获取设置`,
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

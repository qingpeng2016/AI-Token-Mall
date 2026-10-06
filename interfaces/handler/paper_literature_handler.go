package handler

import (
	"net/http"
	"strconv"
	"strings"

	papersvc "github.com/qingpeng2016/ai-token-mall/application/core-service/paper"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/gin-gonic/gin"
)

type PaperLiteratureHandler struct {
	lit *papersvc.LiteratureSearchService
}

func NewPaperLiteratureHandler(lit *papersvc.LiteratureSearchService) *PaperLiteratureHandler {
	return &PaperLiteratureHandler{lit: lit}
}

// Search GET /api/v1/paper/literature/search?query=&sources=arxiv,openalex&limit=20
func (h *PaperLiteratureHandler) Search(c *gin.Context) {
	_, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var sources []string
	if raw := strings.TrimSpace(c.Query("sources")); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			if s := strings.TrimSpace(p); s != "" {
				sources = append(sources, s)
			}
		}
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	data, err := h.lit.Search(c.Request.Context(), query, sources, limit)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

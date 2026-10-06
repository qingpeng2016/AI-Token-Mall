package paper

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	httpentity "github.com/qingpeng2016/ai-token-mall/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-token-mall/domain/http/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

const (
	sourceArxiv           = "arxiv"
	sourceOpenAlex        = "openalex"
	sourceSemanticScholar = "semantic_scholar"
)

type LiteratureSearchService struct {
	arxiv           httprepo.ArxivRepo
	openalex        httprepo.OpenAlexRepo
	semanticScholar httprepo.SemanticScholarRepo
}

func NewLiteratureSearchService(
	arxiv httprepo.ArxivRepo,
	openalex httprepo.OpenAlexRepo,
	semanticScholar httprepo.SemanticScholarRepo,
) *LiteratureSearchService {
	return &LiteratureSearchService{
		arxiv:           arxiv,
		openalex:        openalex,
		semanticScholar: semanticScholar,
	}
}

func (s *LiteratureSearchService) Search(ctx context.Context, query string, sources []string, limit int) (*response.PaperLiteratureSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errorx.ErrParamsError
	}
	if len(sources) == 0 {
		sources = []string{sourceArxiv, sourceOpenAlex, sourceSemanticScholar}
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	out := &response.PaperLiteratureSearchResult{
		Query:   query,
		Sources: make([]response.PaperLiteratureSourceBlock, 0, len(sources)),
	}
	for _, raw := range sources {
		code := normalizeSourceCode(raw)
		if code == "" {
			continue
		}
		block := response.PaperLiteratureSourceBlock{SourceCode: code}
		switch code {
		case sourceArxiv:
			papers, err := s.arxiv.Search(ctx, httpentity.ArxivSearchQuery{Query: query, MaxResults: limit})
			if err != nil {
				block.Error = err.Error()
			} else {
				block.Hits = arxivPapersToHits(papers)
			}
		case sourceOpenAlex:
			works, err := s.openalex.Search(ctx, httpentity.OpenAlexSearchQuery{Query: query, MaxResults: limit})
			if err != nil {
				block.Error = err.Error()
			} else {
				block.Hits = openAlexWorksToHits(works)
			}
		case sourceSemanticScholar:
			papers, err := s.semanticScholar.Search(ctx, httpentity.SemanticScholarSearchQuery{Query: query, MaxResults: limit})
			if err != nil {
				block.Error = err.Error()
			} else {
				block.Hits = semanticScholarPapersToHits(papers)
			}
		default:
			continue
		}
		out.Sources = append(out.Sources, block)
	}
	if len(out.Sources) == 0 {
		return nil, errorx.ErrParamsError
	}
	return out, nil
}

func normalizeSourceCode(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "-", "_")
	if s == "semanticscholar" {
		return sourceSemanticScholar
	}
	return s
}

func arxivPapersToHits(papers []httpentity.ArxivPaper) []response.PaperLiteratureHit {
	hits := make([]response.PaperLiteratureHit, 0, len(papers))
	for _, p := range papers {
		hits = append(hits, response.PaperLiteratureHit{
			SourceCode:    sourceArxiv,
			ExternalKey:   p.ExternalKey,
			Title:         p.Title,
			Authors:       p.Authors,
			Abstract:      p.Abstract,
			PublishedYear: p.PublishedYear,
			URL:           p.URL,
		})
	}
	return hits
}

func openAlexWorksToHits(works []httpentity.OpenAlexWork) []response.PaperLiteratureHit {
	hits := make([]response.PaperLiteratureHit, 0, len(works))
	for _, w := range works {
		hits = append(hits, response.PaperLiteratureHit{
			SourceCode:    sourceOpenAlex,
			ExternalKey:   w.ExternalKey,
			Title:         w.Title,
			Authors:       w.Authors,
			PublishedYear: w.PublishedYear,
			DOI:           w.DOI,
			URL:           w.URL,
		})
	}
	return hits
}

func semanticScholarPapersToHits(papers []httpentity.SemanticScholarPaper) []response.PaperLiteratureHit {
	hits := make([]response.PaperLiteratureHit, 0, len(papers))
	for _, p := range papers {
		hits = append(hits, response.PaperLiteratureHit{
			SourceCode:    sourceSemanticScholar,
			ExternalKey:   p.ExternalKey,
			Title:         p.Title,
			Authors:       p.Authors,
			Abstract:      p.Abstract,
			PublishedYear: p.PublishedYear,
			DOI:           p.DOI,
			URL:           p.URL,
		})
	}
	return hits
}

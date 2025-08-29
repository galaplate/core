package utils

import (
	"strconv"

	"gorm.io/gorm"
)

type PaginatorInfo struct {
	Count        int  `json:"count"`
	CurrentPage  int  `json:"current_page"`
	FirstItem    int  `json:"first_item"`
	HasMorePages bool `json:"has_more_pages"`
	HasPrevPages bool `json:"has_prev_pages"`
	LastItem     int  `json:"last_item"`
	LastPage     int  `json:"last_page"`
	PerPage      int  `json:"per_page"`
	Total        int  `json:"total"`
}

type PaginaterResolver struct {
	PaginatorInfo *PaginatorInfo `json:"paginatorInfo"`
	Data          interface{}    `json:"data"`
	model         interface{}
	request       Request
	stmt          *gorm.DB
}

type Request struct {
	Offset   int
	Page     int
	PageSize int
}

func (s *PaginaterResolver) Stmt(stmt *gorm.DB) *PaginaterResolver {
	s.stmt = stmt
	return s
}

func (s *PaginaterResolver) Model(model interface{}) *PaginaterResolver {
	s.model = model
	return s
}

func (s *PaginaterResolver) Request(p map[string]string) *PaginaterResolver {
	var page = 1
	if p["page"] != "" {
		page, _ = strconv.Atoi(p["page"])
	}

	var pageSize = 10
	if p["page_size"] != "" {
		pageSize, _ = strconv.Atoi(p["page_size"])
	}

	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	s.request = Request{Offset: offset, Page: page, PageSize: pageSize}

	return s
}

func (s *PaginaterResolver) Paginate() (PaginaterResolver, error) {
	var totalCount int64

	s.stmt.Model(s.model).Count(&totalCount)
	limit := s.request.PageSize
	page := s.request.Page
	offset := s.request.Offset
	lastPage := int((totalCount + int64(limit) - 1) / int64(limit))
	firstPage := page == 1

	result := s.stmt.Scopes(s.Paging()).Find(s.model)
	if result.RowsAffected == 0 {
		return PaginaterResolver{
			PaginatorInfo: &PaginatorInfo{
				Count:        int(result.RowsAffected),
				CurrentPage:  page,
				FirstItem:    offset + 1,
				HasMorePages: page < lastPage,
				HasPrevPages: firstPage,
				LastItem:     offset + int(result.RowsAffected),
				LastPage:     lastPage,
				PerPage:      limit,
				Total:        int(totalCount),
			},
			Data: []interface{}{},
		}, nil
	}

	paginatorInfo := &PaginatorInfo{
		Count:        int(result.RowsAffected),
		CurrentPage:  page,
		FirstItem:    offset + 1,
		HasMorePages: page < lastPage,
		HasPrevPages: !firstPage,
		LastItem:     offset + int(result.RowsAffected),
		LastPage:     lastPage,
		PerPage:      limit,
		Total:        int(totalCount),
	}

	s.PaginatorInfo = paginatorInfo
	s.Data = s.model

	return *s, nil
}

func (s *PaginaterResolver) Paging() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(s.request.Offset).Limit(s.request.PageSize)
	}
}

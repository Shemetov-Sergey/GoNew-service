package services

import (
	"github.com/Shemetov-Sergey/GoNew-service/pkg/cache"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/pb/comment"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/pb/gonews"

	"context"
	"net/http"
	"time"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/storage"
)

type Server struct {
	H storage.PostsInterface
	P *cache.PaginationCache
	C comment.CommentServiceClient
}

func (s *Server) Posts(ctx context.Context, req *gonews.PostsRequest) (*gonews.PostsResponse, error) {
	news, err := s.H.Posts(int(req.NewsCountGet))

	if err != nil {
		return &gonews.PostsResponse{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}, nil
	}

	return &gonews.PostsResponse{
		Status: http.StatusOK,
		Posts:  news,
	}, nil
}

func (s *Server) NewsFullDetailed(ctx context.Context, req *gonews.OneNewsRequest) (*gonews.DetailedNewsResponse, error) {
	news, err := s.H.OneNews(int64(req.NewsId))

	if err != nil {
		return &gonews.DetailedNewsResponse{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}, nil
	}

	r := comment.CommentsByNewsRequest{NewsId: uint64(req.NewsId)}
	comments, err := s.C.CommentsByNews(ctx, &r)

	return &gonews.DetailedNewsResponse{
		Status:   http.StatusOK,
		Post:     news,
		Comments: comments.Comments,
	}, nil
}

func (s *Server) NewsShortDetailed(ctx context.Context, req *gonews.OneNewsRequest) (*gonews.OnePostResponse, error) {
	news, err := s.H.OneNews(req.NewsId)

	if len(news.Content) > 200 {
		shortContent := news.Content[0:200]
		news.Content = shortContent
	}

	if err != nil {
		return &gonews.OnePostResponse{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}, nil
	}

	return &gonews.OnePostResponse{
		Status: http.StatusOK,
		Posts:  news,
	}, nil
}

func (s *Server) FilterNews(ctx context.Context, req *gonews.FilterNewsRequest) (*gonews.ListPostsResponse, error) {
	paginationObject, ok := s.P.Sessions[req.UserId]

	if !ok {
		posts, err := s.H.FilterNews(req.FilterValue)
		if err != nil {
			return &gonews.ListPostsResponse{
				Status: http.StatusNotFound,
				Error:  err.Error(),
			}, nil
		}
		start := time.Now()
		session := cache.NewSession(start, 5, int64(len(posts)))
		s.P.AddSession(session, posts, req.UserId)
		paginationObject = s.P.Sessions[req.UserId]
	}

	posts := paginationObject.Values

	var postsToShow []*gonews.Post
	var paginationInfo *gonews.Pagination
	var pageSize int32
	var page int32

	if req.PageSize == 0 {
		pageSize = 1
	} else {
		pageSize = req.PageSize
	}

	if req.Page > 0 {
		page = req.Page
	} else {
		page = 1
	}

	if int64(pageSize*page) > int64(len(posts)) {
		pages := 1
		paginationInfo = cache.NewsPaginationInfo(int32(pages), 1, int32(len(posts)))
		postsToShow = posts[0 : int64(len(posts))-1]
	} else {
		basePages := int32(len(posts)) / pageSize
		lastPage := 0
		if int32(len(posts))%pageSize > 0 {
			lastPage++
		}
		pages := basePages + int32(lastPage)
		currentOffset := (page - 1) * pageSize
		paginationInfo = cache.NewsPaginationInfo(pages, page, pageSize)
		postsToShow = posts[currentOffset : currentOffset+pageSize]
	}

	return &gonews.ListPostsResponse{
		Status:         http.StatusOK,
		PaginationInfo: paginationInfo,
		Posts:          postsToShow,
	}, nil
}

func (s *Server) ListNews(ctx context.Context, req *gonews.ListPostsRequest) (*gonews.ListPostsResponse, error) {
	paginationObject, ok := s.P.Sessions[req.UserId]

	if !ok {
		posts, err := s.H.Posts(int(req.NewsCountGet))
		if err != nil {
			return &gonews.ListPostsResponse{
				Status: http.StatusNotFound,
				Error:  err.Error(),
			}, nil
		}
		start := time.Now()
		session := cache.NewSession(start, 5, req.NewsCountGet)
		s.P.AddSession(session, posts, req.UserId)
		paginationObject = s.P.Sessions[req.UserId]
	}

	if req.NewsCountGet != paginationObject.Session.NewsCount {
		posts, err := s.H.Posts(int(req.NewsCountGet))
		if err != nil {
			return &gonews.ListPostsResponse{
				Status: http.StatusNotFound,
				Error:  err.Error(),
			}, nil
		}
		start := time.Now()
		session := cache.NewSession(start, 5, req.NewsCountGet)
		s.P.AddSession(session, posts, req.UserId)
		paginationObject = s.P.Sessions[req.UserId]
	}

	posts := paginationObject.Values
	var postsToShow []*gonews.Post
	var paginationInfo *gonews.Pagination

	if int64(req.PageSize*req.Page) > req.NewsCountGet {
		pages := 1
		paginationInfo = cache.NewsPaginationInfo(int32(pages), 1, int32(req.NewsCountGet))
		postsToShow = posts[0 : req.NewsCountGet-1]
	} else {
		basePages := int32(len(posts)) / req.PageSize
		lastPage := 0
		if int32(len(posts))%req.PageSize > 0 {
			lastPage++
		}
		pages := basePages + int32(lastPage)

		currentOffset := (req.Page - 1) * req.PageSize
		paginationInfo = cache.NewsPaginationInfo(pages, req.Page, req.PageSize)
		postsToShow = posts[currentOffset : currentOffset+req.PageSize]
	}

	return &gonews.ListPostsResponse{
		Status:         http.StatusOK,
		PaginationInfo: paginationInfo,
		Posts:          postsToShow,
	}, nil
}

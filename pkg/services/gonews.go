package services

import (
	"GoNew-service/pkg/pb"
	"GoNew-service/pkg/storage"
	"context"
	"net/http"
)

type Server struct {
	H storage.PostsInterface
}

func (s *Server) Posts(ctx context.Context, req *pb.PostsRequest) (*pb.PostsResponse, error) {
	news, err := s.H.Posts(int(req.NewsCountGet))

	if err != nil {
		return &pb.PostsResponse{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}, nil
	}

	return &pb.PostsResponse{
		Status: http.StatusOK,
		Posts:  news,
	}, nil
}

func (s *Server) NewsFullDetailed(ctx context.Context, req *pb.OneNewsRequest) (*pb.OnePostResponse, error) {
	news, err := s.H.OneNews(req.NewsId)

	if err != nil {
		return &pb.OnePostResponse{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}, nil
	}

	return &pb.OnePostResponse{
		Status: http.StatusOK,
		Posts:  news,
	}, nil
}

func (s *Server) FilterNews(ctx context.Context, req *pb.FilterNewsRequest) (*pb.PostsResponse, error) {
	news, err := s.H.FilterNews(req.FilterValue)

	if err != nil {
		return &pb.PostsResponse{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}, nil
	}

	return &pb.PostsResponse{
		Status: http.StatusOK,
		Posts:  news,
	}, nil
}

func (s *Server) NewsShortDetailed(ctx context.Context, req *pb.OneNewsRequest) (*pb.OnePostResponse, error) {
	news, err := s.H.OneNews(req.NewsId)

	if len(news.Content) > 200 {
		shortContent := news.Content[0:200]
		news.Content = shortContent
	}

	if err != nil {
		return &pb.OnePostResponse{
			Status: http.StatusNotFound,
			Error:  err.Error(),
		}, nil
	}

	return &pb.OnePostResponse{
		Status: http.StatusOK,
		Posts:  news,
	}, nil
}

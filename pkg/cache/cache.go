package cache

import (
	"time"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/pb/gonews"
)

type PaginationSession struct {
	Start           time.Time
	DurationMinutes uint8
	NewsCount       int64
}

func NewSession(start time.Time, duration uint8, newCount int64) *PaginationSession {
	return &PaginationSession{
		Start:           start,
		DurationMinutes: duration,
		NewsCount:       newCount,
	}
}

func NewsPaginationInfo(pages, currentPage, PostsOnPage int32) *gonews.Pagination {
	return &gonews.Pagination{
		Pages:       pages,
		CurrentPage: currentPage,
		PostsOnPage: PostsOnPage,
	}
}

type NewsPagination struct {
	Session *PaginationSession
	Values  []*gonews.Post
}

func NewPagination(session *PaginationSession, values []*gonews.Post) *NewsPagination {
	return &NewsPagination{
		Session: session,
		Values:  values,
	}
}

type PaginationCache struct {
	Sessions      map[int64]*NewsPagination
	checkInterval time.Duration
}

func NewPaginationCache() *PaginationCache {
	return &PaginationCache{
		Sessions:      make(map[int64]*NewsPagination),
		checkInterval: 5 * time.Minute,
	}
}

func (pc *PaginationCache) AddSession(session *PaginationSession, posts []*gonews.Post, userId int64) {
	p := NewPagination(session, posts)
	pc.Sessions[userId] = p
}

func (pc *PaginationCache) checkSession() {
	go func() {
		ticker := time.NewTicker(pc.checkInterval)
		for {
			select {
			case <-ticker.C:
				for newsId, s := range pc.Sessions {
					timeFromStart := int64(s.Session.Start.Nanosecond()) + (time.Minute * time.Duration(s.Session.DurationMinutes)).Nanoseconds()
					now := int64(time.Now().Nanosecond())
					if now-timeFromStart < 0 {
						delete(pc.Sessions, newsId)
					}
				}
			}
		}
	}()
}

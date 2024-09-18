package rss

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/storage"
)

func readFromPostsChan(postsChan chan storage.Post) {
	for item := range postsChan {
		log.Println(item)
	}
}

func readFromErrorChan(postsChan chan error) {
	for item := range postsChan {
		log.Println(item)
	}
}

func TestNewPostsService(t *testing.T) {
	type args struct {
		source        string
		postChan      chan storage.Post
		errChan       chan error
		requestPeriod int
	}

	postsChan := make(chan storage.Post)
	errChan := make(chan error)

	tests := []struct {
		name string
		args args
		want *PostsServiceInstance
	}{
		{
			name: "NewPostsService test",
			args: args{
				source:        "https://habr.com/ru/rss/hub/go/all/?fl=ru?limit=10",
				postChan:      postsChan,
				errChan:       errChan,
				requestPeriod: 1,
			},
			want: &PostsServiceInstance{
				sourceXML:       "https://habr.com/ru/rss/hub/go/all/?fl=ru?limit=10",
				LastPubDateUnix: 0,
				PostsChan:       postsChan,
				ErrorChan:       errChan,
				RequestPeriod:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPostsService(tt.args.source, tt.args.postChan, tt.args.errChan, tt.args.requestPeriod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPostsService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSourceRss(t *testing.T) {
	type args struct {
		rssConfigFile string
		postsChan     chan storage.Post
		errChan       chan error
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "NewSourceRss test",
			args: args{
				rssConfigFile: "../../cmd/config.json",
				postsChan:     make(chan storage.Post),
				errChan:       make(chan error),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSourceRss(tt.args.rssConfigFile, tt.args.postsChan, tt.args.errChan)
			if err != tt.wantErr {
				t.Errorf("NewSourceRss() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPostsServiceInstance_AddInfoFromSource(t *testing.T) {
	type fields struct {
		sourceXML       string
		LastPubDateUnix int64
		PostsChan       chan storage.Post
		ErrorChan       chan error
		RequestPeriod   int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "AddInfoFromSource test",
			fields: fields{
				sourceXML:       "https://habr.com/ru/rss/hub/go/all/?fl=ru?limit=10",
				LastPubDateUnix: 1671032459,
				PostsChan:       make(chan storage.Post),
				ErrorChan:       make(chan error),
				RequestPeriod:   1,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PostsServiceInstance{
				sourceXML:       tt.fields.sourceXML,
				LastPubDateUnix: tt.fields.LastPubDateUnix,
				PostsChan:       tt.fields.PostsChan,
				ErrorChan:       tt.fields.ErrorChan,
				RequestPeriod:   tt.fields.RequestPeriod,
			}
			// Читаем из каналов, иначе зависает
			go readFromPostsChan(p.PostsChan)
			go readFromErrorChan(p.ErrorChan)
			if err := p.AddInfoFromSource(); err != tt.wantErr {
				t.Errorf("AddInfoFromSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostsServiceInstance_checkDateActuality(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	type fields struct {
		sourceXML       string
		LastPubDateUnix int64
		PostsChan       chan storage.Post
		ErrorChan       chan error
		RequestPeriod   int
	}
	type args struct {
		pubDate time.Time
	}

	// Первый тест для проходит, второй для проверки случая с ошибкой
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "checkDateActuality test",
			fields: fields{
				sourceXML:       "https://habr.com/ru/rss/hub/go/all/?fl=ru?limit=10",
				LastPubDateUnix: 1671032459,
				PostsChan:       make(chan storage.Post),
				ErrorChan:       make(chan error),
				RequestPeriod:   1,
			},
			args: args{
				pubDate: time.Date(2022, 11, 30, 15, 0, 0, 0, loc),
			},
			want: false,
		},
		{
			name: "checkDateActuality test",
			fields: fields{
				sourceXML:       "https://habr.com/ru/rss/hub/go/all/?fl=ru?limit=10",
				LastPubDateUnix: 1671032459,
				PostsChan:       make(chan storage.Post),
				ErrorChan:       make(chan error),
				RequestPeriod:   1,
			},
			args: args{
				pubDate: time.Date(2022, 11, 30, 15, 0, 0, 0, loc),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PostsServiceInstance{
				sourceXML:       tt.fields.sourceXML,
				LastPubDateUnix: tt.fields.LastPubDateUnix,
				PostsChan:       tt.fields.PostsChan,
				ErrorChan:       tt.fields.ErrorChan,
				RequestPeriod:   tt.fields.RequestPeriod,
			}
			if got := p.checkDateActuality(tt.args.pubDate); got != tt.want {
				t.Errorf("checkDateActuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

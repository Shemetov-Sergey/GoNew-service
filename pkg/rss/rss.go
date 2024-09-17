package rss

import (
	"GoNew-service/pkg/storage"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	strip "github.com/grokify/html-strip-tags-go"
)

type SourceRss struct {
	SourceRssList     []string `json:"rss"`
	RequestedPeriod   int      `json:"request_period"`
	Channel           Channel  `xml:"channel"`
	PostsChan         chan storage.Post
	ErrorChan         chan error
	LastPubTimeFromDB map[string]int64
}

type Channel struct {
	XMLName     xml.Name          `xml:"channel"`
	Title       string            `xml:"title"`
	Link        string            `xml:"link"`
	Description string            `xml:"description"`
	Items       []storage.RawPost `xml:"item"`
}

// NewSourceRss создает новый экземпляр структуры SourceRss. Метод получает данные об имени
// файла конфигурации и данные о каналах, которые затем передает инстансам PostsServiceInstance.
func NewSourceRss(rssConfigFile string, postsChan chan storage.Post, errChan chan error) (SourceRss, error) {
	log.Println("Start NewSourceRss")
	file, err := os.Open(rssConfigFile)
	var SourceRssInfo SourceRss
	SourceRssInfo.PostsChan = postsChan
	SourceRssInfo.ErrorChan = errChan

	if err != nil {
		errChan <- err
		return SourceRss{}, err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			errChan <- err
		}
	}(file)

	configInfoBytes, err := io.ReadAll(file)

	if err != nil {
		errChan <- err
		return SourceRss{}, err
	}

	err = json.Unmarshal(configInfoBytes, &SourceRssInfo)

	if err != nil {
		errChan <- err
		return SourceRss{}, err
	}

	return SourceRssInfo, nil
}

// RunGetSourcesInfo создает столько инстансов PostsServiceInstance, сколько есть источников RSS.
// Затем каждый PostsServiceInstance запускает функцию для получения данных из ленты RSS.
func (s *SourceRss) RunGetSourcesInfo() {
	log.Println("Start RunGetSourcesInfo")
	for _, source := range s.SourceRssList {
		postsService := NewPostsService(source, s.PostsChan, s.ErrorChan, s.RequestedPeriod)
		if postsService.LastPubDateUnix == 0 && s.LastPubTimeFromDB[source] != 0 {
			postsService.LastPubDateUnix = s.LastPubTimeFromDB[source]
		}
		postsService.RunRss()
	}
}

type PostsServiceInstance struct {
	sourceXML       string
	LastPubDateUnix int64
	PostsChan       chan storage.Post
	ErrorChan       chan error
	RequestPeriod   int
}

// NewPostsService создает новый экземпляр PostsServiceInstance
func NewPostsService(source string, postChan chan storage.Post, errChan chan error, requestPeriod int) *PostsServiceInstance {
	log.Println("Start NewPostsService")
	return &PostsServiceInstance{
		sourceXML:     source,
		PostsChan:     postChan,
		ErrorChan:     errChan,
		RequestPeriod: requestPeriod,
	}
}

// AddInfoFromSource получает данные из источника, который добавлен в поле sourceXML инстанста PostsServiceInstance.
// Данные записываются, только если они удовлетворяют условию, что они созданы позже, чем существует последняя
// запись для этой ленты RSS. Это позволяет игнорировать старые записи и возвращать в ленту только новые актуальные
// записи.
func (p *PostsServiceInstance) AddInfoFromSource() error {
	log.Println("Start AddInfoFromSource")
	resp, err := http.Get(p.sourceXML)

	var lastDateUnixMax int64 = 0

	if err != nil {
		p.ErrorChan <- err
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			p.ErrorChan <- err
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.ErrorChan <- err
		return err
	}
	rss := SourceRss{}
	xml.Unmarshal(body, &rss)
	for _, item := range rss.Channel.Items {
		// Для разных форматов времени парсинг будет идти по разному
		pd, err := time.Parse(time.RFC1123Z, item.PubTime)

		if err != nil {
			p.ErrorChan <- err
			pd, err = time.Parse(time.RFC1123, item.PubTime)
		}

		// Проверяем актуальность новости. Чтобы не забивать бд дублями при перезапусках
		if ok := p.checkDateActuality(pd); ok {
			post := storage.Post{
				Title:         item.Title,
				Content:       strip.StripTags(item.Content),
				PubTime:       pd.Unix(),
				Link:          item.Link,
				SourceXmlLink: p.sourceXML,
			}
			if lastDateUnixMax < post.PubTime {
				lastDateUnixMax = post.PubTime
			}
			p.PostsChan <- post
		}
	}

	if lastDateUnixMax > p.LastPubDateUnix {
		p.LastPubDateUnix = lastDateUnixMax
	}

	return nil
}

// checkDateActuality сравнивает дату публикации полученную из rss с последней датой публикации
func (p *PostsServiceInstance) checkDateActuality(pubDate time.Time) bool {
	log.Println("Start checkDateActuality")
	if pubDate.Unix() > p.LastPubDateUnix {
		return true
	}
	return false
}

// RunRss позволяет PostsServiceInstance начать обход ленты RSS согласно периоду, указанному в конфигурации приложения
func (p *PostsServiceInstance) RunRss() {
	log.Println("Start RunRss")
	go func() {
		ticker := time.NewTicker(time.Duration(p.RequestPeriod) * time.Minute)
		for {
			select {
			case <-ticker.C:
				err := p.AddInfoFromSource()
				if err != nil {
					p.ErrorChan <- err
				}
			}
		}
	}()
}

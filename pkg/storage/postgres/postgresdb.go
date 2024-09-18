package postgres

import (
	"context"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/pb/gonews"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/storage"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Storage struct {
	Db        *pgxpool.Pool
	postsChan chan storage.Post
	errorChan chan error
}

// New Создает новый экземпляр структуры Storage
func New(ctx context.Context, constr string, posts chan storage.Post, errChan chan error) (*Storage, error) {
	db, err := pgxpool.Connect(ctx, constr)
	if err != nil {
		return nil, err
	}
	s := Storage{
		Db:        db,
		postsChan: posts,
		errorChan: errChan,
	}
	return &s, nil
}

// Posts возвращает указанное в переменной countPosts количество записей из базы данных
func (pg *Storage) Posts(countPosts int) ([]*gonews.Post, error) {

	query := `SELECT id, title, content, pub_time, link, source_link FROM posts ORDER BY pub_time DESC LIMIT $1;`
	rows, rowsErr := pg.Db.Query(context.Background(), query, countPosts)

	if rowsErr != nil {
		pg.errorChan <- rowsErr
		return nil, rowsErr
	}

	var posts []*gonews.Post

	for rows.Next() {
		var post gonews.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.PubTime,
			&post.Link,
			&post.SourceXmlLink,
		)

		if err != nil {
			pg.errorChan <- err
			return nil, err
		}

		posts = append(posts, &post)
	}

	return posts, rows.Err()
}

// OneNews возвращает одну новость по newsId
func (pg *Storage) OneNews(newsId int64) (*gonews.Post, error) {
	query := `SELECT id, title, content, pub_time, link, source_link FROM posts WHERE id=$1;`
	row := pg.Db.QueryRow(context.Background(), query, newsId)

	var post gonews.Post

	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.PubTime,
		&post.Link,
		&post.SourceXmlLink,
	)

	if err != nil {
		pg.errorChan <- err
		return nil, err
	}

	return &post, nil
}

func (pg *Storage) FilterNews(filterTitle string) ([]*gonews.Post, error) {
	queryParameter := "%" + filterTitle + "%"
	query := `SELECT id, title, content, pub_time, link, source_link FROM posts WHERE title ILIKE $1;`
	rows, rowsErr := pg.Db.Query(context.Background(), query, queryParameter)

	if rowsErr != nil {
		pg.errorChan <- rowsErr
		return nil, rowsErr
	}

	var posts []*gonews.Post

	for rows.Next() {
		var post gonews.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.PubTime,
			&post.Link,
			&post.SourceXmlLink,
		)

		if err != nil {
			pg.errorChan <- err
			return nil, err
		}

		posts = append(posts, &post)
	}

	return posts, rows.Err()
}

// AddPost создает запись в базе данных о storage.Post
func (pg *Storage) AddPost(post storage.Post) error {

	query := `INSERT INTO posts(title, content, pub_time, link, source_link) VALUES($1, $2, $3, $4, $5);`
	_, err := pg.Db.Exec(context.Background(), query, post.Title, post.Content, post.PubTime, post.Link, post.SourceXmlLink)

	return err
}

// RunInsertPosts запускает горутины, которые записывают в базу данных записи storage.Post,
// когда они поступают в  канал postsChan
func (pg *Storage) RunInsertPosts() {
	go func() {
		for {
			select {
			case post := <-pg.postsChan:
				err := pg.AddPost(post)
				if err != nil {
					pg.errorChan <- err
					return
				}
			}
		}
	}()
}

// GetLastPubDateForSources нужен для получения данных о последней записи из базы данных для
// соответствующего источника RSS. Данный метод нужен при повторных запусках приложения,
// чтобы не получать дублирующиеся записи в базе данных при повторном обходе ленты RSS
func (pg *Storage) GetLastPubDateForSources(sourceSlice []string) (map[string]int64, error) {
	query := `SELECT pub_time FROM posts WHERE source_link = $1 ORDER BY pub_time DESC LIMIT 1;`

	lastPubTimeMap := make(map[string]int64)

	for _, source := range sourceSlice {
		var lastPubTime int64
		rows, err := pg.Db.Query(context.Background(), query, source)

		if err != nil {
			pg.errorChan <- err
			return nil, err
		}

		for rows.Next() {
			err = rows.Scan(&lastPubTime)
			if err != nil {
				pg.errorChan <- err
				return nil, err
			}
		}

		lastPubTimeMap[source] = lastPubTime
	}

	return lastPubTimeMap, nil
}

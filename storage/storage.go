package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/luqus/templater/types"
)

type Storage interface {
	CreateArticle(article *types.Article) error
	GetAllArticles() ([]*types.Article, error)
	GetArticle(articleID string) (*types.Article, error)
	UpdateArticle(id string, article *types.Article) error
	DeleteArticle(articleID string) error
}

type SqliteStorage struct {
	db *sql.DB
}

func NewSqliteStorage() (*SqliteStorage, error) {
	var err error
	db, err := sql.Open("sqlite3", "./data.sqlite")
	if err != nil {
		return nil, err
	}

	sqlStmt := `create table if not exists articles (id integer not null primary key autoincrement, title text, content text)`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &SqliteStorage{db: db}, nil
}

func (s *SqliteStorage) CreateArticle(article *types.Article) error {
	query, err := s.db.Prepare("insert into articles(title, content) values (?,?)")
	if err != nil {
		return err
	}
	defer query.Close()

	_, err = query.Exec(article.Title, article.Content)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqliteStorage) GetAllArticles() ([]*types.Article, error) {
	query, err := s.db.Prepare("select id, title, content from articles")
	if err != nil {
		return nil, err
	}

	defer query.Close()

	result, err := query.Query()

	if err != nil {
		return nil, err
	}

	articles := make([]*types.Article, 0)

	for result.Next() {
		data := new(types.Article)
		err := result.Scan(
			&data.ID,
			&data.Title,
			&data.Content,
		)

		if err != nil {
			return nil, err
		}

		articles = append(articles, data)
	}

	return articles, nil
}

func (s *SqliteStorage) GetArticle(articleID string) (*types.Article, error) {
	query, err := s.db.Prepare("select id, title, content from articles where id = ?")
	if err != nil {
		return nil, err
	}

	defer query.Close()

	result := query.QueryRow(articleID)
	data := new(types.Article)
	err = result.Scan(&data.ID, &data.Title, &data.Content)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *SqliteStorage) UpdateArticle(id string, article *types.Article) error {
	query, err := s.db.Prepare("update articles set (title, content) = (?, ?) where id = ?")
	if err != nil {
		return err
	}

	defer query.Close()

	_, err = query.Exec(article.Title, article.Content, id)

	if err != nil {
		return err
	}

	return nil
}

func (s *SqliteStorage) DeleteArticle(articleID string) error {
	query, err := s.db.Prepare("delete from articles where id = ?")
	if err != nil {
		return err
	}

	defer query.Close()

	_, err = query.Exec(articleID)
	if err != nil {
		return err
	}

	return err
}

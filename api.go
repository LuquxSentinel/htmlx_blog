package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luqus/templater/storage"
	"github.com/luqus/templater/types"
)

type APIServer struct {
	listenAddress string
	router        chi.Router
	storage       storage.Storage
}

func NewAPIServer(listenAddress string, storage storage.Storage) *APIServer {
	return &APIServer{
		listenAddress: listenAddress,
		router:        chi.NewRouter(),
		storage:       storage,
	}
}

func (s *APIServer) Run() error {

	s.router.Use(middleware.Recoverer)

	s.router.Use(ChangeMethod)
	s.router.Get("/", s.getAllArticles)
	s.router.Route("/articles", func(r chi.Router) {
		r.Get("/", s.NewArticle)
		r.Post("/", s.CreateArticle)
		r.Route("/{articleID}", func(r chi.Router) {
			r.Use(s.ArticleCtx)
			r.Get("/", s.GetArticle)
			r.Put("/", s.UpdateArticle)
			r.Delete("/", s.DeleteArticle)
			r.Get("/edit", s.EditArticle)
		})
	})

	return http.ListenAndServe(s.listenAddress, s.router)
}

func (s *APIServer) getAllArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := s.storage.GetAllArticles()
	catch(err)

	t, _ := template.ParseFiles("templates/index.html")
	err = t.Execute(w, articles)
	catch(err)
}

func (s *APIServer) NewArticle(w http.ResponseWriter, r *http.Request) {
	// TODO: Render template
}

func (s *APIServer) CreateArticle(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")
	article := &types.Article{
		Title:   title,
		Content: template.HTML(content),
	}

	err := s.storage.CreateArticle(article)
	catch(err)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *APIServer) GetArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*types.Article)
	log.Println(article)

	// TODO: Render template
}

func (s *APIServer) EditArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*types.Article)
	log.Println(article)

	// TODO: Render template
}

func (s *APIServer) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*types.Article)

	title := r.FormValue("title")
	content := r.FormValue("content")
	newArticle := &types.Article{
		Title:   title,
		Content: template.HTML(content),
	}

	err := s.storage.UpdateArticle(strconv.Itoa(article.ID), newArticle)
	catch(err)

	http.Redirect(w, r, fmt.Sprintf("/articles/%d", article.ID), http.StatusFound)
}

func (s *APIServer) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*types.Article)
	err := s.storage.DeleteArticle(strconv.Itoa(article.ID))
	catch(err)

	http.Redirect(w, r, "/", http.StatusFound)
}

func ChangeMethod(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			switch method := r.PostFormValue("_method"); method {
			case http.MethodPut:
				fallthrough

			case http.MethodPatch:
				fallthrough

			case http.MethodDelete:
				r.Method = method

			default:
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) ArticleCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := chi.URLParam(r, "articleID")
		article, err := s.storage.GetArticle(articleID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(404), 404)
			return
		}

		ctx := context.WithValue(r.Context(), "article", article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func catch(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

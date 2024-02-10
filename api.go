package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	s.router.Post("/upload", s.UploadHandler)
	s.router.Get("/images/*", s.ServeImages)

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

	log.Println()

	t, err := template.ParseFiles("templates/base.html", "templates/index.html")
	catch(err)
	err = t.Execute(w, articles)
	catch(err)
}

func (s *APIServer) NewArticle(w http.ResponseWriter, r *http.Request) {
	// TODO: Render template
	t, _ := template.ParseFiles("templates/base.html", "templates/new.html")
	err := t.Execute(w, nil)
	catch(err)
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

	// TODO: Render template
	t, _ := template.ParseFiles("templates/base.html", "templates/article.html")
	err := t.Execute(w, article)
	catch(err)
}

func (s *APIServer) EditArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*types.Article)

	// TODO: Render template
	t, _ := template.ParseFiles("templates/base.html", "templates/edit.html")
	err := t.Execute(w, article)
	catch(err)

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

func (s *APIServer) UploadHandler(w http.ResponseWriter, r *http.Request) {
	const MAX_UPLOAD_SIZE = 10 << 20 // Set the max upload size to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	err := r.ParseMultipartForm(MAX_UPLOAD_SIZE)
	if err != nil {
		http.Error(w, "The uploaded file is too big. Please choose a file that's less than 10 MB", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	// Create the uploads folder if it doesn't already exist
	err = os.MkdirAll("./images", os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("/images/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
	dst, err := os.Create("." + filename)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer dst.Close()

	// Copy the uploaded file to specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(filename)
	response, _ := json.Marshal(map[string]string{"location": filename})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func (s *APIServer) ServeImages(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	fs := http.StripPrefix("/images/", http.FileServer(http.Dir("./images")))
	fs.ServeHTTP(w, r)
}

func catch(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

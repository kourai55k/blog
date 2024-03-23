package main

import (
	. "blog/data"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func main() {
	// Подключаемся к базе данных
	db, err := ConnectToDB()
	if err != nil {
		log.Fatal(err)
	}

	// Создаём таблицы в базе данных
	err = CreateTables(db)
	if err != nil {
		log.Fatal(err)
	}

	// Создаём роутер mux
	mux := http.NewServeMux()

	// Создаём обработчик статических файлов CSS
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Обрабатываем главную страницу
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Парсим шаблон из файла
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Fatal(err)
		}
		// Получаем данные из БД
		data, err := GetPosts(db)
		if err != nil {
			log.Fatal(err)
		}
		// Запускаем шаблон, передаём в него данные
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Обрабатываем GET /register -- страница для регистрации
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		// Парсим шаблон из файла
		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			log.Fatal(err)
		}

		// Запускаем шаблон, передаём в него данные
		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Обрабатываем POST /register -- страница для регистрации
	mux.HandleFunc("POST /register", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		login := r.FormValue("login")
		password := r.FormValue("password")

		// Хеширование пароля
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}
		strHashedPass := string(hashedPassword)

		// Добавление нового пользователя в базу данных
		user := User{Name: name, Login: login, HashedPass: strHashedPass}
		err = CreateUser(db, user)
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	})

	// Обрабатываем GET /login -- страница для логина
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		// Парсим шаблон из файла
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			log.Fatal(err)
		}

		// Запускаем шаблон, передаём в него данные
		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Обрабатываем POST /login -- страница для логина
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		// Получаем данные из формы
		login := r.FormValue("login")
		pass := r.FormValue("password")

		//Берём пароль пользователя из базы данных для аутентификации
		user, err := GetUserByLogin(db, login)
		if err != nil {
			log.Fatal(err)
		}
		hashedPass := []byte(user.HashedPass)
		// Сравниваем пароль из базы данных с введённым
		err = bcrypt.CompareHashAndPassword(hashedPass, []byte(pass))
		if err != nil {
			fmt.Println("Password does not match!")
		} else {
			fmt.Println("Password matches!")
		}
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	})

	// Обрабатываем маршрут GET /posts -- Все посты
	mux.HandleFunc("GET /posts", func(w http.ResponseWriter, r *http.Request) {
		// Парсим шаблон из файла
		tmpl, err := template.ParseFiles("templates/posts.html")
		if err != nil {
			log.Fatal(err)
		}
		// Получаем данные из БД
		data, err := GetPosts(db)
		if err != nil {
			log.Fatal(err)
		}
		// Запускаем шаблон, передаём в него данные
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Обрабатываем маршрут GET /posts/{id} -- Один пост
	mux.HandleFunc("GET /posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Получаем id поста из запроса
		strId := r.PathValue("id")
		id, err := strconv.Atoi(strId)
		if err != nil {
			log.Fatal(err)
		}

		// Получаем данные из БД
		data, err := GetPost(db, id)
		if err != nil {
			log.Fatal(err)
		}

		// Парсим шаблон из файла
		tmpl, err := template.ParseFiles("templates/post.html")
		if err != nil {
			log.Fatal(err)
		}

		// Запускаем шаблон, передаём в него данные
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Обрабатываем маршрут GET /posts/new -- Новый пост
	mux.HandleFunc("GET /posts/new", func(w http.ResponseWriter, r *http.Request) {
		// Парсим шаблон из файла
		tmpl, err := template.ParseFiles("templates/createPost.html")
		if err != nil {
			log.Fatal(err)
		}

		// Запускаем шаблон, передаём в него данные
		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Обрабатываем маршрут POST /posts/new -- Новый пост
	mux.HandleFunc("POST /posts/new", func(w http.ResponseWriter, r *http.Request) {
		// Получаем данные из формы
		title := r.FormValue("title")
		body := r.FormValue("body")

		// Создаём объект структуры
		post := Post{Title: title, Body: body}
		// Передаём объект в функцию создания новой записи в БД
		err := CreatePost(db, post)
		if err != nil {
			log.Fatal(err)
		}

		// Перенаправляем на страницу со всеми постами
		http.Redirect(w, r, "/posts", http.StatusPermanentRedirect)
	})

	// Обрабатываем маршрут /posts/delete/{id} -- Удаление поста
	mux.HandleFunc("/posts/delete/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Получаем id поста из запроса
		strId := r.PathValue("id")
		id, err := strconv.Atoi(strId)
		if err != nil {
			log.Fatal(err)
		}
		err = DeletePost(db, id)
		if err != nil {
			log.Fatal(err)
		}

		http.Redirect(w, r, "/posts", http.StatusPermanentRedirect)
	})

	// Запускаем сервер
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

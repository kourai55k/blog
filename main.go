package main

import (
	. "blog/data"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"math/rand"
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

	// Сессии
	// Создаем буфер для хранения случайных байтов
	key := make([]byte, 32)

	// Читаем случайные байты в буфер
	_, err = rand.Read(key)
	if err != nil {
		panic(err)
	}

	// Преобразуем случайные байты в строку в шестнадцатеричном формате
	secretKey := hex.EncodeToString(key)

	// Инициализация хранилища сессий с секретным ключом
	store := sessions.NewCookieStore([]byte(secretKey))

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

		// Получение или создание сессии
		session, _ := store.Get(r, "user-session")

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
			// Аутентификация успешна, установка флага аутентификации в сессии
			session.Values["authenticated"] = true
			// Сохранение ID пользователя в сессии
			session.Values["userId"] = user.Id
			// Сохранение имени пользователя в сессии
			session.Values["author"] = user.Name

			err = session.Save(r, w)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Session saved")
		}
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	})

	// Обрабатываем GET /logout -- страница для выхода из аккаунта
	mux.HandleFunc("GET /logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "user-session")
		session.Values["authenticated"] = false // Установка флага аутентификации в false
		session.Values["userId"] = nil          // Очистка ID пользователя из сессии
		session.Values["author"] = nil          // Очистка имени пользователя из сессии
		err := session.Save(r, w)
		if err != nil {
			log.Fatal(err)
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

		userId := data.UserId

		session, err := store.Get(r, "user-session") // Получаем сессию из запроса

		// Проверяем значение аутентифицированности в сессии
		auth, ok := session.Values["authenticated"].(bool)

		if !auth || !ok {
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
		} else if session.Values["authenticated"].(bool) && session.Values["userId"].(uint) == userId {
			// Парсим шаблон из файла
			tmpl, err := template.ParseFiles("templates/authPost.html")
			if err != nil {
				log.Fatal(err)
			}

			// Запускаем шаблон, передаём в него данные
			err = tmpl.Execute(w, data)
			if err != nil {
				log.Fatal(err)
			}
		} else {
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

		// Получение или создание сессии
		session, _ := store.Get(r, "user-session")

		// Аутентификация успешна, установка флага аутентификации в сессии
		if session.Values["authenticated"] == true {
			// Получаем ID пользователя из сессии
			userId := session.Values["userId"].(uint)

			// Получаем имя пользователя из сессии
			author := session.Values["author"].(string)

			// Создаём объект структуры
			post := Post{Title: title, Author: author, UserId: userId, Body: body}

			// Передаём объект в функцию создания новой записи в БД
			err := CreatePost(db, post)
			if err != nil {
				log.Fatal(err)
			}

			// Перенаправляем на страницу со всеми постами
			http.Redirect(w, r, "/posts", http.StatusPermanentRedirect)
		} else {
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		}
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

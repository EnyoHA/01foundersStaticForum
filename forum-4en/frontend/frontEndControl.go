package frontend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/backend"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Drum struct {
	Base *backend.Base
}

func FindIP(r *http.Request) string {
	address := r.Header.Get("X-FORWARDED-FOR")
	if address != "" {
		return address
	}

	return r.RemoteAddr
}

func (drum *Drum) StartPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}

	switch r.Method {
	default:
		http.Error(w, "400 Bad Request 400", http.StatusBadRequest)
	case "POST":
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "application/text")
		w.Write([]byte("Registered successfully - You may now log in"))
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "startpage", "")
	}
}

func (drum *Drum) Homepage(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Cookies interface{}
		Posts   interface{}
	}
	var pagePres pageData
	web, err := template.ParseFiles("templates/homepage.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	sortBy := r.FormValue("sortBy")
	if err != nil {
		pagePres = pageData{
			Cookies: err.Error(),
			Posts:   drum.Base.PostIndex(sortBy, ""),
		}
		if err := web.Execute(w, pagePres); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

	}
	files := GetTemplates()
	RenderTemplate(w, r, files, "Homepage", drum.Base.PostIndex(sortBy, ""))
}

func (drum *Drum) FEndPosts(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Cookies interface{}
		Posts   interface{}
	}
	var pagePres pageData

	web, err := template.ParseFiles("templates/frontendcat.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	sortBy := r.FormValue("sortBy")
	if err != nil {
		pagePres = pageData{
			Cookies: err.Error(),
			Posts:   drum.Base.PostIndex(sortBy, ""),
		}
		if err := web.Execute(w, pagePres); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

	}
	files := GetTemplates()
	RenderTemplate(w, r, files, "FrontEndCat", drum.Base.FrontEndPosts(sortBy))
}

func (drum *Drum) BEndPosts(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Cookies interface{}
		Posts   interface{}
	}
	var pagePres pageData

	web, err := template.ParseFiles("templates/backendcat.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	sortBy := r.FormValue("sortBy")
	if err != nil {
		pagePres = pageData{
			Cookies: err.Error(),
			Posts:   drum.Base.PostIndex(sortBy, ""),
		}
		if err := web.Execute(w, pagePres); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

	}
	files := GetTemplates()
	RenderTemplate(w, r, files, "BackEndCat", drum.Base.BackEndPosts(sortBy))
}

func (drum *Drum) FilterByCategory(w http.ResponseWriter, r *http.Request) {
	cartegory := r.FormValue("category")
	log.Println(cartegory)
	files := GetTemplates()
	if r.Method != http.MethodPost {
		RenderTemplate(w, r, files, "Error", "ERROR")
		return
	}
	cat,err:=strconv.Atoi(cartegory)
	if err!=nil{
		RenderTemplate(w, r, files, "Error", "ERROR")
		return
	}
	RenderTemplate(w, r, files, "Category", drum.Base.FilterByCategory(cat))
}

func (drum *Drum) UsersPosts(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Cookies interface{}
		Posts   interface{}
	}
	var pagePres pageData
	cky, _ := r.Cookie("session")

	currentUser := GetCurrentUser(w, r, cky)
	web, err := template.ParseFiles("templates/yourposts.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	sortBy := r.FormValue("sortBy")
	if err != nil {
		pagePres = pageData{
			Cookies: err.Error(),
			Posts:   drum.Base.PostIndex(sortBy, ""),
		}
		if err := web.Execute(w, pagePres); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

	}
	files := GetTemplates()
	RenderTemplate(w, r, files, "YourPosts", drum.Base.YourPosts(sortBy, currentUser.UserID))
}

func (drum *Drum) LikedPosts(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Cookies interface{}
		Posts   interface{}
	}
	var pagePres pageData
	cky, _ := r.Cookie("session")

	currentUser := GetCurrentUser(w, r, cky)

	web, err := template.ParseFiles("templates/likedposts.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err != nil {
		pagePres = pageData{
			Cookies: err.Error(),
			// To fix
			Posts:   drum.Base.YourLikedPosts("",currentUser.UserID),
		}
		if err := web.Execute(w, pagePres); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

	}
	files := GetTemplates()
	RenderTemplate(w, r, files, "LikedPosts", drum.Base.YourLikedPosts("",currentUser.UserID))
}

func (drum *Drum) PostComments(w http.ResponseWriter, r *http.Request) {

	_, err := template.ParseFiles("templates/comments.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error1", http.StatusInternalServerError)
		return
	}

	c, _ := r.Cookie("session")

	validCookie := drum.IsCookieValid(w, c)

	user := GetCurrentUser(w, r, c)
	postID := r.URL.Query().Get("postID")
	log.Println("The user id 2", user.UserID)

	commentBody := r.FormValue("comment")
	log.Println("the comment ", commentBody)

	if commentBody != "" {
		drum.Base.CommentComment(validCookie[0], postID, commentBody)
	}

	result := map[string]interface{}{
		"postId":   postID,
		"comments": drum.Base.CommentIndex(postID),
	}
	files := GetTemplates()
	RenderTemplate(w, r, files, "Comments", result)
}

func (drum *Drum) Register(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}

	switch r.Method {
	default:
		http.Error(w, "400 Bad Request 400", http.StatusBadRequest)
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "Register", "")
	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")
		if username == "" || password == "" || email == "" {
			http.Error(w, "400 Bad Request 400", http.StatusBadRequest)
			return
		}

		_, _, _, err := drum.Base.Register(username, email, password)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-type", "application/text")
			w.Write([]byte("0" + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "application/text")

		files := GetTemplates()
		RenderTemplate(w, r, files, "Register", "Registered successfully - You may now log in")
	}
}
func (drum Drum) IsLoggedIn(w http.ResponseWriter, r *http.Request) {
	files := GetTemplates()
	if r.Method != http.MethodGet {
		RenderTemplate(w, r, files, "Error", "ERROR bad request")
		return
	}
	cky, err := r.Cookie("session")
	if err != nil {
		log.Println("Error COOKIE", err.Error())
		w.Write([]byte("NOT LOGGED"))
		return
	}

	cook := drum.IsCookieValid(w, cky)
	user,ok:=drum.Base.GetUser(cook[0])
	if !ok{
		log.Println("Error USER COOKIE", err.Error())
		w.Write([]byte("NOT LOGGED"))
		return
	}
	user_id := r.URL.Query().Get("user_id")
	post_id := r.URL.Query().Get("post_id")
	//To do handle error
	reactions,_:= drum.Base.GetReactionsByPostID(post_id)

	for i:=range reactions{
		log.Println(reactions[i].UserID,user.UserID,"here")
		if strconv.Itoa(user.UserID)==reactions[i].UserID{
			w.Write([]byte("HAS LIKED"))
			return
		}
	}
	
	// if user.HasLiked !="0"{
	// 	w.Write([]byte("HAS LIKED"))
	// 	return
	// }
	log.Println(user_id)
}

func (drum *Drum) MyCrewIsLoggingOn(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}

	switch r.Method {
	default:
		http.Error(w, "400 bad boy request 400", http.StatusBadRequest)
	case "POST":
		nameUser := r.FormValue("username")
		password := r.FormValue("password")
		if nameUser == "" || password == "" {
			http.Error(w, "400 bad request 400", http.StatusBadRequest)
			return
		}
		_, _, sessionID, err := drum.Base.LoginUser(nameUser, password)
		if err != nil {
			log.Println("[ERROR LOGIN]",err)
			w.Header().Set("Content-Type", "application/text")
			w.Write([]byte("USER NOT FOUND" + err.Error()))
			return
		}

		c := &http.Cookie{
			Name: "session",

			Value: sessionID,

			Path: "/",
		}
		c.MaxAge = 30000
		http.SetCookie(w, c)
		backend.SessionDB[c.Value] = backend.Session{}
		http.Redirect(w, r, "http://localhost:8080", http.StatusSeeOther)
		return

	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "LogIn", "")
	}
}

func (drum *Drum) LogOut(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}

	switch r.Method {
	default:
		http.Error(w, "400 bad request 400", http.StatusBadRequest)
		return
	case "GET":
		files := GetTemplates()
		cky, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp := cky.Value
		err = drum.Base.DeleteSession(resp)
		if err != nil {
			log.Fatal(err)
		}
		// Delete the cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   "",
			Expires: time.Now(),
		})
		RenderTemplate(w, r, files, "LoggedOut", "")
	case "POST":
		cky, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := cky.Value
		err = drum.Base.DeleteSession(resp)
		if err != nil {
			log.Fatal(err)
		}
		// Delete the cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   "",
			Expires: time.Now(),
		})
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("Logged out succesfully"))
	}
}

func (drum *Drum) MakePost(w http.ResponseWriter, r *http.Request) {
	// .Request) {
		// var user backend.User
		var filterBy [2]string
		var allFilters string = ""
		if r.URL.Path != "/post" {
			http.Error(w, "404 page not found 404", http.StatusNotFound)
			return
		}
	
		cky, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	
		// need to check cookies validity, new func of course :(
		cook := drum.IsCookieValid(w, cky)
		if len(cook) == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		switch r.Method {
		default:
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
		case "GET":
			files := GetTemplates()
			RenderTemplate(w, r, files, "Post", "")
			return
		case "POST":
			title := r.FormValue("title")
			post := r.FormValue("body")
	
			// userID := /
			filterBy[0] = r.FormValue("frontendcat")
			filterBy[1] = r.FormValue("backendcat")
			for _, filters := range filterBy {
				switch filters {
				case "FrontEnd":
					allFilters += "frontEnd" + "/"
				case "BackEnd":
					allFilters += "backEnd" + "/"
				}
			}
			currentUser := GetCurrentUser(w, r, cky)
	
			// currentUser := user
			_, err := drum.Base.PostPost(title, filterBy, post, strconv.Itoa(currentUser.UserID))
	
			if err != nil {
				log.Println("PostPost RETURN ERROR",err)
	
			}
	
		}
		// Redirect to homepage.
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request, c *http.Cookie) backend.User {
	db, _ := sql.Open("sqlite3", "forum.db")
	defer db.Close()

	rows, err := db.Query("SELECT * FROM Session WHERE sessionID=?;", c.Value)
	sess := QuerySession(rows, err)
	rows.Close() //good habit to close

	if sess.SessionID != c.Value {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	// Get current user username.
	id, _ := strconv.Atoi(sess.UserID)
	rows2, err2 := db.Query("SELECT * FROM User WHERE userID=?;", id)
	currentUserData := QueryUser(rows2, err2)
	rows.Close() //good habit to close
	return currentUserData
}

func QuerySession(rows *sql.Rows, err error) backend.Session {
	// Variables for line after for rows.Next()
	var sessionID string
	var userID string

	var sess backend.Session
	// Scan all the data from that row.
	for rows.Next() {
		err = rows.Scan(&sessionID, &userID)
		temp := backend.Session{
			SessionID: *&sessionID,
			UserID:    *&userID,
		}
		sess = temp
	}
	rows.Close()
	return sess
}

func QueryUser(rows *sql.Rows, err error) backend.User {

	var id int
	var username string
	var email string
	var password string
	var sessId string
	var loggedIn interface{}

	var usr backend.User

	for rows.Next() {
		err = rows.Scan(&id, &username, &email, &password, &sessId, &loggedIn)
		if err != nil {
			log.Println("QUERY USER ERROR",err)
			break
		}
		temp := backend.User{
			UserID:   id,
			Username: username,
			Email:    email,
			Password: password,
		}
		usr = temp
	}
	rows.Close()
	return usr
}

func (drum *Drum) WriteComment(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/comment" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}

	c, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	cook := drum.IsCookieValid(w, c)
	if len(cook) == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch r.Method {
	default:
		http.Error(w, "400 Bad Boi Requested 400", http.StatusBadRequest)
	case "GET":
		files := GetTemplates()
		postID := r.URL.Query().Get("postid")
		RenderTemplate(w, r, files, "Comment", postID)
	case "POST":
		body := r.FormValue("body")

		user := GetCurrentUser(w, r, c)
		postID := r.URL.Query().Get("postid")
		_, err := drum.Base.CommentComment(strconv.Itoa(user.UserID), postID, body)
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		files := GetTemplates()
		RenderTemplate(w, r, files, "Comment", "Thanks for contributing to the discurssion")
	}
}

// has to be a new func as you can do it using http.SetCookie(something, somethingElse) but I don't think
// you can call that into other funcs
// eg something.Valid is a way to check within a func
// https://github.com/golang/go/issues/46370

func (drum *Drum) IsCookieValid(w http.ResponseWriter, c *http.Cookie) []string {
	cky := []string{}
	if strings.Contains(c.String(), "&") {
		cky = strings.Split(c.Value, "&")
	} else if c != nil {
		cky = append(cky, c.Value)
	} else {
		return cky
	}
	if len(cky) != 0 {
		if !(drum.Base.IsSessionValid(cky[0])) {
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   "",
				Expires: time.Now(),
			})
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/text")
			w.Write([]byte("OI BLUD you're not signed in geeeeeezer"))
		} else {

			return cky
		}
	}
	return cky
}

// Render template on get request.
func RenderTemplate(w http.ResponseWriter, r *http.Request, files []string, templateName string, data interface{}) {

	tmplSet, err := template.ParseFiles(files...)
	if err != nil {
		fmt.Print(err)
	}

	err = tmplSet.ExecuteTemplate(w, templateName, data)
	if err != nil {
		fmt.Print(err)
	}
}

func GetTemplates() []string {
	files := []string{}
	folder, _ := ioutil.ReadDir("templates")
	for _, subitem := range folder {
		files = append(files, "./templates/"+subitem.Name())
	}

	return files
}

func (drum *Drum) Likes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/like" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}

	c, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	cook := drum.IsCookieValid(w, c)
	if len(cook) == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	reaction := backend.Reaction{}
	if err := json.NewDecoder(r.Body).Decode(&reaction); err != nil {
		log.Println(err)
		return
	}
	user,ok:=drum.Base.GetUser(cook[0])
	if!ok{
		log.Println("CREATE LIKE ERROR",err)
		return
	}
	reaction.UserID=strconv.Itoa(user.UserID)
	 reactionID, err := drum.Base.ReactToPost(reaction)
	 log.Println(c,reaction.UserID)
	if err != nil {
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/text")
	w.Header().Set("reactionID", reactionID)

}

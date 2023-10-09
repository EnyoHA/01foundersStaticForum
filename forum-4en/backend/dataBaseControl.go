package backend

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// some global varibles to talk to some of the structs
var SessionDB = map[string]Session{}
var UserDB = map[string]User{}

// make the structs to call into, need to relate to the database
type Base struct {
	DB *sql.DB
}

// make them all strings so easier to read and post to

// struct to keep track of the users details
type User struct {
	UserID    int
	Username  string
	Email     string
	Password  string
	SessionID string
	LoggedIn  string
}

type Session struct {
	SessionID string
	UserID    string
}

type Post struct {
	PostID      string
	UserID      string
	CreatedDate string
	Title       string
	Body        string
	Category    int
	Category2   int
	NumComments int
	Comments    []Comment
	Reactions   Reaction
}

type Comment struct {
	commentID   string
	UserID      string
	postID      string
	createdDate string
	body        string
	reactions   Reaction
}

type Reaction struct {
	UserID      string `json:"user_ID"`
	PostID      string `json:"post_ID"`
	ReactID     string `json:"reaction_ID"`
	UpVotes     int    `json:"likes"`
	CommentID   string `json:"comment_ID"`
	NumOfReacts int    `json:"numberOfReactions"`
	DownVotes   int    `json:"dislike"`
}

// a lot of the functions here end up being the same as its just
// inserting the info into different databases,
// will then need to add other functions to handle the rest
// eg update and delete

// UUID seems easier to use/no harder than learning the packages within go
// call NewV4 as it adds a random Unique identifier
// https://pkg.go.dev/github.com/satori/go.uuid

// doesnt allow this as a constant, is there another way?
// const date = time.Now().Format("2004.04.20 04:20:00")

/* only use (base *Base) and not (base *Base, user *User) as there can only be
one reciever in each function. Could get round it and simplify code if using a
JSON as well as a SQL database
*/

// HashPassword turns the password into a hashed string
func CreateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

// CheckPasswordHash checks the entered password against the hashed password
func CheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (base *Base) GetUser(SessionID string) (User, bool) {

	var user User
	// db, err := sql.Open("sqlite3", "forum.db")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer db.Close()
	if SessionID != "" {
		err := base.DB.QueryRow("SELECT * FROM User WHERE sessionID = '" + SessionID + "'").Scan(&user.UserID,&user.Username,&user.Email,&user.Password,&user.SessionID,&user.LoggedIn)
		if err == sql.ErrNoRows {
			return user, false
		} else {
			return user, true
		}
	} else {
		return user, false
	}
}
func (base *Base) GetUserByID(UserID string) (User, bool) {
	var user User
	if UserID != "" {
		err := base.DB.QueryRow("SELECT userID,username,email,password,sessionID,ifnull(loggedIn, '') FROM User WHERE userID = '"+UserID+"'").Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.SessionID, &user.LoggedIn)
		if err == sql.ErrNoRows {
			log.Println(err.Error())
			return user, false
		} else if err != nil {
			log.Println(err.Error())
			return user, false
		} else {
			return user, true
		}
	} else {
		return user, false
	}
}

// more general functions.

// date as a global variable?
var date = time.Now().Format("2004.04.20 04:20:00")

// function to call into others to update the database info
func (base *Base) Update(table, set, to, where, id string) error {
	log.Println("Updating "+table)
	update := "UPDATE " + table + " SET " + set + " = '" + to + "' WHERE " + where + " = '" + id + "'"
	basedata, err := base.DB.Prepare(update)
	if err!=nil{
		log.Println("[UPDATE ERROR]",err)
		return err
	}
	_, err = basedata.Exec()
	if err != nil {
		fmt.Println("UPDATE ERROR: ", err)
		return err
	}
	log.Println(base.GetUserByID(id))
	return nil
}

// https://www.tutorialspoint.com/sqlite/sqlite_delete_query.htm
// delete data from the database
func (base *Base) Delete(table, where, value string) error {
	// NEED TO PUT IN SPACES SO IT EXECUTES CORRECTLY
	remove := "DELETE FROM " + table + " WHERE " + where
	basedata, err := base.DB.Prepare(remove + " = (?)")
	if err != nil {
		return err
	}
	_, err = basedata.Exec(value)
	if err != nil {
		return err
	}
	return nil
}

// registration function
func (base *Base) Register(username, email, passw string) (string, string, string, error) {
	// date := time.Now().Format("2004.04.20 04:20:00")
	// how to do this not using UUID? use the built in UUID package but
	// doesnt seem to be easier? Not sure how to track. Cookies package?

	basedata, err := base.DB.Prepare(`
		INSERT INTO User (username, email, password,sessionID,loggedIn) values (?, ?, ?,?,?)
	`)
	if err != nil {
		return "", "", "", err
		// log.Fatal(err)
	}
	_, err = basedata.Exec(username, email, passw,"","")
	if err != nil {
		// log.Fatal(err)
		return "", "", "", err
	}
	return "", "", "", nil
}

// func to start the session
func (base *Base) StartSession(userID string) (string, error) {
	fmt.Println("START SESSION")
	// date := time.Now().Format("2004.04.20 04:20:00")
	sessionID := uuid.NewV4()
	// use quotes around things as seem to get an error for some things
	// maybe just because of the other SQL install you tried?
	// a way to avoid doing this for all funcs?
	basedata, _ := base.DB.Prepare(`
		INSERT INTO Session (sessionID, userID) values (?, ?)
	`)
	mu:=sync.Mutex{}
	mu.Lock()
	_, err := basedata.Exec(sessionID, userID)
	if err != nil {
		// needs two values so one can be a blank string
		log.Println("START SESSION ERROR: ", err)
		// why does this make it not work?
		// just because its log.Fatal
		return "", err
	}
	mu.Unlock()
	base.Update("User", "sessionID", sessionID.String(), "userID", userID)
	return sessionID.String(), nil
}

func (base *Base) IsSessionValid(sessionID string) bool {
	var sessuuid string
	err := base.DB.QueryRow("SELECT * FROM Session WHERE sessionID = '" + sessionID + "'").Scan(&sessuuid)
	if err == sql.ErrNoRows {
		log.Print("IsSessionValid",err)
		return false
	}
	// initiate a new variable to compare against the sessionID
	// var inputedSession string
	// for horizontal.Next() {
	// 	horizontal.Scan(&inputedSession)
	// }
	return true
}

// delete session info using the previous function and sessionID
// https://www.sqlitetutorial.net/sqlite-delete/
// https://stackoverflow.com/questions/68322484/how-to-delete-row-in-go-sqlite3
func (base *Base) DeleteSession(sessionID string) error {
	// little unsure about the multiple calls of sessionID
	// table = user, where = sessionID, rest?
	err := base.Update("User", "sessionID", "", "sessionID", sessionID)
	if err != nil {
		return err
	}
	// table = session, where = sessionID, value = sessionID
	err = base.Delete("Session", "sessionID", sessionID)
	if err != nil {
		return err
	}
	return nil
}

// checks the database for the info, not disimilar from the first thing I tried
func (base *Base) LoginUser(userName, passw string) (string, string, string, error) {
	// establish a variable that talks to the struct
	var users User
	// cant use "" for User will it be alright without?
	userRow, err := base.DB.Query("SELECT * FROM User WHERE username = '" + userName + "'")
	if err != nil {
		return "", "", "", errors.New("UNABLE TO QUERY DATABASE"+err.Error())
	}
	// more variables to talk to the struct
	// then relate them to each part of the struct
	// var usID, sesID, usNm, eMa, CreatedDate, pass string
	for userRow.Next() {
		if err:=userRow.Scan(&users.UserID, &users.Username, &users.Email, &users.Password, &users.SessionID, &users.LoggedIn);err!=nil{
			return "", "", "", errors.New("LOGIN QUERY ERROR: "+err.Error())
		}
	}

	// if the entry doesnt match the database
	// return the error
	if users.Username == "" {
		return "", "", "", errors.New("USER NOT FOUND")
	}
	// checks the entered password against the hashed one
	// 24/07 not matching passwords correctly
	// if !(CheckHash(passw, users.password)) {
	if passw != users.Password {
		return "", "", "", errors.New("UNMATCHED PASSWORDS")
	}
	// if not a new session/blank remove it
	// 29/07 is this why you're autologged out when clicking on something?
	if users.SessionID != "" {
		log.Println("SESSION ID IS",users.SessionID)
		base.DeleteSession(users.SessionID)
	}
	// then create a new session
	seshion, err := base.StartSession(strconv.Itoa(users.UserID))
	if err != nil {
		log.Println("Error starting session",err)
		return "", "", "", err
	}
	// pass this new session into the struct
	users.SessionID = seshion
	return strconv.Itoa(users.UserID), users.Username, users.SessionID, nil
}

// funcs to deal with posts

// similar to the other funcs now
func (base *Base) PostPost(title string, category [2]string, body string, userId string) (string, error) {

	postID := uuid.NewV4()
	basedata, err := base.DB.Prepare(`
	INSERT INTO Post (postID, userID, title, frontendcat, backendcat, datePosted, body) values (?, ?, ?, ?, ?, datetime('now','localtime'), ?)
	`)
	if err != nil {
		return "", err
	}
	_, err = basedata.Exec(postID, userId, title, category[0], category[1], body)
	if err != nil {
		return "", errors.New("POST QUERY ERROR: "+err.Error())
	}
	return postID.String(), nil
}

// two funcs for reactions? Can I make it one? Better to have two to distinguish
// basically the same code anyway just one pointing to posts other comments
// do I need to reaction database tables?
func (base *Base) ReactToPost(reaction Reaction) (string, error) {
	reactionID := uuid.NewV4()
	row := base.DB.QueryRow(fmt.Sprintf(`
	SELECT * FROM Reaction WHERE userID="%s" AND postID="%s"
	`, reaction.UserID, reaction.PostID))

	if row.Scan() != sql.ErrNoRows {
		basedata, _ := base.DB.Prepare(`
		UPDATE Reaction SET total_reactions=?,likes =?, Dislikes=? WHERE postID=? AND commentID=? AND userID=? 
		`)
		_, err := basedata.Exec(reaction.NumOfReacts, reaction.UpVotes, reaction.DownVotes, reaction.PostID, reaction.CommentID, reaction.UserID)
		if err != nil {
			log.Fatal(err)
		}
		return "", nil
	}

	basedata, _ := base.DB.Prepare(`
	INSERT OR REPLACE  INTO Reaction (reactionID, postID,commentID,userID, total_reactions,likes,Dislikes) values (?, ?, ?, ?,?,?,?)
	`)
	_, err := basedata.Exec(reactionID, reaction.PostID, reaction.CommentID, reaction.UserID, reaction.NumOfReacts, reaction.UpVotes, reaction.DownVotes)
	if err != nil {
		log.Fatal(err)
	}
	return reactionID.String(), nil
}

// same func as IsCommentReactionValid
func (base *Base) IsPostReactionValid(posID, usID string) (string, int) {
	var reaction Reaction
	// var ReactionID, PostID, UserID string
	// var reactions int
	// maybe shouldnt do "" quotes? will test
	// dont use `` either?
	horizontal, err := base.DB.Query("SELECT reactionID, postID, userID, react FROM Reaction WHERE postID = '" + posID + "' AND userID = '" + usID + "' AND commentID IS NULL")
	// handle the error
	if err != nil {
		log.Println("IsPostReactionValid",err)
		return "", 0
	}
	// scans
	for horizontal.Next() {
		horizontal.Scan(&reaction.ReactID, &reaction.PostID, &reaction.UserID, &reaction.NumOfReacts)
	}
	return reaction.ReactID, reaction.NumOfReacts
}

// last func to make, basically the backend to homepage
// not last func as needed the other indexes
// need to select from the database by row to capture all the info
func (base *Base) PostIndex(sortBy, usID string) []map[string]interface{} {
	// var post Post
	var posts []map[string]interface{}
	postRows, err := base.DB.Query("SELECT * FROM Post ORDER BY datePosted DESC")
	if err != nil {
		log.Println("PostIndex",err)
		return posts
	}

	var posID, uID, Title, subForum, dateCreated, content, subForum2 interface{}
	for postRows.Next() {
		err = postRows.Scan(&posID, &uID, &Title, &subForum, &subForum2, &dateCreated, &content)
		reactions := base.PostReactionIndex(posID.(string))
		upCount := 0
		downCount := 0
		hasUP := false
		hasDown := false
		for _, rea := range reactions {
			if rea.UpVotes > 0 {
				upCount++
			}
			if rea.DownVotes > 0 {
				downCount++
			}
			if rea.UserID == usID {
				if rea.UpVotes > 0 {
					hasUP = true
				}
				if rea.DownVotes > 0 {
					hasDown = true
				}
			}
		}
		posts = append(posts, map[string]interface{}{
			"postID":     posID,
			"userID":     uID,
			"title":      Title,
			"category":   subForum,
			"category2":  subForum2,
			"datePosted": dateCreated,
			"body":       content,
			"upvotes":    upCount,
			"downvotes":  downCount,
			"hasUP":      hasUP,
			"hasDOWN":    hasDown,
		})
		if err != nil {
			log.Println(err.Error())
		}

	}
	return posts
}

func (base *Base) YourPosts(sortBy string, usID int) []map[string]interface{} {
	var posts []map[string]interface{}

	postRows, err := base.DB.Query("SELECT * FROM Post WHERE userID =?", usID)
	if err != nil {
		log.Println("YourPosts",err)
		return posts
	}

	var posID, uID, Title, subForum, dateCreated, content, subForum2 interface{}
	for postRows.Next() {
		err = postRows.Scan(&posID, &uID, &Title, &subForum, &subForum2, &dateCreated, &content)
		reactions := base.PostReactionIndex(posID.(string))
		upCount := 0
		downCount := 0
		hasUP := false
		hasDown := false
		for _, rea := range reactions {
			if rea.UpVotes > 0 {
				upCount++
			}
			if rea.DownVotes > 0 {
				downCount++
			}
			if rea.UserID == fmt.Sprint(usID) {
				if rea.UpVotes > 0 {
					hasUP = true
				}
				if rea.DownVotes > 0 {
					hasDown = true
				}
			}
		}
		posts = append(posts, map[string]interface{}{
			"postID":     posID,
			"userID":     uID,
			"title":      Title,
			"category":   subForum,
			"category2":  subForum2,
			"datePosted": dateCreated,
			"body":       content,
			"upvotes":    upCount,
			"downvotes":  downCount,
			"hasUP":      hasUP,
			"hasDOWN":    hasDown,
		})
		if err != nil {
			log.Println(err.Error())
		}

	}
	return posts
}
func (base *Base)FrontEndPosts(sort string)string{
	return "To FIX"
}
func (base *Base)BackEndPosts(sort string)string{
	return "To FIX"
}

func (base *Base) YourLikedPosts(sortBy string, usID int) []map[string]interface{} {
	var posts []map[string]interface{}

	postRows, err := base.DB.Query(`SELECT * FROM Post WHERE Post.postID in (SELECT postID FROM Reaction WHERE userID=?)`, usID)
	if err != nil {
		fmt.Print(err)
		return posts
	}
	var posID, uID, Title, subForum, dateCreated, content, subForum2 interface{}
	for postRows.Next() {
		err := postRows.Scan(&posID, &uID, &Title, &subForum, &subForum2, &dateCreated, &content)
		reactions := base.PostReactionIndex(posID.(string))
		upCount := 0
		downCount := 0
		hasUP := false
		hasDown := false
		for _, rea := range reactions {
			if rea.UpVotes > 0 {
				upCount++
			}
			if rea.DownVotes > 0 {
				downCount++
			}
			if rea.UserID == fmt.Sprint(usID) {
				if rea.UpVotes > 0 {
					hasUP = true
				}
				if rea.DownVotes > 0 {
					hasDown = true
				}
			}
		}
		posts = append(posts, map[string]interface{}{
			"postID":     posID,
			"userID":     uID,
			"title":      Title,
			"category":   subForum,
			"category2":  subForum2,
			"datePosted": dateCreated,
			"body":       content,
			"upvotes":    upCount,
			"downvotes":  downCount,
			"hasUP":      hasUP,
			"hasDOWN":    hasDown,
		})
		if err != nil {
			log.Println(err.Error())
		}
		fmt.Println(posts)
	}
	// }

	return posts
}
func (base *Base)GetReactionsByPostID(postID string)([]Reaction,error){
	var reactions []Reaction
	reactionRows,err:= base.DB.Query(fmt.Sprintf("SELECT * FROM Reaction WHERE postID ='%s'", postID))
	if err!=nil{
			log.Println("[error GetReactionsByPostID]",err)
			return reactions,err
	}
	var reaction Reaction
	for reactionRows.Next(){
		err=reactionRows.Scan(&reaction.ReactID,&reaction.PostID,&reaction.CommentID,&reaction.UserID,&reaction.NumOfReacts,&reaction.UpVotes,&reaction.DownVotes)
		if err!=nil{
			log.Println("[error GetReactionsByPostID]",err)
			return reactions,err
		}
		reactions = append(reactions, reaction)
	}
	return reactions,nil
}

func (base *Base) FilterByCategory(category int) []Post {
	var posts []Post
	postRows, err := base.DB.Query(fmt.Sprintf("SELECT * FROM Post WHERE frontendcat ='%d' OR backendcat = '%d'", category, category))
	if err != nil {
		log.Println("FilterByCategory",err)
		return posts
	}
	var post Post
	for postRows.Next() {
		err = postRows.Scan(&post.PostID, &post.UserID, &post.Title, &post.Category, &post.Category2, &post.CreatedDate, &post.Body)
		reactions := base.PostReactionIndex(post.PostID)
		if len(reactions) > 0 {
			upCount := 0
			downCount := 0
			for _, rea := range reactions {
				if rea.UpVotes > 0 {
					upCount++
				}
				if rea.DownVotes > 0 {
					downCount++
				}
			}
			reactions[0].UpVotes = upCount
			reactions[0].DownVotes = downCount
			reactions[0].NumOfReacts = upCount + downCount
			post.Reactions = *reactions[0]
		}

		if err != nil {
			log.Println("FilterByCategory",err.Error())
		}
		posts = append(posts, post)
	}
	return posts
}

// need 2 reaction indexes one for posts the other for comments

func (base *Base) PostReactionIndex(posID string) []*Reaction {
	res := make([]*Reaction, 0)
	// only two kinds of reaction so don't need the array here
	reactRows, err := base.DB.Query("SELECT reactionID, postID, commentID,userID, total_reactions,likes,Dislikes FROM Reaction WHERE postID = '" + posID + "' ")
	if err != nil {
		log.Println("ERROR POSTREACTION INDEX: ",err.Error())
		return res
	}

	for reactRows.Next() {
		var reaction Reaction
		reactRows.Scan(&reaction.ReactID, &reaction.PostID, &reaction.CommentID, &reaction.UserID, &reaction.NumOfReacts, &reaction.UpVotes, &reaction.DownVotes)
		res = append(res, &reaction)
	}
	return res
}


// funcs to deal with Comments

func (base *Base) CommentComment(userID, postID, body string) (string, error) {
	// date := time.Now().Format("2004.04.20 04:20:00")
	log.Printf("values logging into the database are userID: %s postID: %s body: %s ", userID, postID, body)
	commentID := uuid.NewV4()
	basedata, err := base.DB.Prepare(`
	INSERT INTO Comment (commentID, userID, postID, createdDate, body) values (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return "", err
	}
	result, err := basedata.Exec(commentID, userID, postID, date, body)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected < 1 {
		if err != nil {
			panic(fmt.Errorf("Error: The error here is ", err))
		}
		return "", fmt.Errorf("Invald query %v", err)
	}

	return commentID.String(), nil
}

// two funcs for reactions? Can I make it one? Better to have two to distinguish
// basically the same code anyway just one pointing to posts other comments
// do I need to reaction database tables?
func (base *Base) ReactToComment(postID, commentID string, user User, reacted int) (string, string, error) {
	reactionID := uuid.NewV4()
	basedata, _ := base.DB.Prepare(`
	INSERT INTO Reaction (reactionID, postID, commentID, userID, reacted) values (?, ?, ?, ?, ?)
	`)
	_, err := basedata.Exec(reactionID, postID, commentID, user.UserID, reacted)
	if err != nil {
		log.Fatal(err)
	}
	return "", reactionID.String(), nil
}

// create an index for the comments to return them in each post
// similar to creating an index for the posts in the homepage
func (base *Base) CommentIndex(poID string) []map[string]interface{} {
	// var comment Comment
	var comments []map[string]interface{}
	// var comRows *sql.Rows
	// var err error
	comRows, err := base.DB.Query("SELECT * FROM Comment WHERE postID = ?", poID)
	if err != nil {
		fmt.Println("CommentIndex",err)
		return comments
	}

	var comID, postID, usID, dateCreated, content string
	for comRows.Next() {
		err := comRows.Scan(&comID, &postID, &usID, &dateCreated, &content)
		if err != nil {
			panic(err)
		}
		comments = append(comments, map[string]interface{}{
			"commentID":   comID,
			"postID":      postID,
			"userID":      usID,
			"createdDate": dateCreated,
			"body":        content,
			// another for reactions need to do the func
		})
		log.Printf("The values respectively %s %s %s %s", comID, postID, usID, dateCreated, content)
		if err != nil {
			log.Println(err.Error())
		}

	}
	return comments
}

func StartDatabase(db *sql.DB) *Base {
	// createUser(db)
	// createSession(db)
	// createPost(db)
	// createComment(db)
	// createReaction(db)
	return &Base{
		DB: db,
	}
}

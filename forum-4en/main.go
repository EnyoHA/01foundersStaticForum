package main

// massive issues with cyclical imports
// make it one massive main file?
// that is so ugly though
// https://stackoverflow.com/questions/26942150/importing-go-files-in-same-folder
// still causing issues and I don't quite understand why may do one big main
// https://stackoverflow.com/questions/14155122/how-to-call-function-from-another-file-in-go
// I swear this is what I'm doing
import (
	"database/sql"
	"fmt"
	"forum/backend"
	"forum/frontend"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func Init(db *sql.DB) {
	// "userID"			INT PRIMARY KEY AUTOINCREMENT
	db.Exec(`
	CREATE TABLE IF NOT EXISTS "User" (
		"userID"		INTEGER PRIMARY KEY AUTOINCREMENT,
		"username"		VARCHAR(64) NOT NULL UNIQUE,
		"email"			TEXT NOT NULL UNIQUE,
		"password"		TEXT NOT NULL,
		"sessionID"		TEXT NOT NULL,
		"loggedIn"		TEXT,
		FOREIGN KEY (sessionID) REFERENCES "Session" ("sessionID")
	);
	 `)
	//  "ipAddress"	TEXT NOT NULL,
	// 	"timeOfLog"	TEXT NOT NULL,
	// 	"identity"	TEXT NOT NULL,
	db.Exec(`
	CREATE TABLE IF NOT EXISTS "Session" (
		"sessionID"	TEXT PRIMARY KEY,
		"userID"	INTEGER NOT NULL,
		FOREIGN KEY (userID)
			REFERENCES "User" ("userID")
	);
	`)
	// add username so it can be shown on posts
	db.Exec(`
	CREATE TABLE IF NOT EXISTS "Post" (
		"postID"	TEXT UNIQUE NOT NULL PRIMARY KEY,
		"userID"	TEXT NOT NULL,
		"title"     TEXT NOT NULL,
		"frontendcat"	INTEGER,
		"backendcat"	INTEGER,
		"datePosted" TEXT NOT NULL,
		"body"	TEXT NOT NULL,
		FOREIGN KEY ("userID")
			REFERENCES "User" ("userID")
	);
	`)
	// add username so it can be shown on comments
	db.Exec(`
	CREATE TABLE IF NOT EXISTS "Comment" (
		"commentID" 	TEXT UNIQUE NOT NULL PRIMARY KEY,
		"postID"		TEXT NOT NULL,
		"userID"		TEXT NOT NULL,
		"createdDate" 	TEXT NOT NULL,
		"body"			TEXT NOT NULL,
		FOREIGN KEY ("postID")
			REFERENCES "Post" ("postID")
		FOREIGN KEY ("userID")
			REFERENCES "User" ("userID")
	);
	`)
	db.Exec(`
	CREATE TABLE IF NOT EXISTS "Reaction" (
		"reactionID" TEXT NOT NULL PRIMARY KEY,
		"postID"	TEXT NOT NULL,
		"commentID" TEXT NOT NULL,
		"userID"	TEXT NOT NULL,
		"total_reactions"		int,
		"likes"     int,
		"Dislikes"  int,
		FOREIGN KEY ("postID")
			REFERENCES "Post" ("postID")
		FOREIGN KEY (commentID)
			REFERENCES "Comment" ("commentID")
		FOREIGN KEY ("userID")
			REFERENCES "User" ("userID")
	);
	`)
}

// open the DB here rather than in a different function so you only have to do it once
func main() {
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		log.Fatal(err)
	}
	Base := &frontend.Drum{
		Base: backend.StartDatabase(db),
	}
	Init(db)
	defer db.Close()
	fs := http.FileServer(http.Dir("./templates"))
	http.Handle("/templates/", http.StripPrefix("/templates/", fs))
	http.HandleFunc("/", Base.StartPage)
	http.HandleFunc("/homepage", Base.Homepage)
	http.HandleFunc("/comments", Base.PostComments)
	http.HandleFunc("/frontendcat", Base.FEndPosts)
	http.HandleFunc("/backendcat", Base.BEndPosts)
	http.HandleFunc("/yourposts", Base.UsersPosts)
	http.HandleFunc("/likedposts", Base.LikedPosts)
	http.HandleFunc("/isLoggedIn", Base.IsLoggedIn)
	http.HandleFunc("/register", Base.Register)
	http.HandleFunc("/login", Base.MyCrewIsLoggingOn)
	http.HandleFunc("/logout", Base.LogOut)
	http.HandleFunc("/post", Base.MakePost)
	http.HandleFunc("/like", Base.Likes)
	http.HandleFunc("/comment", Base.WriteComment)
	http.HandleFunc("/filterby/category", Base.FilterByCategory)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Printf("internet at http://localhost:8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

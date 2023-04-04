package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UserInfo struct {
	username string `json:"username"`
}

func sessionInit(c *gin.Context, userInfo UserInfo) {
	session := sessions.Default(c)
	session.Clear()
	session.Set("count", 0)
	// session.Set()
	session.Set("userinfo", userInfo.username)
	session.Save()
}

var DefaultUserInfo UserInfo = UserInfo{username: ""}

func sessionRead(c *gin.Context) UserInfo {
	session := sessions.Default(c)
	var count int
	v := session.Get("count")
	if v == nil {
		return DefaultUserInfo
	} else {
		count = v.(int)
		count++
	}
	session.Set("count", count)
	/*str := ""
	for i := 0; i < count; i++ {
		str += strconv.Itoa(i)
	}*/
	session.Set("count", count)

	userinfo := session.Get("userinfo")

	if userinfo == nil {
		return DefaultUserInfo
	}
	//userinfo := userinfo.(UserInfo)
	// session.Set("hashkey", str)
	session.Save()
	//c.JSON(200, gin.H{"username": userinfo.username})
	return UserInfo{username: userinfo.(string)}
	// c.JSON(200, gin.H{"count": count})
}

func login(c *gin.Context) {
	fmt.Println("login")
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "login.html", nil)
		/*t, _ := template.ParseFiles("login.html")
		log.Println(t.Execute(c.Writer, nil))*/
	} else if c.Request.Method == "POST" {
		username := c.PostForm("username")
		password := c.PostForm("password")
		rows, err := myDB.Query("SELECT * FROM users WHERE username='" + username + "' and password='" + password + "'")
		checkErr(err)

		sessionInit(c, UserInfo{username: username})
		// sessionCheck(c)

		res := ""
		if rows.Next() {
			//登录成功应该有个状态维护和页面的转变
			res = "login successful"
			c.Redirect(http.StatusMovedPermanently, "/welcome")
		} else {
			res = "login failed"
			fmt.Fprintf(c.Writer, res)
		}
		//fmt.Fprintf(c.Writer, res)

	}
}

func loginToRegister(c *gin.Context) {
	if c.Request.Method == "POST" {
		//c.HTML(http.StatusOK, ".html", nil)
		c.Redirect(http.StatusMovedPermanently, "/welcome/register")
	}
}

func register(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "register.html", nil)
	} else if c.Request.Method == "POST" {
		username := c.PostForm("username")
		password := c.PostForm("password")
		rows, err := myDB.Query("SELECT * FROM users WHERE username='" + username + "'")
		checkErr(err)

		res := ""
		if rows.Next() {
			//登录成功应该有个状态维护和页面的转变
			res = "username exist!"
		} else {
			stmt, err := myDB.Prepare("INSERT INTO users SET id=null,username=?,password=?")
			checkErr(err)

			_, err = stmt.Exec(username, password)
			checkErr(err)
			//得检查是否插入成功

			res = "register successful!"
		}
		/*c.HTML(http.StatusOK, "register", gin.H{
			"registerResult": res,
		})*/
		fmt.Fprintf(c.Writer, res)
	}
}

type Article struct {
	ArticleID         int
	ArticleWriterName string
	ArticleWriteTime  string
	ArticleReNumber   int64
	ArticleTitle      string
	ArticleHher       string
	ArticleContent    string
}

func welcome(c *gin.Context) {
	fmt.Println("welcome")
	if c.Request.Method == "GET" {
		var articles []Article

		rows, _ := myDB.Query("SELECT articles.ArticleID, articles.Title, articles.Content, articles.ReNumber, articles.WriteTime, users.username " +
			"FROM articles, users WHERE articles.WriterID = users.id;")
		{
			for rows.Next() {
				var article Article
				rows.Scan(&article.ArticleID, &article.ArticleTitle, &article.ArticleContent, &article.ArticleReNumber, &article.ArticleWriteTime, &article.ArticleWriterName)
				trow, err := myDB.Query("select count(*) from replies where ArticleID=" +
					strconv.Itoa(article.ArticleID) + ";")
				checkErr(err)
				trow.Next()
				trow.Scan(&article.ArticleReNumber)
				article.ArticleHher = "/welcome/article/" + strconv.Itoa(article.ArticleID) + "#"
				articles = append(articles, article)
			}
		}

		s := ""
		for _, atc := range articles {
			t, _ := template.ParseFiles("templates/welcome/articlecard.html")
			buf := new(bytes.Buffer)
			t.Execute(buf, atc)
			s += buf.String()
		}
		/*res := map[string]interface{}{
			"Articles": articless,
		}*/

		t, _ := template.ParseFiles("templates/welcome/welcome.html")
		t.Execute(c.Writer, template.HTML(s))

		// c.HTML(http.StatusOK, "welcome.html", nil)
		/*t, _ := template.ParseFiles("login.html")
		log.Println(t.Execute(c.Writer, nil))*/
	}
}

type Reply struct {
	WriterName string
	WriteTime  string
	Content    string
}

func viewArticle(c *gin.Context) {
	v := strings.Split(c.Request.URL.String(), "/")
	articleID, _ := strconv.Atoi(v[len(v)-1])

	if c.Request.Method == "GET" {
		var title string
		var content string
		{
			rows, _ := myDB.Query("SELECT articles.Title, articles.Content FROM articles " +
				"WHERE articles.ArticleID = " + strconv.Itoa(articleID) + ";")
			if !rows.Next() {
				return
			} else {
				rows.Scan(&title, &content)
			}
		}

		replies := ""
		{
			rows, err := myDB.Query("SELECT replies.Content, replies.ReplyTime, users.username " +
				"FROM replies, users " +
				"WHERE replies.ArticleID = " +
				strconv.Itoa(articleID) +
				" and replies.WriterID = users.id ORDER BY replies.ReplyTime DESC;")

			// fmt.Println("SELECT replies.Content, replies.ReplyTime, users.username FROM replies, users WHERE replies.ArticleID = " + strconv.Itoa(articleID) + "and replies.WriterID = users.id;")

			if err != nil {
				return
			}
			var replyarr []Reply
			for rows.Next() {
				var reply Reply
				rows.Scan(&reply.Content, &reply.WriteTime, &reply.WriterName)
				replyarr = append(replyarr, reply)

			}

			for _, reply := range replyarr {
				t, _ := template.ParseFiles("templates/article/articlereply.html")
				buf := new(bytes.Buffer)
				t.Execute(buf, reply)
				replies += buf.String()
			}
		}

		res := map[string]interface{}{
			"Title":   title,
			"Content": content,
			"Replies": template.HTML(replies),
			"URL":     c.Request.URL.String(),
		}
		t, _ := template.ParseFiles("templates/article/article.html")
		t.Execute(c.Writer, res)
	} else if c.Request.Method == "POST" {
		// fmt.Println("TEST OK")
		currentTime := time.Now()
		fmt.Println(c.PostForm("reply-content"))
		fmt.Println(currentTime)
		user := sessionRead(c)
		if user == DefaultUserInfo {
			fmt.Fprintf(c.Writer, "请登录！")
		} else {
			splits := strings.Split(c.Request.URL.String(), "/")
			articleID, _ := strconv.Atoi(splits[len(splits)-1])
			content := c.PostForm("reply-content")
			var writerID int
			rows, _ := myDB.Query("SELECT id FROM users WHERE username=\"" +
				user.username + "\";")
			if !rows.Next() {
				return
			}
			rows.Scan(&writerID)

			var ntime string
			ntime = currentTime.Format("2006-01-02 15:04:05")
			sqlq := "INSERT INTO replies SET ArticleID=?, WriterID=?, Content=?, ReplyTime=?"
			stmt, err := myDB.Prepare(sqlq)
			checkErr(err)
			_, err = stmt.Exec(articleID, writerID, content, ntime)
			checkErr(err)

			url := c.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, url)
		}

	}
	// c.Redirect(http.StatusMovedPermanently, /)
}

func viewArticleReply(c *gin.Context) {
	v := strings.Split(c.Request.URL.String(), "/")
	cnt, _ := strconv.Atoi(v[len(v)-1])
	fmt.Fprintf(c.Writer, strconv.Itoa(cnt))
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

var myDB *sql.DB = nil

// var redisStore redis.Store = nil
var cookieStore cookie.Store = nil

func main() {
	mainEngine := gin.Default()

	//暂时开个全局的数据库连接
	tempDB, err := sql.Open("mysql", "root:yswdra@tcp(121.41.40.73:3306)/goweb?charset=utf8")
	myDB = tempDB
	checkErr(err)

	//
	/*redisStore, _ := redis.NewStore(10, "tcp", "121.41.40.73:6379", "", []byte("secret"))
	mainEngine.Use(sessions.Sessions("mysession", redisStore))*/

	cookieStore := cookie.NewStore([]byte("secret11111"))
	mainEngine.Use(sessions.Sessions("mysession", cookieStore))
	cookieStore.Options(sessions.Options{MaxAge: 60 * 60, Path: "/"})

	//载入网页模板
	mainEngine.LoadHTMLGlob("templates/*.html")

	//设置路由
	mainEngine.Any("/welcome", welcome)
	welcomeRouter := mainEngine.Group("/welcome")
	{
		welcomeRouter.Any("/login", login)
		welcomeRouter.Any("/loginToRegister", loginToRegister)
		welcomeRouter.Any("/register", register)
		welcomeRouter.Any("/article/*#", viewArticle)
		/*articleRouter := welcomeRouter.Group("/article/*#")
		{
			articleRouter.Any("/*page", viewArticleReply)
		}*/

	}
	/*mainEngine.GET("/welcome/article/*#", func(c *gin.Context) {
		v := strings.Split(c.Request.URL.String(), "/")
		cnt, _ := strconv.Atoi(v[len(v)-1])
		fmt.Fprintf(c.Writer, strconv.Itoa(cnt))
	})*/
	/*r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})*/

	mainEngine.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

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

		res := ""
		if rows.Next() {
			//登录成功应该有个状态维护和页面的转变
			res = "login successful"
		} else {
			res = "login failed"
		}
		fmt.Fprintf(c.Writer, res)
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

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

var myDB *sql.DB = nil

func main() {
	mainEngine := gin.Default()

	//暂时开个全局的数据库连接
	tempDB, err := sql.Open("mysql", "root:yswdra@tcp(localhost:3306)/goweb?charset=utf8")
	myDB = tempDB
	checkErr(err)

	//载入网页模板
	mainEngine.LoadHTMLGlob("templates/*")

	//设置路由
	welcomeRouter := mainEngine.Group("/welcome")
	{
		welcomeRouter.Any("/login", login)
		welcomeRouter.Any("/loginToRegister", loginToRegister)
		welcomeRouter.Any("/register", register)
	}
	/*r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})*/

	mainEngine.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

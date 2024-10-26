// package main

// import (
//   "github.com/gin-gonic/gin"
// )

// func main () {

//   router := gin.Default()

//   router.GET("/", func(c *gin.Context) {
//     c.String(200, "Hello World")
//   })

//   router.GET("/bye", func(c *gin.Context) {
//     c.String(200, "See you later")
//   })

//   router.Run(":8080")

// }


package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)
func main() {
	r := gin.Default()
	//ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//load html file
	r.LoadHTMLGlob("templates/**/*.tmpl")

	//static path
	r.Static("/assets", "./assets")

	//show home
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/index.tmpl", gin.H{
			"title":    "Home Page",
		})
	})

	//show user template
	r.GET("/users", func(c *gin.Context) {
		c.HTML(http.StatusOK, "users/users.tmpl", gin.H{
			"title": "Users Page",
		})
	})
	//run
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
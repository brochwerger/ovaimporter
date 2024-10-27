package main

import (
	"net/http"
	"path/filepath"
	"github.com/gin-gonic/gin"
)
func main() {
	r := gin.Default()

	r.Static("/assets", "templates/static")

	//load html file
	r.LoadHTMLGlob("templates/**/*.tmpl")

	//show home
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/ocpinfo.tmpl", gin.H{
			"title":    "OCP-V data",
		})
	})

	r.POST("/fileinfo", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/fileinfo.tmpl", gin.H{
			"title":    "OVA data",
		})
	})

	r.POST("/upload", func(c *gin.Context) {

		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}

		filename := filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, "/data/"+filename); err != nil {
			c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
			return
		}

		c.HTML(http.StatusOK, "home/upload.tmpl", gin.H{"filename": file.Filename,})
		// c.String(http.StatusOK, "File %s uploaded successfully\n", file.Filename)

		// err = Untar(filename, ".")
		// if err != nil {
		// 	c.String(http.StatusBadRequest, "uptar file err: %s", err.Error())
		// 	return
		// }

		// c.String(http.StatusOK, "Successfully extracted files from %s OVA\n", file.Filename)


		// c.Redirect(http.StatusOK, "/publish")
	})

	

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
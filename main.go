package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"

)

type  HWRequirements struct {
	diskSize int
	numberOfVCpus int
	memorySize int
	operatingSystem string
}

const DATADIR = "/data/"
// const DATADIR = "./"

var messages chan string 
var wg sync.WaitGroup

func main() {

	messages = make(chan string)
	defer close(messages)

	r := gin.Default()

	r.Static("/assets", "templates/static")

	//load html file
	r.LoadHTMLGlob("templates/**/*.tmpl")

	//show home
	r.GET("/", func(c *gin.Context) {
	//** <Running in-cluster --> no need to access for cluster details>
	// 	c.HTML(http.StatusOK, "home/ocpinfo.tmpl", gin.H{
	// 		"title":    "OCP-V data",
	// 	})
	// })

	// r.POST("/fileinfo", func(c *gin.Context) {
	//** </Running in-cluster>
		c.HTML(http.StatusOK, "home/fileinfo.tmpl", gin.H{
			"title":    "Select OVA source",
		})
	})

	r.POST("/fileinfo", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/fileinfo.tmpl", gin.H{
			"title":    "Select OVA source",
		})
	})

	r.POST("/upload", func(c *gin.Context) {

		file, err := c.FormFile("file")
		if err != nil {
			// cCp.String(http.StatusBadRequest, "get form err: %s", err.Error())
			report("ERROR: " + err.Error())
			return
		}

		filename := filepath.Base(file.Filename)
		fullname := DATADIR + filename

		if err := c.SaveUploadedFile(file, fullname); err != nil {
			// cCp.String(http.StatusBadRequest, "upload file err: %s", err.Error())
			report("ERROR: " + err.Error())
			return
		}
		// report(fmt.Sprintf("Uploaded OVA %s", filename))

		wg.Add(1)
		go func() {

			wg.Wait()

			err = Untar(fullname, DATADIR)
			if err != nil {
				// cCp.String(http.StatusBadRequest, "untar file err: %s", err.Error())
				report("ERROR: " + err.Error())
				return
			}
			report(fmt.Sprintf("Extracted files from %s", filename))

			var hwreqs HWRequirements
			err = ExtractHwRequirements(DATADIR, &hwreqs)
			if err != nil {
				// cCp.String(http.StatusBadRequest, "extract hw info err: %s", err.Error())
				report("ERROR: " + err.Error())
				return
			}
			report(fmt.Sprintf("Extracted HW info:\n\tDisk size: %v\n\tNumber of vCPUS: %v\n\tMemory Size: %v\n\tOperation System: %v",
				hwreqs.diskSize, 
				hwreqs.numberOfVCpus, 
				hwreqs.memorySize, 
				hwreqs.operatingSystem,
			))

			report("Converting vmdk to qcow2 ...")
			var qcow2file string
			qcow2file, err = ConverVMDK(DATADIR)
			if err != nil {
				report("ERROR:" + err.Error())
			}
			report("Creating data volume for " + qcow2file)

			err = CreateResources()
			if err != nil {
				report("ERROR:" + err.Error())
			}
			

		}()

		c.HTML(http.StatusOK, "home/upload.tmpl", gin.H{"filename": filename,})
		wg.Done()

	})

	// Add event-streaming headers
	r.GET("/stream", HeadersMiddleware(), func(c *gin.Context) {
		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			// if msg, ok := <-ClientChan; ok {
			for msg := range messages {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	})
	
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func report(msg string) {
	log.Println(msg)
	messages <- strings.Replace(msg, "\n\t", "<br>", -1)
}

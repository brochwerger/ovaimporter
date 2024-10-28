package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"

	// "mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	// "time"

	"github.com/antchfx/xmlquery"
	"github.com/gin-gonic/gin"

)

type  HWRequirements struct {
	diskSize int
	numberOfVCpus int
	memorySize int
	operatingSystem string
}

const DATADIR = "/data/"

var messages chan string 
var wg sync.WaitGroup

func main() {

	messages = make(chan string)
	defer close(messages)
	wg.Add(1)

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

		go func() {

			wg.Wait()

			err = untar(fullname, DATADIR)
			if err != nil {
				// cCp.String(http.StatusBadRequest, "untar file err: %s", err.Error())
				report("ERROR: " + err.Error())
				return
			}
			report(fmt.Sprintf("Extracted files from %s", filename))

			var hwreqs HWRequirements
			err = extractHwRequirements(DATADIR, &hwreqs)
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
	messages <- msg
}

func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractHwRequirements(dirname string, hwreqs *HWRequirements) error {

	dir, err := os.Open(dirname)
	if err != nil {
		return(err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
	   return err
	}
 
	var filename string
	found := false
	for _, file := range files {
		filename = file.Name()
	   	if filepath.Ext(filename) == ".ovf" {
			found = true
			break
	   	}
	}
	if !found {
		return fmt.Errorf("ERROR: OVF file not found")
	}
 
	// Open the XML file
	file, err := os.Open(DATADIR + filename)
	if err != nil {
		return(err)
	}
	defer file.Close()

	// Parse the XML file
	doc, err := xmlquery.Parse(file)
	if err != nil {
		return(err)
	}

	// Define the XPath expression to find the key value
	expr := "//Disk[@ovf:capacity]"
	node := xmlquery.FindOne(doc, expr)
	if node != nil {
		hwreqs.diskSize, _ = strconv.Atoi(node.SelectAttr("ovf:capacity"))
		// fmt.Println("Disk Capacity:", node.SelectAttr("ovf:capacity"))
		// fmt.Println("Allocation Unit:", node.SelectAttr("ovf:capacityAllocationUnits"))
	} else {
		err := fmt.Errorf("Key not found: %s", expr)
		return err
	}

	expr = "//OperatingSystemSection/Description"
	node = xmlquery.FindOne(doc, expr)
	if node != nil {
		hwreqs.operatingSystem = node.FirstChild.Data
		// fmt.Println("Operating System:", node.FirstChild.Data) //.SelectAttr("ovf:capacity"))
	} else {
		err := fmt.Errorf("Key not found: %s", expr)
		return err
	}

	expr = "//Item/rasd:Description[text()=\"Number of Virtual CPUs\"]/../rasd:VirtualQuantity"
	node = xmlquery.FindOne(doc, expr)
	if node != nil {
		hwreqs.numberOfVCpus, _ = strconv.Atoi(node.FirstChild.Data)
		// fmt.Println("Number of vCPUs:", node.FirstChild.Data)
	} else {
		err := fmt.Errorf("Key not found: %s", expr)
		return err
	}

	expr = "//Item/rasd:Description[text()=\"Memory Size\"]/../rasd:VirtualQuantity"
	node = xmlquery.FindOne(doc, expr)
	if node != nil {
		hwreqs.memorySize, _ = strconv.Atoi(node.FirstChild.Data)
		// fmt.Println("Memory Size:", node.FirstChild.Data)
	} else {
		err := fmt.Errorf("Key not found: %s", expr)
		return err
	}

	return nil

}


package main

import (
	"os"
	"io"
	"fmt"
	"archive/tar"
	"path/filepath"
	"strconv"
	// "context"
	"os/exec"
	// "fmt"
	// "time"
	"strings"

	"github.com/antchfx/xmlquery"

	// // "k8s.io/apimachinery/pkg/api/errors"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/rest"
)

func Untar(tarball, target string) error {
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

func ExtractHwRequirements(dirname string, hwreqs *HWRequirements) error {

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
		err := fmt.Errorf("key not found: %s", expr)
		return err
	}

	expr = "//OperatingSystemSection/Description"
	node = xmlquery.FindOne(doc, expr)
	if node != nil {
		hwreqs.operatingSystem = node.FirstChild.Data
		// fmt.Println("Operating System:", node.FirstChild.Data) //.SelectAttr("ovf:capacity"))
	} else {
		err := fmt.Errorf("key not found: %s", expr)
		return err
	}

	expr = "//Item/rasd:Description[text()=\"Number of Virtual CPUs\"]/../rasd:VirtualQuantity"
	node = xmlquery.FindOne(doc, expr)
	if node != nil {
		hwreqs.numberOfVCpus, _ = strconv.Atoi(node.FirstChild.Data)
		// fmt.Println("Number of vCPUs:", node.FirstChild.Data)
	} else {
		err := fmt.Errorf("key not found: %s", expr)
		return err
	}

	expr = "//Item/rasd:Description[text()=\"Memory Size\"]/../rasd:VirtualQuantity"
	node = xmlquery.FindOne(doc, expr)
	if node != nil {
		hwreqs.memorySize, _ = strconv.Atoi(node.FirstChild.Data)
		// fmt.Println("Memory Size:", node.FirstChild.Data)
	} else {
		err := fmt.Errorf("key not found: %s", expr)
		return err
	}

	return nil

}

func ConverVMDK(dirname string) (string, error) {

	dir, err := os.Open(dirname)
	if err != nil {
		return "", err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return "", err
	}
 
	var filename, vmdkname, qcow2name string
	found := 0
	for _, file := range files {
		filename = file.Name()
	   	if filepath.Ext(filename) == ".ova" {
			qcow2name = strings.Replace(filename, filepath.Ext(filename), ".qcow2", 1)
			found++
	   	}
	   	if filepath.Ext(filename) == ".vmdk" {
			vmdkname = filename
			found++
		}
		if found == 2 {
			break
		}
	}

	if found != 2 {
		return "", fmt.Errorf("ERROR: VMDK file not found")
	}

	cmd := exec.Command("qemu-img", "convert", "-cpf", "vmdk", "-O", "qcow2", 
		dirname + "/" + vmdkname, 
		dirname + "/" + qcow2name)

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	// running := make(chan bool)
	// go func (running chan bool) {
		
	// 	for {
	// 		select {
	// 		case <-running :
	// 			fmt.Println("")
	// 			return
	// 		default:
	// 			time.Sleep(time.Second)
	// 			fmt.Print(".")
	// 		}
	// 	}	  

	// } (running)
	// err = cmd.Start()
	// if err == nil {
	// 	cmd.Wait()
	// }
	// running <- false
	return qcow2name, nil

}

func CreateResources() error {
	return nil
	// // creates the in-cluster config
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	panic(err.Error())
	// }
	// // creates the clientset
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// for {
	// 	// get pods in all the namespaces by omitting namespace
	// 	// Or specify namespace to get pods in particular namespace
	// 	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	// }
	// return err
}
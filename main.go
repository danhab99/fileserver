package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	cwdPath, err := os.Getwd()
	// cwd += "/"
	fmt.Printf("cwd: %v\n", cwdPath)
	if err != nil {
		panic(err)
	}

	cwdPtr := flag.String("path", cwdPath, "the path to display files from")
	port := flag.Int("port", 8080, "the port to listen to")
	flag.Parse()

	router := gin.Default()

	cwd := *cwdPtr

	router.GET("/*dirPath", func(c *gin.Context) {
		dirPath := c.Param("dirPath")
		fmt.Printf("Getting dirPath: %v\n", dirPath)
		info, err := os.Stat(cwd + dirPath)
		if err != nil {
			panic(err)
		}

		write := func(d string) {
			c.Writer.Write([]byte(d))
		}

		c.Status(200)
		if info.IsDir() {
			c.Header("Content-Type", "text/html")
			write(`
<html>
<body>
<form enctype="multipart/form-data" method="POST" action="`)
			write(dirPath)
			write(`">
<input type="file" multiple name="files" id="files" />
<input type="submit" />
</form>
<ol>
<li><a href="..">../</a></li>
`)
			items, err := os.ReadDir(cwd + dirPath)
			if err != nil {
				panic(err)
			}

			for _, item := range items {
				fmt.Printf("item: %v\n", item)
				write("<li><a href=\"")
				write(dirPath)
				if len(dirPath) > 1 {
					write("/")
				}
				write(item.Name())
				write("\">")
				if item.IsDir() {
					write("/")
				}
				write(item.Name())
				write("</a></li>\n")
			}

			write(`
</ol>
</body>
</html>
`)
		} else {
			fmt.Println("Sending file", cwd, dirPath)
			c.File(cwd + dirPath)
		}
	})

	router.POST("/*dirPath", func(c *gin.Context) {
		dirPath := c.Param("dirPath")
		info, err := os.Stat(cwd + dirPath)
		if err != nil {
			panic(err)
		}

		if !info.IsDir() {
			panic("Cannot upload to file")
		}

		form, err := c.MultipartForm()
		if err != nil {
			panic(err)
		}

		fmt.Printf("form: %v\n", form)

		for _, files := range form.File {
			for i, file := range files {
				name := file.Filename
				if i > 0 {
					name += fmt.Sprintf(".%d", i)
				}
				fmt.Println("Receiving file", name)

				if len(strings.Split(name, "/")) > 1 {
					base := cwd + "/" + path.Base(name)
					_, err := os.Stat(base)

					if os.IsNotExist(err) {
						err = os.MkdirAll(base, 0666)
						if err != nil {
							panic(err)
						}
					}
				}

				f, err := os.Create(cwd + "/" + name)
				if err != nil {
					panic(err)
				}

				r, err := file.Open()
				if err != nil {
					panic(err)
				}

				_, err = io.Copy(f, r)
				if err != nil {
					panic(err)
				}

			}
		}

		c.Redirect(302, dirPath)
	})

	router.Run(fmt.Sprintf("0.0.0.0:%d", *port))
}

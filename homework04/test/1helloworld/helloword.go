package main
import (
    "net/http"
	"github.com/gin-gonic/gin"
)
func main() {
    r := gin.Default()
    r.GET("/hello", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Hello World",
        })
    })
	r.GET("/ping", func(c *gin.Context) {
        c.String(http.StatusOK, "pong")
    })

	r.LoadHTMLGlob("templates/*")
	r.GET("/helloHtml", func(c *gin.Context) {
        c.HTML(http.StatusOK, "hello.html", gin.H{
            "title": "Hello Html",
        })
    })
    r.Run()
}
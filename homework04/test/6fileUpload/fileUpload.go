package main
import (
    // "fmt"
    "net/http"
	"path/filepath"
    "os"
	"github.com/gin-gonic/gin"
)
func main() {
    r := gin.Default()

	//创建上传目录
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil{
		panic(err)
	}

	//单文件上传
	r.POST("/upload", uploadFile)

	//多个文件上传
	r.POST("/uploads", uploadMultipleFiles)

	//下载文件
	r.GET("/download/:filename", downloadFile)
	
	r.Static("/static", "./uploads/text.txt")
	r.StaticFS("/files", http.Dir("./uploads"))

	r.Run(":8080")
}
func uploadFile(c *gin.Context) {
	// 获取上传的文件
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 保存文件
	filename := filepath.Base(file.Filename)
	dst := filepath.Join("uploads", filename)
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully" ,"filename": filename,"size": file.Size})
}
//上传多个文件
func uploadMultipleFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	files := form.File["files"] // "files" 是表单字段名
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}
	var uploadedFiles []string
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		dst := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		uploadedFiles = append(uploadedFiles, filename)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Files uploaded successfully", "files": uploadedFiles,"count": len(uploadedFiles)})
}

//下载文件、
func downloadFile(c *gin.Context) {
	filename := c.Param("filename")
	filePath := filepath.Join("uploads", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	//设置响应头
	c.Header("Content-Description","File Transfer")
	c.Header("Content-Transfer-Encoding","binary")
	c.Header("Content-Disposition","attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	//读取文件并返回
	c.File(filePath)
}
package main
import (
	// "github.com/joho/godotenv"
	"fmt"
	// "os"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/spf13/viper"
	"log"
)
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT JWTConfig `mapstructure:"jwt"`
}
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Mode string `mapstructure:"mode"`
}
type DatabaseConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	User string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName string `mapstructure:"dbname"`
}
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire string `mapstructure:"expires"`
}
func init(){
	//设置配置文件名称（不含拓展名）
	viper.SetConfigName("config")
	//设置配置文件类型
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.app")

	//读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	//设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.mode", "debug")

	//加载配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Error reading config file: %v", err)
		log.Println("Using default values and environment variables")
	}else{
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

}

//加载并分析配置文件，将viper中的配置数据映射到Config结构体
//返回解析后的配置对象指针和可能的错误
//配置数据来源包括：配置文件、环境变量、默认值（在init函数中已配置）
func loadConfig() (*Config ,error){
	var config Config

	//使用 viper.Unmarshal() 将viper中的配置数据映射到Config结构体中
	//如果解析失败，则返回错误
	if err := viper.Unmarshal(&config); err != nil {
		return nil,err
	}

	//返回解析成功的配置对象指针
	return &config,nil
}
func main() { 
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	//设置gin模式
	gin.SetMode(config.Server.Mode)
	r := gin.Default()

	r.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Go Web App",
			"server": config.Server,
			"database": gin.H{
				"host": config.Database.Host,
				"port": config.Database.Port,
				"user": config.Database.User,
				"dbname": config.Database.DBName,
				//不返回密码
			},
			"jwt": gin.H{
				"expires": config.JWT.Expire,
				//不返回密钥
			},
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Server is healthy",
			"status": "ok",
			"config_file": viper.ConfigFileUsed(),
		})
	})

	addr := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
	log.Printf("Server listening on %s", addr)
	r.Run(addr)
}
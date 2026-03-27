package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"

	"protobuf/pb"
)

func main() {
	r := gin.Default()

	api := r.Group("/api/proto")
	{
		//获取单个用户（protobuf）
		api.GET("/user/:id", getUserProtobuf)

		//获取用户列表（protobuf）
		api.GET("/users", getUsersListProtobuf)
	}

	r.Run(":8080")
}
func getUserProtobuf(c *gin.Context) {
	id := c.Param("id")

	//模拟从数据库中获取用户数据
	user := &pb.User{
		Id:       1,
		Username: "alice",
		Email:    "alice@example.com",
		Age:      20,
		Active:   true,
		Tags:     []string{"admin", "developer"},
		Metadata: map[string]string{
			"department": "engineering",
			"location":   "Beijing",
		},
	}

	//如果提供了id参数，可以设置不同的id
	if id == "2" {
		user.Id = 2
		user.Username = "bob"
		user.Email = "bob@example.com"
		user.Age = 25
		user.Active = true
		user.Tags = []string{"admin", "developer"}
		user.Metadata = map[string]string{
			"department": "engineering",
			"location":   "Beijing",
		}
	}

	//序列化 protobuf
	data, err := proto.Marshal(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to marshal protobuf: %v", err),
		})
		return
	}

	//设置响应头
	c.Header("Content-Type", "application/x-protobuf")
	c.Data(http.StatusOK, "application/protobuf", data)

}
func getUsersListProtobuf(c *gin.Context) {
	//模拟从数据库中获取用户列表数据
	users := []*pb.User{
		{
			Id:       1,
			Username: "alice",
			Email:    "alice@example.com",
			Age:      20,
			Active:   true,
			Tags:     []string{"admin", "developer"},
			Metadata: map[string]string{
				"department": "engineering",
			},
		},
		{
			Id:       2,
			Username: "bob",
			Email:    "bob@example.com",
			Age:      25,
			Active:   true,
			Tags:     []string{"admin", "developer"},
			Metadata: map[string]string{
				"department": "engineering",
			},
		},
		{
			Id:       3,
			Username: "jon",
			Email:    "jon@example.com",
			Age:      22,
			Active:   true,
			Tags:     []string{"admin", "developer"},
			Metadata: map[string]string{
				"department": "engineering",
			},
		},
	}
	userList := &pb.UserList{
		Users: users,
		Total: int32(len(users)),
	}
	data, err := proto.Marshal(userList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to marshal protobuf: %v", err),
		})
		return
	}
	c.Header("Content-Type", "application/x-protobuf")
	c.Data(http.StatusOK, "application/protobuf", data)
}

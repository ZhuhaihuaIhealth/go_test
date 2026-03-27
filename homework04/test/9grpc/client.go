package main

import (
	"context"
	"fmt"
	"grpc/pb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserServiceClient 客户端封装
type UserServiceClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

// NewUserServiceClient 创建新的客户端
func NewUserServiceClient(addr string) (*UserServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	return &UserServiceClient{
		conn:   conn,
		client: pb.NewUserServiceClient(conn),
	}, nil
}

// close 关闭连接
func (c *UserServiceClient) Close() error {
	return c.conn.Close()
}

// GetUser 获取单个用户
func (c *UserServiceClient) GetUser(id int64) (*pb.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.GetUserRequest{
		Id: id,
	}
	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return resp, nil
}

// CreateUser 创建用户
func (c *UserServiceClient) CreateUser(username, email string, age int32, tags []string) (*pb.CreateUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.CreateUserRequest{
		Username: username,
		Email:    email,
		Age:      age,
		Tags:     tags,
		Metadata: map[string]string{"source": "grpc-client"},
	}
	resp, err := c.client.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	return resp, nil
}

// 获取用户列表
func (c *UserServiceClient) ListUsers(page, pageSize int32) (*pb.ListUsersResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
	}
	resp, err := c.client.ListUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}
	return resp, nil
}

// UpdateUser更新用户
func (c *UserServiceClient) UpdateUser(id int64, username, email string, age int32, active bool) (*pb.UpdateUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.UpdateUserRequest{
		Id:       id,
		Username: username,
		Email:    email,
		Age:      age,
		Active:   active,
	}
	resp, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}
	return resp, nil
}

// DeleteUser删除用户
func (c *UserServiceClient) DeleteUser(id int64) (*pb.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.DeleteUserRequest{
		Id: id,
	}
	resp, err := c.client.DeleteUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %v", err)
	}
	return resp, nil
}

// StreamUsers流式获取用户
func (c *UserServiceClient) StreamUsers(limit int32, intervalMs int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req := &pb.StreamUsersRequest{
		Limit:      limit,
		IntervalMs: intervalMs,
	}
	stream, err := c.client.StreamUsers(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stream users: %v", err)
	}
	fmt.Println("streaming users...")
	count := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive users: %v", err)
		}
		count++
		fmt.Printf("[%d] User: %s\n ( ID: %d, Email: %s)", count, resp.Username, resp.Id, resp.Email)
	}
	fmt.Println("Total Users: %d ", count)
	return nil
}

// BatchCreateUsers批量创建用户（客户端流）
func (c *UserServiceClient) BatchCreateUsers(user []*pb.CreateUserRequest) (*pb.BatchCreateUsersResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := c.client.BatchCreateUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to batch create users: %v", err)
	}
	for _, req := range user {
		if err := stream.Send(req); err != nil {
			return nil, fmt.Errorf("failed to send user: %v", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("failed to batch create users: %v", err)
	}
	return resp, nil
}

// chatUsers 双向流（聊天式交互）
func (c *UserServiceClient) ChatUsers(messages []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	stream, err := c.client.ChatUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to chat users: %v", err)
	}
	done := make(chan error, 1)

	//启动接收协程
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- nil
				return
			}
			if err != nil {
				done <- fmt.Errorf("failed to receive message: %v", err)
				return
			}
			fmt.Printf("Server:%s at %s\n", resp.Message, time.Unix(resp.Timestamp, 0).Format(time.RFC3339))
		}

	}()

	//发送消息
	for i, msg := range messages {
		req := &pb.ChatMessage{
			UserId:    "client",
			Message:   msg,
			Timestamp: time.Now().Unix(),
		}

		if err := stream.Send(req); err != nil {
			return fmt.Errorf("failed to send message: %v", err)
		}
		fmt.Printf("Client: %d ---  %s\n", i+1, msg)
	}
	//关闭发送
	if err := stream.CloseSend(); err != nil {
		return fmt.Errorf("failed to close send: %v", err)
	}
	//等待接收完成
	return <-done
}

// runClientDemo	运行客户端示例
func runClientDemo(addr string) {
	client, err := NewUserServiceClient(addr)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()
	fmt.Println("=== gRPC Client Demo ===\n")

	fmt.Println("1.GetUser (ID: 1)\n")
	user, err := client.GetUser(1)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("User: %s\n ( ID: %d, Email: %s)", user.Username, user.Id, user.Email)
	}
	fmt.Printf("User: %s\n ( ID: %d, Email: %s, Age:%d)", user.Username, user.Id, user.Email, user.Age)

	//创建用户
	fmt.Println("\n2.CreateUser\n")
	resp, err := client.CreateUser("test", "test@test.com", 20, []string{"developer"})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("User created: %s\n ( ID: %d, Email: %s)", resp.User.Username, resp.User.Id, resp.User.Email)
	}

	//3.获取用户列表
	fmt.Println("\n3. ListUsers (page: 1, pageSize: 5)")
	listResp, err := client.ListUsers(1, 5)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Total Users: %d\n", listResp.Total)
		for _, user := range listResp.Users {
			fmt.Printf("User: %s\n ( ID: %d, Email: %s)", user.Username, user.Id, user.Email)
		}
	}

	//4.更新用户
	fmt.Println("\n4. UpdateUser (ID: 1)")
	updateResp, err := client.UpdateUser(1, "test2", "test2@test.com", 26, true)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("%s \n", updateResp.Message)
	}

	//5.流式获取用户
	fmt.Println("\n5. StreamUsers (limit: 5, interval: 100ms)")
	err = client.StreamUsers(5, 100)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	//6.批量创建用户
	fmt.Println("\n6. BatchCreateUsers")
	users := []*pb.CreateUserRequest{
		{
			Username: "test3",
			Email:    "test3@test.com",
			Age:      20,
		},
		{
			Username: "test4",
			Email:    "test4@test.com",
			Age:      21,
		},
		{
			Username: "test5",
			Email:    "test5@test.com",
			Age:      22,
		},
	}
	batchResp, err := client.BatchCreateUsers(users)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Users Created: %d\n", resp.Message)
		fmt.Printf("Success: %d\n Failed: %d\n", batchResp.SuccessCount, batchResp.FailCount)
	}

	//7.双向流
	fmt.Println("\n7. ChatUsers")
	messages := []string{
		"Hello, server!",
		"How are you?",
		"I'm fine, thanks!",
		"Bye!",
	}
	err = client.ChatUsers(messages)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Println("Chat completed")
}

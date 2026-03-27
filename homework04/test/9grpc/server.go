package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"grpc/pb"
)

// UserServiceServer 实现UserService接口
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	users  map[int64]*pb.User
	mu     sync.RWMutex
	nextID int64
}

// NewUserServiceServer 创建新的用户服务实例
func NewUserServiceServer() *UserServiceServer {
	server := &UserServiceServer{
		users:  make(map[int64]*pb.User),
		nextID: 1,
	}
	server.initSampleData() //初始化一些示例数据
	return server
}

// initSampleData 初始化一些示例用户数据
func (s *UserServiceServer) initSampleData() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[1] = &pb.User{
		Id:       1,
		Username: "张三",
		Email:    "zhangsan@example.com",
		Age:      18,
		Active:   true,
		Tags:     []string{"admin", "developer"},
		Metadata: map[string]string{
			"department": "engineering",
			"location":   "Beijing",
		},
	}
	s.users[2] = &pb.User{
		Id:       2,
		Username: "李四",
		Email:    "lisi@example.com",
		Age:      20,
		Active:   true,
		Tags:     []string{"user", "tester"},
		Metadata: map[string]string{
			"department": "qa",
			"location":   "Beijing",
		},
	}
	s.nextID = 3
}

// GetUser 获取单个用户
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, exists := s.users[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "id %d 用户不存在", req.Id)
	}
	return user, nil
}

// CreateUser 创建用户
func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.Username == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "username and email are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	//检查用户名是否已存在
	for _, user := range s.users {
		if user.Username == req.Username {
			return nil, status.Errorf(codes.AlreadyExists, "name %s 用户已存在", req.Username)
		}
	}

	//创建新用户
	user := &pb.User{
		Id:       s.nextID,
		Username: req.Username,
		Email:    req.Email,
		Age:      req.Age,
		Active:   true,
		Tags:     req.Tags,
		Metadata: req.Metadata,
	}
	if user.Metadata == nil {
		user.Metadata = make(map[string]string)
	}
	user.Metadata["created_at"] = time.Now().Format(time.RFC3339)

	s.users[s.nextID] = user
	s.nextID++
	return &pb.CreateUserResponse{
		User:    user,
		Success: true,
		Message: fmt.Sprintf("用户 %s 创建成功", req.Username),
	}, nil
}

// ListUsers 获取用户列表
func (s *UserServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	//分页参数
	page := req.Page
	if page <= 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	//获取所有用户
	allUsers := make([]*pb.User, 0, len(s.users))
	for _, user := range s.users {
		allUsers = append(allUsers, user)
	}

	//计算分页
	total := int32(len(allUsers))
	usersLen := len(allUsers)
	start := int((page - 1) * pageSize)
	end := start + int(pageSize)

	if start > usersLen {
		return &pb.ListUsersResponse{
			Users:    []*pb.User{},
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}, nil

	}
	if end > usersLen {
		end = usersLen
	}
	return &pb.ListUsersResponse{
		Users:    allUsers[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateUser 更新用户
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, exists := s.users[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "id %d 用户不存在", req.Id)
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Age > 0 {
		user.Age = req.Age
	}
	if len(req.Tags) > 0 {
		user.Tags = req.Tags
	}
	if len(req.Metadata) > 0 {
		if user.Metadata == nil {
			user.Metadata = make(map[string]string)
		}
		for k, v := range req.Metadata {
			user.Metadata[k] = v
		}
	}
	return &pb.UpdateUserResponse{
		User:    user,
		Success: true,
		Message: fmt.Sprintf("用户 %s 更新成功", user.Username),
	}, nil
}

// DeleteUser 删除用户
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.users[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "id %d 用户不存在", req.Id)
	}
	delete(s.users, req.Id)
	return &pb.DeleteUserResponse{
		Success: true,
		Message: fmt.Sprintf("用户 %d 删除成功", req.Id),
	}, nil
}

// 流式获取用户（服务端流）
func (s *UserServiceServer) StreamUsers(req *pb.StreamUsersRequest, stream pb.UserService_StreamUsersServer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*pb.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = len(users)
	}
	if limit > len(users) {
		limit = len(users)
	}
	interval := time.Duration(req.IntervalMs * int32(time.Millisecond))
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	for i := 0; i < limit; i++ {
		if err := stream.Send(users[i]); err != nil {
			return err
		}
		if i < limit-1 {
			time.Sleep(interval)
		}

	}
	return nil
}

// BatchCreateUsers 批量创建用户 (客户端流)
func (s *UserServiceServer) BatchCreateUsers(stream pb.UserService_BatchCreateUsersServer) error {
	var createdUsers []*pb.User
	successCount := 0
	failCount := 0
	for {
		req, err := stream.Recv()
		if err != nil {
			//流结束
			return err
		}
		resp, err := s.CreateUser(stream.Context(), req)
		if err != nil {
			failCount++
			log.Printf("创建用户 %s 失败: %v", req.Username, err)
			continue
		}
		createdUsers = append(createdUsers, resp.User)
		successCount++
	}
	return stream.SendAndClose(&pb.BatchCreateUsersResponse{
		Users:        createdUsers,
		SuccessCount: int32(successCount),
		FailCount:    int32(failCount),
		Message:      fmt.Sprintf("成功创建 %d 个用户，失败 %d 个", successCount, failCount),
	})
}

// ChatUsers 双向流（聊天式交互）
func (s *UserServiceServer) ChatUsers(stream pb.UserService_ChatUsersServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Printf("收到来自 %s 的消息: %s", msg.UserId, msg.Message)
		response := &pb.ChatMessage{
			UserId:    "server",
			Message:   fmt.Sprintf("收到来自 %s 的消息: %s", msg.UserId, msg.Message),
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

// startServer 启动grpc服务
func startServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	//创建grpc服务器
	grpcServer := grpc.NewServer()

	//注册服务
	userService := NewUserServiceServer()
	pb.RegisterUserServiceServer(grpcServer, userService)

	log.Printf("grpc server started at %s", port)
	log.Println("Available methods:")
	log.Println("  - GetUser")
	log.Println("  - CreateUser")
	log.Println("  - ListUsers")
	log.Println("  - UpdateUser")
	log.Println("  - DeleteUser")
	log.Println("  - StreamUsers (server streaming)")
	log.Println("  - BatchCreateUsers (client streaming)")
	log.Println("  - ChatUsers (bidirectional streaming)")

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

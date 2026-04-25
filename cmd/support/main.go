package main

import (
	"context"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "guidely-app/pkg/pb/support/v1"
)

type supportServer struct {
	pb.UnimplementedSupportServiceServer
	// TODO: добавить репозиторий и сервис
}

func (s *supportServer) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.Ticket, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	//    TODO: заменить на реальную бизнес-логику + сохранение в БД
	log.Printf("CreateTicket: user_id=%d, category=%s, subject=%s", userID, req.Category, req.Subject)

	return &pb.Ticket{
		Id:        1,
		UserId:    userID,
		Category:  req.Category,
		Status:    "open",
		Subject:   req.Subject,
		CreatedAt: "2025-04-13T10:00:00Z",
	}, nil
}

func (s *supportServer) ListMyTickets(ctx context.Context, req *pb.ListTicketsRequest) (*pb.ListTicketsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	log.Printf("ListMyTickets: user_id=%d", userID)

	return &pb.ListTicketsResponse{
		Tickets: []*pb.Ticket{},
	}, nil
}

func (s *supportServer) GetTicket(ctx context.Context, req *pb.GetTicketRequest) (*pb.Ticket, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	log.Printf("GetTicket: user_id=%d, ticket_id=%d", userID, req.TicketId)

	return &pb.Ticket{
		Id:        req.TicketId,
		UserId:    userID,
		Category:  "bug",
		Status:    "open",
		Subject:   "Пример обращения",
		CreatedAt: "2025-04-13T10:00:00Z",
		Messages: []*pb.Message{
			{
				Id:         1,
				TicketId:   req.TicketId,
				AuthorType: "user",
				Text:       "Текст сообщения",
				CreatedAt:  "2025-04-13T10:05:00Z",
			},
		},
	}, nil
}

func (s *supportServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	log.Printf("SendMessage: user_id=%d, ticket_id=%d, text=%s", userID, req.TicketId, req.Text)

	return &pb.Message{
		Id:         1,
		TicketId:   req.TicketId,
		AuthorType: "user",
		Text:       req.Text,
		CreatedAt:  "2025-04-13T10:10:00Z",
	}, nil
}

func (s *supportServer) GetStats(ctx context.Context, req *pb.Empty) (*pb.StatsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	log.Printf("GetStats: запрошено админом user_id=%d", userID)
	return &pb.StatsResponse{
		OpenCount:        5,
		InProgressCount:  3,
		ClosedCount:      10,
		ByCategory:       map[string]int64{"bug": 8, "proposal": 6, "complaint": 4},
		TotalTickets:     18,
		TotalMessages:    42,
	}, nil
}

func (s *supportServer) ListOpenTickets(ctx context.Context, req *pb.ListTicketsRequest) (*pb.ListTicketsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	log.Printf("ListOpenTickets: запрошено админом user_id=%d", userID)

	return &pb.ListTicketsResponse{Tickets: []*pb.Ticket{}}, nil
}

func (s *supportServer) ReplyAsAdmin(ctx context.Context, req *pb.ReplyAsAdminRequest) (*pb.Message, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	log.Printf("ReplyAsAdmin: admin_id=%d, ticket_id=%d, text=%s", userID, req.TicketId, req.Text)

	return &pb.Message{
		Id:         1,
		TicketId:   req.TicketId,
		AuthorType: "admin",
		Text:       req.Text,
		CreatedAt:  "2025-04-13T10:15:00Z",
	}, nil
}

func (s *supportServer) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.Ticket, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	log.Printf("UpdateTicketStatus: admin_id=%d, ticket_id=%d, new_status=%s", userID, req.TicketId, req.Status)

	return &pb.Ticket{Id: req.TicketId, Status: req.Status}, nil
}

func getUserIDFromContext(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("x-user-id")
	if len(userIDs) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing user id")
	}

	userID, err := strconv.ParseInt(userIDs[0], 10, 64)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid user id format")
	}

	return userID, nil
}

func userIDInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func adminInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	roles := md.Get("x-role")
	if len(roles) == 0 || roles[0] != "admin" {
		return nil, status.Error(codes.PermissionDenied, "admin role required")
	}

	return handler(ctx, req)
}

func main() {
	myServer := &supportServer{}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(userIDInterceptor),

	)

	pb.RegisterSupportServiceServer(grpcServer, myServer)

	listener, err := net.Listen("tcp", ":8084")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server for support service is running on port 8084...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
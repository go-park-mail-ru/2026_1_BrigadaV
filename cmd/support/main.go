package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "guidely-app/pkg/pb/support/v1"
)


type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(dbURL string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

func (r *PostgresRepo) CreateTicket(userID int64, category, subject string) (*pb.Ticket, error) {
	query := `INSERT INTO tickets (user_id, category, subject, created_at) 
			  VALUES ($1, $2, $3, NOW()) 
			  RETURNING id, created_at`

	var id int64
	var createdAtStr string
	err := r.db.QueryRow(query, userID, category, subject).Scan(&id, &createdAtStr)
	if err != nil {
		return nil, err
	}

	return &pb.Ticket{
		Id:        id,
		UserId:    userID,
		Category:  category,
		Status:    "open",
		Subject:   subject,
		CreatedAt: createdAtStr,
	}, nil
}

func (r *PostgresRepo) ListTicketsByUser(userID int64) ([]*pb.Ticket, error) {
	query := `SELECT id, user_id, category, status, subject, created_at 
			  FROM tickets WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*pb.Ticket
	for rows.Next() {
		var t pb.Ticket
		var createdAtStr string
		err := rows.Scan(&t.Id, &t.UserId, &t.Category, &t.Status, &t.Subject, &createdAtStr)
		if err != nil {
			continue
		}
		t.CreatedAt = createdAtStr
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *PostgresRepo) GetTicketByID(ticketID int64) (*pb.Ticket, []*pb.Message, error) {
	ticketQuery := `SELECT id, user_id, category, status, subject, created_at 
					FROM tickets WHERE id = $1`
	var ticket pb.Ticket
	var createdAtStr string
	err := r.db.QueryRow(ticketQuery, ticketID).Scan(
		&ticket.Id, &ticket.UserId, &ticket.Category,
		&ticket.Status, &ticket.Subject, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, err
		}
		return nil, nil, err
	}
	ticket.CreatedAt = createdAtStr

	msgQuery := `SELECT id, ticket_id, author_type, text, created_at 
				FROM messages WHERE ticket_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(msgQuery, ticketID)
	if err != nil {
		return &ticket, nil, nil
	}
	defer rows.Close()

	var messages []*pb.Message
	for rows.Next() {
		var msg pb.Message
		var msgCreatedAt string
		err := rows.Scan(&msg.Id, &msg.TicketId, &msg.AuthorType, &msg.Text, &msgCreatedAt)
		if err != nil {
			continue
		}
		msg.CreatedAt = msgCreatedAt
		messages = append(messages, &msg)
	}

	return &ticket, messages, nil
}

func (r *PostgresRepo) AddMessage(ticketID int64, authorType, text string) (*pb.Message, error) {
	query := `INSERT INTO messages (ticket_id, author_type, text, created_at) 
			  VALUES ($1, $2, $3, NOW()) 
			  RETURNING id, created_at`

	var id int64
	var createdAtStr string
	err := r.db.QueryRow(query, ticketID, authorType, text).Scan(&id, &createdAtStr)
	if err != nil {
		return nil, err
	}

	return &pb.Message{
		Id:         id,
		TicketId:   ticketID,
		AuthorType: authorType,
		Text:       text,
		CreatedAt:  createdAtStr,
	}, nil
}

func (r *PostgresRepo) UpdateTicketStatus(ticketID int64, status string) error {
	query := `UPDATE tickets SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, ticketID)
	return err
}

func (r *PostgresRepo) GetStats() (open, inProgress, closed int64, byCategory map[string]int64, err error) {
	// Статусы
	statusQuery := `SELECT status, COUNT(*) FROM tickets GROUP BY status`
	rows, err := r.db.Query(statusQuery)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		rows.Scan(&status, &count)
		switch status {
		case "open":
			open = count
		case "in_progress":
			inProgress = count
		case "closed":
			closed = count
		}
	}

	// Категории
	catQuery := `SELECT category, COUNT(*) FROM tickets GROUP BY category`
	catRows, err := r.db.Query(catQuery)
	if err != nil {
		return open, inProgress, closed, nil, err
	}
	defer catRows.Close()

	byCategory = make(map[string]int64)
	for catRows.Next() {
		var category string
		var count int64
		catRows.Scan(&category, &count)
		byCategory[category] = count
	}

	return open, inProgress, closed, byCategory, nil
}

func (r *PostgresRepo) ListOpenTickets() ([]*pb.Ticket, error) {
	query := `SELECT id, user_id, category, status, subject, created_at 
			  FROM tickets WHERE status IN ('open', 'in_progress') 
			  ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*pb.Ticket
	for rows.Next() {
		var t pb.Ticket
		var createdAtStr string
		err := rows.Scan(&t.Id, &t.UserId, &t.Category, &t.Status, &t.Subject, &createdAtStr)
		if err != nil {
			continue
		}
		t.CreatedAt = createdAtStr
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// gRPC Сервер
// ─────────────────────────────────────────────────────────────────────────────

type supportServer struct {
	pb.UnimplementedSupportServiceServer
	repo *PostgresRepo
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

func checkAdminRole(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	roles := md.Get("x-role")
	if len(roles) == 0 || roles[0] != "admin" {
		return status.Error(codes.PermissionDenied, "admin role required")
	}

	return nil
}

// Пользовательские методы
func (s *supportServer) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.Ticket, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Category != "bug" && req.Category != "proposal" && req.Category != "complaint" {
		return nil, status.Error(codes.InvalidArgument, "category must be bug, proposal, or complaint")
	}

	if req.Subject == "" {
		return nil, status.Error(codes.InvalidArgument, "subject is required")
	}

	ticket, err := s.repo.CreateTicket(userID, req.Category, req.Subject)
	if err != nil {
		log.Printf("CreateTicket error: %v", err)
		return nil, status.Error(codes.Internal, "failed to create ticket")
	}

	log.Printf("Ticket created: id=%d, user_id=%d, category=%s", ticket.Id, userID, req.Category)
	return ticket, nil
}

func (s *supportServer) ListMyTickets(ctx context.Context, req *pb.ListTicketsRequest) (*pb.ListTicketsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tickets, err := s.repo.ListTicketsByUser(userID)
	if err != nil {
		log.Printf("ListMyTickets error: %v", err)
		return nil, status.Error(codes.Internal, "failed to list tickets")
	}

	return &pb.ListTicketsResponse{Tickets: tickets}, nil
}

func (s *supportServer) GetTicket(ctx context.Context, req *pb.GetTicketRequest) (*pb.Ticket, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ticket, messages, err := s.repo.GetTicketByID(req.TicketId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "ticket not found")
		}
		log.Printf("GetTicket error: %v", err)
		return nil, status.Error(codes.Internal, "failed to get ticket")
	}

	// Проверяем доступ (владелец или админ)
	if ticket.UserId != userID {
		if err := checkAdminRole(ctx); err != nil {
			return nil, err
		}
	}

	ticket.Messages = messages
	return ticket, nil
}

func (s *supportServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что тикет существует и принадлежит пользователю
	ticket, _, err := s.repo.GetTicketByID(req.TicketId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "ticket not found")
		}
		return nil, status.Error(codes.Internal, "failed to verify ticket")
	}

	if ticket.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "not your ticket")
	}

	if req.Text == "" {
		return nil, status.Error(codes.InvalidArgument, "message text is required")
	}

	message, err := s.repo.AddMessage(req.TicketId, "user", req.Text)
	if err != nil {
		log.Printf("SendMessage error: %v", err)
		return nil, status.Error(codes.Internal, "failed to send message")
	}

	log.Printf("Message sent: ticket_id=%d, user_id=%d", req.TicketId, userID)
	return message, nil
}

// Админские методы
func (s *supportServer) GetStats(ctx context.Context, req *pb.Empty) (*pb.StatsResponse, error) {
	if err := checkAdminRole(ctx); err != nil {
		return nil, err
	}

	open, inProgress, closed, byCategory, err := s.repo.GetStats()
	if err != nil {
		log.Printf("GetStats error: %v", err)
		return nil, status.Error(codes.Internal, "failed to get stats")
	}

	return &pb.StatsResponse{
		OpenCount:        open,
		InProgressCount:  inProgress,
		ClosedCount:      closed,
		ByCategory:       byCategory,
		TotalTickets:     open + inProgress + closed,
	}, nil
}

func (s *supportServer) ListOpenTickets(ctx context.Context, req *pb.ListTicketsRequest) (*pb.ListTicketsResponse, error) {
	if err := checkAdminRole(ctx); err != nil {
		return nil, err
	}

	tickets, err := s.repo.ListOpenTickets()
	if err != nil {
		log.Printf("ListOpenTickets error: %v", err)
		return nil, status.Error(codes.Internal, "failed to list open tickets")
	}

	return &pb.ListTicketsResponse{Tickets: tickets}, nil
}

func (s *supportServer) ReplyAsAdmin(ctx context.Context, req *pb.ReplyAsAdminRequest) (*pb.Message, error) {
	if err := checkAdminRole(ctx); err != nil {
		return nil, err
	}

	if req.Text == "" {
		return nil, status.Error(codes.InvalidArgument, "reply text is required")
	}

	_, _, err := s.repo.GetTicketByID(req.TicketId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "ticket not found")
		}
		return nil, status.Error(codes.Internal, "failed to verify ticket")
	}

	message, err := s.repo.AddMessage(req.TicketId, "admin", req.Text)
	if err != nil {
		log.Printf("ReplyAsAdmin error: %v", err)
		return nil, status.Error(codes.Internal, "failed to send reply")
	}

	log.Printf("Admin replied: ticket_id=%d", req.TicketId)
	return message, nil
}

func (s *supportServer) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.Ticket, error) {
	if err := checkAdminRole(ctx); err != nil {
		return nil, err
	}

	if req.Status != "open" && req.Status != "in_progress" && req.Status != "closed" {
		return nil, status.Error(codes.InvalidArgument, "status must be open, in_progress, or closed")
	}

	ticket, messages, err := s.repo.GetTicketByID(req.TicketId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "ticket not found")
		}
		return nil, status.Error(codes.Internal, "failed to find ticket")
	}

	err = s.repo.UpdateTicketStatus(req.TicketId, req.Status)
	if err != nil {
		log.Printf("UpdateTicketStatus error: %v", err)
		return nil, status.Error(codes.Internal, "failed to update status")
	}

	ticket.Status = req.Status
	ticket.Messages = messages

	log.Printf("Ticket status updated: id=%d, status=%s", req.TicketId, req.Status)
	return ticket, nil
}

//TODO: сделать функционал вместо заглушки
func (s *supportServer) SubscribeToTicket(req *pb.SubscribeRequest, stream pb.SupportService_SubscribeToTicketServer) error {
	return status.Error(codes.Unimplemented, "SubscribeToTicket not implemented yet")
}

func main() {
	dbURL := os.Getenv("SUPPORT_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:1111@postgres:5432/texnopark?sslmode=disable"
	}

	repo, err := NewPostgresRepo(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer repo.Close()
	log.Println("Connected to PostgreSQL")

	grpcServer := grpc.NewServer()
	pb.RegisterSupportServiceServer(grpcServer, &supportServer{repo: repo})

	listener, err := net.Listen("tcp", ":8084")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server for support service is running on port 8084...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
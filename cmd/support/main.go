package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"
	"time"

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

func mapCategoryToID(category string) int64 {
	switch category {
	case "bug":
		return 1
	case "proposal":
		return 2
	case "complaint":
		return 3
	default:
		return 1
	}
}

func mapIDToCategory(id int64) string {
	switch id {
	case 1:
		return "bug"
	case 2:
		return "proposal"
	case 3:
		return "complaint"
	default:
		return "bug"
	}
}

func mapStatusToID(status string) int64 {
	switch status {
	case "open":
		return 1
	case "in_progress":
		return 2
	case "closed":
		return 3
	default:
		return 1
	}
}

func mapIDToStatus(id int64) string {
	switch id {
	case 1:
		return "open"
	case 2:
		return "in_progress"
	case 3:
		return "closed"
	default:
		return "open"
	}
}

func (r *PostgresRepo) CreateTicket(userID int64, category, subject, body string) (*pb.Ticket, error) {
	categoryID := mapCategoryToID(category)
	statusID := mapStatusToID("open")

	query := `INSERT INTO ticket (user_id, category_id, status_id, title, body, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) 
			  RETURNING id, created_at`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRow(query, userID, categoryID, statusID, subject, body).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	return &pb.Ticket{
		Id:        id,
		UserId:    userID,
		Category:  category,
		Status:    "open",
		Subject:   subject,
		Body:      body,
		CreatedAt: createdAt.Format(time.RFC3339),
	}, nil
}

func (r *PostgresRepo) ListTicketsByUser(userID int64) ([]*pb.Ticket, error) {
	query := `SELECT t.id, t.user_id, tc.name, ts.name, t.title, t.body, t.created_at 
			  FROM ticket t
			  JOIN ticket_category tc ON t.category_id = tc.id
			  JOIN ticket_status ts ON t.status_id = ts.id
			  WHERE t.user_id = $1
			  ORDER BY t.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*pb.Ticket
	for rows.Next() {
		var t pb.Ticket
		var categoryName, statusName string
		var createdAt time.Time

		err := rows.Scan(&t.Id, &t.UserId, &categoryName, &statusName, &t.Subject, &t.Body, &createdAt)
		if err != nil {
			continue
		}

		t.Category = mapIDToCategory(mapCategoryNameToID(categoryName))
		t.Status = mapIDToStatus(mapStatusNameToID(statusName))
		t.CreatedAt = createdAt.Format(time.RFC3339)
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *PostgresRepo) GetTicketByID(ticketID int64) (*pb.Ticket, []*pb.Message, error) {
	ticketQuery := `SELECT t.id, t.user_id, tc.name, ts.name, t.title, t.body, t.created_at 
					FROM ticket t
					JOIN ticket_category tc ON t.category_id = tc.id
					JOIN ticket_status ts ON t.status_id = ts.id
					WHERE t.id = $1`

	var ticket pb.Ticket
	var categoryName, statusName string
	var createdAt time.Time

	err := r.db.QueryRow(ticketQuery, ticketID).Scan(
		&ticket.Id, &ticket.UserId, &categoryName, &statusName,
		&ticket.Subject, &ticket.Body, &createdAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, err
		}
		return nil, nil, err
	}

	ticket.Category = mapIDToCategory(mapCategoryNameToID(categoryName))
	ticket.Status = mapIDToStatus(mapStatusNameToID(statusName))
	ticket.CreatedAt = createdAt.Format(time.RFC3339)

	msgQuery := `SELECT id, ticket_id, sender_id, content, created_at 
				 FROM message WHERE ticket_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.Query(msgQuery, ticketID)
	if err != nil {
		return &ticket, nil, nil
	}
	defer rows.Close()

	var messages []*pb.Message
	for rows.Next() {
		var msg pb.Message
		var senderID int64
		var msgCreatedAt time.Time

		err := rows.Scan(&msg.Id, &msg.TicketId, &senderID, &msg.Text, &msgCreatedAt)
		if err != nil {
			continue
		}

		msg.AuthorType = "user"
		msg.CreatedAt = msgCreatedAt.Format(time.RFC3339)
		messages = append(messages, &msg)
	}

	return &ticket, messages, nil
}

func (r *PostgresRepo) AddMessage(ticketID int64, senderID int64, content string) (*pb.Message, error) {
	query := `INSERT INTO message (ticket_id, sender_id, content, created_at) 
			  VALUES ($1, $2, $3, NOW()) 
			  RETURNING id, created_at`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRow(query, ticketID, senderID, content).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	return &pb.Message{
		Id:         id,
		TicketId:   ticketID,
		AuthorType: "user",
		Text:       content,
		CreatedAt:  createdAt.Format(time.RFC3339),
	}, nil
}

func (r *PostgresRepo) UpdateTicketStatus(ticketID int64, status string) error {
	statusID := mapStatusToID(status)
	query := `UPDATE ticket SET status_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, statusID, ticketID)
	return err
}

func (r *PostgresRepo) GetStats() (open, inProgress, closed int64, byCategory map[string]int64, err error) {
	statusQuery := `SELECT ts.name, COUNT(*) 
					FROM ticket t
					JOIN ticket_status ts ON t.status_id = ts.id
					GROUP BY ts.name`

	rows, err := r.db.Query(statusQuery)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var count int64
		rows.Scan(&name, &count)
		switch name {
		case "Открыто":
			open = count
		case "В работе":
			inProgress = count
		case "Закрыто":
			closed = count
		}
	}

	catQuery := `SELECT tc.name, COUNT(*) 
				 FROM ticket t
				 JOIN ticket_category tc ON t.category_id = tc.id
				 GROUP BY tc.name`

	catRows, err := r.db.Query(catQuery)
	if err != nil {
		return open, inProgress, closed, nil, err
	}
	defer catRows.Close()

	byCategory = make(map[string]int64)
	for catRows.Next() {
		var name string
		var count int64
		catRows.Scan(&name, &count)
		key := mapCategoryNameToKey(name)
		byCategory[key] = count
	}

	return open, inProgress, closed, byCategory, nil
}

func (r *PostgresRepo) ListOpenTickets() ([]*pb.Ticket, error) {
	query := `SELECT t.id, t.user_id, tc.name, ts.name, t.title, t.body, t.created_at 
			  FROM ticket t
			  JOIN ticket_category tc ON t.category_id = tc.id
			  JOIN ticket_status ts ON t.status_id = ts.id
			  WHERE ts.name IN ('Открыто', 'В работе')
			  ORDER BY t.created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*pb.Ticket
	for rows.Next() {
		var t pb.Ticket
		var categoryName, statusName string
		var createdAt time.Time

		err := rows.Scan(&t.Id, &t.UserId, &categoryName, &statusName, &t.Subject, &t.Body, &createdAt)
		if err != nil {
			continue
		}

		t.Category = mapIDToCategory(mapCategoryNameToID(categoryName))
		t.Status = mapIDToStatus(mapStatusNameToID(statusName))
		t.CreatedAt = createdAt.Format(time.RFC3339)
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func mapCategoryNameToID(name string) int64 {
	switch name {
	case "Баг":
		return 1
	case "Предложение":
		return 2
	case "Продуктовая жалоба":
		return 3
	default:
		return 1
	}
}

func mapStatusNameToID(name string) int64 {
	switch name {
	case "Открыто":
		return 1
	case "В работе":
		return 2
	case "Закрыто":
		return 3
	default:
		return 1
	}
}

func mapCategoryNameToKey(name string) string {
	switch name {
	case "Баг":
		return "bug"
	case "Предложение":
		return "proposal"
	case "Продуктовая жалоба":
		return "complaint"
	default:
		return "bug"
	}
}

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

func (s *supportServer) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.Ticket, error) {
	userID := req.UserId

	if req.Category != "bug" && req.Category != "proposal" && req.Category != "complaint" {
		return nil, status.Error(codes.InvalidArgument, "category must be bug, proposal, or complaint")
	}

	if req.Subject == "" {
		return nil, status.Error(codes.InvalidArgument, "subject is required")
	}

	if req.Body == "" {
		return nil, status.Error(codes.InvalidArgument, "body is required")
	}

	ticket, err := s.repo.CreateTicket(userID, req.Category, req.Subject, req.Body)
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

	message, err := s.repo.AddMessage(req.TicketId, userID, req.Text)
	if err != nil {
		log.Printf("SendMessage error: %v", err)
		return nil, status.Error(codes.Internal, "failed to send message")
	}

	log.Printf("Message sent: ticket_id=%d, user_id=%d", req.TicketId, userID)
	return message, nil
}

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

	adminID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	message, err := s.repo.AddMessage(req.TicketId, adminID, req.Text)
	if err != nil {
		log.Printf("ReplyAsAdmin error: %v", err)
		return nil, status.Error(codes.Internal, "failed to send reply")
	}

	message.AuthorType = "admin"
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
package repository

import (
    "database/sql"
    "time"
    
    _ "github.com/lib/pq"
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
    
    query := `INSERT INTO ticket (user_id, category_id, title, body, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, NOW(), NOW()) 
              RETURNING id, created_at`
    
    var id int64
    var createdAt time.Time
    err := r.db.QueryRow(query, userID, categoryID, subject, body).Scan(&id, &createdAt)
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
    
    ticket.Category = mapCategoryToKey(categoryName)
    ticket.Status = mapStatusToKey(statusName)
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
        
        t.Category = mapCategoryToKey(categoryName)
        t.Status = mapStatusToKey(statusName)
        t.CreatedAt = createdAt.Format(time.RFC3339)
        tickets = append(tickets, &t)
    }
    return tickets, nil
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
    // Статусы
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
        // Преобразуем русские названия в английские ключи для proto
        key := mapCategoryToKey(name)
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
        
        t.Category = mapCategoryToKey(categoryName)
        t.Status = mapStatusToKey(statusName)
        t.CreatedAt = createdAt.Format(time.RFC3339)
        tickets = append(tickets, &t)
    }
    return tickets, nil
}

func mapCategoryToKey(russianName string) string {
    switch russianName {
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

func mapStatusToKey(russianName string) string {
    switch russianName {
    case "Открыто":
        return "open"
    case "В работе":
        return "in_progress"
    case "Закрыто":
        return "closed"
    default:
        return "open"
    }
}
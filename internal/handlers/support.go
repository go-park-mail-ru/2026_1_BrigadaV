package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"guidely-app/internal/middleware"
	pb "guidely-app/pkg/pb/support/v1"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/metadata"
)

type SupportHandlers struct {
	supportClient pb.SupportServiceClient
}

func NewSupportHandlers(client pb.SupportServiceClient) *SupportHandlers {
	return &SupportHandlers{supportClient: client}
}

func (h *SupportHandlers) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Category string `json:"category"`
		Subject  string `json:"subject"`
		Body     string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	md := metadata.Pairs("x-user-id", strconv.FormatInt(userID, 10))
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.CreateTicket(ctx, &pb.CreateTicketRequest{
		UserId: userID,
		Category: req.Category,
		Subject:  req.Subject,
		Body:     req.Body,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SupportHandlers) ListMyTickets(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.Pairs("x-user-id", strconv.FormatInt(userID, 10))
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.ListMyTickets(ctx, &pb.ListTicketsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp.Tickets)
}

func (h *SupportHandlers) GetTicket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	ticketID, _ := strconv.ParseInt(vars["id"], 10, 64)

	md := metadata.Pairs("x-user-id", strconv.FormatInt(userID, 10))
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.GetTicket(ctx, &pb.GetTicketRequest{TicketId: ticketID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *SupportHandlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	ticketID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	md := metadata.Pairs("x-user-id", strconv.FormatInt(userID, 10))
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.SendMessage(ctx, &pb.SendMessageRequest{
		TicketId: ticketID,
		Text:     req.Text,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *SupportHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	md := metadata.Pairs("x-role", "admin")
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.GetStats(ctx, &pb.Empty{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *SupportHandlers) ListOpenTickets(w http.ResponseWriter, r *http.Request) {
	md := metadata.Pairs("x-role", "admin")
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.ListOpenTickets(ctx, &pb.ListTicketsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(resp.Tickets)
}

func (h *SupportHandlers) ReplyAsAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	md := metadata.Pairs("x-role", "admin")
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.ReplyAsAdmin(ctx, &pb.ReplyAsAdminRequest{
		TicketId: ticketID,
		Text:     req.Text,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *SupportHandlers) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID, _ := strconv.ParseInt(vars["id"], 10, 64)

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	md := metadata.Pairs("x-role", "admin")
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.supportClient.UpdateTicketStatus(ctx, &pb.UpdateTicketStatusRequest{
		TicketId: ticketID,
		Status:   req.Status,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}
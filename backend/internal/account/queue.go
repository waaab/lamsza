package account

import (
	"encoding/json"
)

type queueUnpublished struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	OwnerEmail string `json:"owner_email"`
}

type queueMember struct {
	EntryID   int    `json:"entry_id"`
	EntryName string `json:"entry_name"`
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
}

type queueWebsite struct {
	ID          int    `json:"id"`
	Domain      string `json:"domain"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Submitter   string `json:"submitter"`
}

type queueClaim struct {
	EntryID   int    `json:"entry_id"`
	EntryName string `json:"entry_name"`
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
}

type queueSuggestion struct {
	ID        int             `json:"id"`
	EntryID   int             `json:"entry_id"`
	EntryName string          `json:"entry_name"`
	Email     string          `json:"email"`
	Changes   json.RawMessage `json:"changes"`
	Note      string          `json:"note"`
}

type queueResponse struct {
	Unpublished []queueUnpublished `json:"unpublished"`
	Members     []queueMember      `json:"members"`
	Claims      []queueClaim       `json:"claims"`
	Suggestions []queueSuggestion  `json:"suggestions"`
	Websites    []queueWebsite     `json:"websites"`
}

type publishBody struct {
	EntryID int `json:"entry_id"`
}

type memberActionBody struct {
	EntryID int    `json:"entry_id"`
	UserID  int    `json:"user_id"`
	Action  string `json:"action"`
}

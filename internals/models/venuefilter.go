package models

type VenueFilter struct {
	City      string
	HostID    string
	VenueID   string
	IsBlocked bool
}

type UpdateVenueData struct {
	Name                 *string `json:"name,omitempty"`
	City                 *string `json:"city,omitempty"`
	State                *string `json:"state,omitempty"`
	IsSeatLayoutRequired *bool   `json:"is_seat_layout_required,omitempty"`
	IsBlocked            *bool   `json:"is_blocked,omitempty"`
}

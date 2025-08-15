package netbird

type NetbirdUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	LastLogin string `json:"last_login"`
}

type NetbirdPeer struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	UserID   string `json:"user_id"`
}

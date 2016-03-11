package nginx

// Instance represents a single NGINX server
type Server struct {
	Sites []Site
}

// Site represents a single virtual host
type Site struct {
	Host    string
	Address string
	Port    string
}

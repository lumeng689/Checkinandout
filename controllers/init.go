package controllers

// CCServer - root struct for the entire server
type CCServer struct {
	Config Config
}

// InitServer - return the reference to a server instance
func InitServer() CCServer {
	return CCServer{}
}

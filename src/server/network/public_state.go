package network

// the actual public server object
var PublicServer = PublicServerType{
	Subscribers: map[string]*ResponseStream{},
}

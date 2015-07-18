package objects

import "github.com/daemonl/informer/objects/server"

type ServerCheck struct {
	Informants
	server.Server
}

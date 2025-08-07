package api

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestServerRoutes(t *testing.T) {
	server, _ := SetupTestServer()
	router := server.Routes()

	assert.IsType(t, &mux.Router{}, router)
}

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	router := gin.Default()
	router.GET("/", s.HelloWorldHandler)
	router.GET("/health", s.healthHandler)

	// apiRoutes := router.Group("/api")
	{
		// apiRoutes.GET("/entries", )  // Entries with pagination & filter
		// apiRoutes.GET("/entries/:id", ) // Only get one entry
		// apiRoutes.POST("/entries") // Create a new entry
		// apiRoutes.PATCH("/entries/:id") // Modify entry
		// apiRoutes.GET("/changes/:id", ) // Get all the logged changes

		// apiRoutes.POST("/addcolumn/")
		// apiRoutes.POST("/removecolumn/")

		// apiRoutes.GET("/pdformats")
		// apiRoutes.POST("/pdformats")
		// apiRoutes.PATCH("/pdformats")

		// apiRoutes.GET("/pdf") // Get pdf to entry

	}

	return router
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

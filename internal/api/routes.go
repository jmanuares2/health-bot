package api

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *Handlers) {
	api := r.Group("/api")
	{
		api.GET("/today", h.Today)
		api.GET("/progress", h.Progress)
		api.GET("/gym", h.GymList)
		api.GET("/gym/:exercise", h.GymExercise)
		api.GET("/body-weight", h.BodyWeight)
	}
}

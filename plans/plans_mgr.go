package plans

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlanManager struct {
	ctx       context.Context
	apiEngine *gin.Engine
	db        *gorm.DB
}

func NewPlanManager(ctx context.Context, apiEngine *gin.Engine, db *gorm.DB) *PlanManager {
	return &PlanManager{db: db}
}

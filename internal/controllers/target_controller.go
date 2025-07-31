package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mail-dispatcher/internal/models"
)

// TargetController 转发目标控制器
type TargetController struct {
	db *gorm.DB
}

// NewTargetController 创建转发目标控制器
func NewTargetController(db *gorm.DB) *TargetController {
	return &TargetController{db: db}
}

// GetTargets 获取所有转发目标
func (c *TargetController) GetTargets(ctx *gin.Context) {
	var targets []models.ForwardTarget
	if err := c.db.Find(&targets).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取转发目标失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  targets,
		"total": len(targets),
	})
}

// GetTarget 获取单个转发目标
func (c *TargetController) GetTarget(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var target models.ForwardTarget
	if err := c.db.First(&target, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "转发目标不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": target})
}

// CreateTarget 创建转发目标
func (c *TargetController) CreateTarget(ctx *gin.Context) {
	var target models.ForwardTarget
	if err := ctx.ShouldBindJSON(&target); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if target.Name == "" || target.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "名称和邮箱地址不能为空"})
		return
	}

	// 检查名称是否已存在
	var existingTarget models.ForwardTarget
	if err := c.db.Where("name = ?", target.Name).First(&existingTarget).Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "转发目标名称已存在"})
		return
	}

	if err := c.db.Create(&target).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建转发目标失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": target})
}

// UpdateTarget 更新转发目标
func (c *TargetController) UpdateTarget(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var target models.ForwardTarget
	if err := c.db.First(&target, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "转发目标不存在"})
		return
	}

	var updateData models.ForwardTarget
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 如果更新名称，检查是否与其他记录冲突
	if updateData.Name != "" && updateData.Name != target.Name {
		var existingTarget models.ForwardTarget
		if err := c.db.Where("name = ? AND id != ?", updateData.Name, id).First(&existingTarget).Error; err == nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "转发目标名称已存在"})
			return
		}
	}

	// 更新字段
	if updateData.Name != "" {
		target.Name = updateData.Name
	}
	if updateData.Email != "" {
		target.Email = updateData.Email
	}
	if updateData.Description != "" {
		target.Description = updateData.Description
	}

	if err := c.db.Save(&target).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新转发目标失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": target})
}

// DeleteTarget 删除转发目标
func (c *TargetController) DeleteTarget(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var target models.ForwardTarget
	if err := c.db.First(&target, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "转发目标不存在"})
		return
	}

	if err := c.db.Delete(&target).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除转发目标失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "转发目标删除成功"})
}

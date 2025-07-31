package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"mail-dispatcher/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AccountController 邮箱账户控制器
type AccountController struct {
	db *gorm.DB
}

// NewAccountController 创建邮箱账户控制器
func NewAccountController(db *gorm.DB) *AccountController {
	return &AccountController{db: db}
}

// getIMAPServer 根据邮箱地址自动获取对应的IMAP服务器
func getIMAPServer(email string) string {
	email = strings.ToLower(email)

	switch {
	case strings.Contains(email, "@gmail.com"):
		return "imap.gmail.com:993"
	case strings.Contains(email, "@qq.com"):
		return "imap.qq.com:993"
	case strings.Contains(email, "@163.com"):
		return "imap.163.com:993"
	case strings.Contains(email, "@126.com"):
		return "imap.126.com:993"
	case strings.Contains(email, "@outlook.com") || strings.Contains(email, "@hotmail.com"):
		return "outlook.office365.com:993"
	case strings.Contains(email, "@yahoo.com"):
		return "imap.mail.yahoo.com:993"
	case strings.Contains(email, "@sina.com"):
		return "imap.sina.com:993"
	case strings.Contains(email, "@sohu.com"):
		return "imap.sohu.com:993"
	default:
		// 对于其他邮箱，尝试使用通用的IMAP服务器
		// 用户可以在创建后手动修改
		return "imap.gmail.com:993"
	}
}

// GetAccounts 获取所有邮箱账户
func (c *AccountController) GetAccounts(ctx *gin.Context) {
	var accounts []models.MailAccount
	if err := c.db.Find(&accounts).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取邮箱账户失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  accounts,
		"total": len(accounts),
	})
}

// GetAccount 获取单个邮箱账户
func (c *AccountController) GetAccount(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var account models.MailAccount
	if err := c.db.First(&account, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "邮箱账户不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": account})
}

// CreateAccount 创建邮箱账户
func (c *AccountController) CreateAccount(ctx *gin.Context) {
	var account models.MailAccount
	if err := ctx.ShouldBindJSON(&account); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 验证必填字段
	if account.Address == "" || account.Username == "" || account.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "邮箱地址、用户名和密码不能为空"})
		return
	}

	// 检查邮箱地址是否已存在
	var existingAccount models.MailAccount
	if err := c.db.Where("address = ?", account.Address).First(&existingAccount).Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "邮箱地址已存在"})
		return
	}

	// 设置默认值
	if account.Server == "" {
		account.Server = getIMAPServer(account.Address)
	}

	if err := c.db.Create(&account).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建邮箱账户失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": account})
}

// UpdateAccount 更新邮箱账户
func (c *AccountController) UpdateAccount(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var account models.MailAccount
	if err := c.db.First(&account, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "邮箱账户不存在"})
		return
	}

	var updateData models.MailAccount
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 更新字段
	if updateData.Address != "" {
		account.Address = updateData.Address
		// 如果邮箱地址改变，自动更新服务器
		account.Server = getIMAPServer(updateData.Address)
	}
	if updateData.Username != "" {
		account.Username = updateData.Username
	}
	if updateData.Password != "" {
		account.Password = updateData.Password
	}
	if updateData.Server != "" {
		account.Server = updateData.Server
	}
	if updateData.Settings != "" {
		account.Settings = updateData.Settings
	}

	if err := c.db.Save(&account).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新邮箱账户失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": account})
}

// DeleteAccount 删除邮箱账户
func (c *AccountController) DeleteAccount(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var account models.MailAccount
	if err := c.db.First(&account, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "邮箱账户不存在"})
		return
	}

	if err := c.db.Delete(&account).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除邮箱账户失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "邮箱账户删除成功"})
}

// ToggleAccountStatus 切换账户状态
func (c *AccountController) ToggleAccountStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var account models.MailAccount
	if err := c.db.First(&account, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "邮箱账户不存在"})
		return
	}

	account.IsActive = !account.IsActive

	if err := c.db.Save(&account).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新账户状态失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    account,
		"message": "账户状态更新成功",
	})
}

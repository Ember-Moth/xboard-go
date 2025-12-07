package handler

import (
	"net/http"

	"xboard/internal/service"

	"github.com/gin-gonic/gin"
)

// GuestRegister 用户注册
func GuestRegister(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email      string `json:"email" binding:"required,email"`
			Password   string `json:"password" binding:"required,min=6"`
			InviteCode string `json:"invite_code"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 处理邀请码
		var inviteUserID *int64
		if req.InviteCode != "" {
			inviteCode, err := services.Invite.ValidateInviteCode(req.InviteCode)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请码"})
				return
			}
			inviteUserID = &inviteCode.UserID
		}

		user, err := services.User.Register(req.Email, req.Password, inviteUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 标记邀请码已使用
		if req.InviteCode != "" {
			services.Invite.UseInviteCode(req.InviteCode, user.ID)
		}

		token, err := services.Auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"token": token,
			},
		})
	}
}

// GuestLogin 用户登录
func GuestLogin(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := services.User.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := services.Auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"token":    token,
				"is_admin": user.IsAdmin,
			},
		})
	}
}

// GuestGetPlans 获取可购买套餐列表
func GuestGetPlans(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		plans, err := services.Plan.GetAvailable()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result := make([]map[string]interface{}, 0, len(plans))
		for _, plan := range plans {
			result = append(result, services.Plan.GetPlanInfo(&plan))
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// PassportLogin Passport 登录
func PassportLogin(services *service.Services) gin.HandlerFunc {
	return GuestLogin(services)
}

// PassportRegister Passport 注册
func PassportRegister(services *service.Services) gin.HandlerFunc {
	return GuestRegister(services)
}


// GetNotices 获取公告列表
func GetNotices(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		notices, err := services.Notice.GetPublic()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": notices})
	}
}

// GetKnowledge 获取知识库列表
func GetKnowledge(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		items, err := services.Knowledge.GetPublic(category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items})
	}
}

// GetKnowledgeCategories 获取知识库分类
func GetKnowledgeCategories(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := services.Knowledge.GetCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": categories})
	}
}

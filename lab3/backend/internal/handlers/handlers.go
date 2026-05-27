package handlers

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strconv"
	"time"
	"transport-auth-server/backend/internal/auth"
	"transport-auth-server/backend/internal/models"
)

type H struct{ DB *gorm.DB }
type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h H) Login(c *gin.Context) {
	var r loginReq
	if c.ShouldBindJSON(&r) != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	var u models.User
	if h.DB.Where("login=?", r.Login).First(&u).Error != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(r.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid login or password"})
		return
	}
	tok, _ := auth.Generate(u.ID, u.Login, u.IsAdmin)
	c.JSON(200, gin.H{"token": tok, "user": u})
}
func id(c *gin.Context) uint { v, _ := strconv.Atoi(c.Param("id")); return uint(v) }

func ctxUserID(c *gin.Context) uint {
	if v, ok := c.Get("user_id"); ok {
		switch x := v.(type) {
		case uint:
			return x
		case int:
			if x >= 0 {
				return uint(x)
			}
		case float64:
			return uint(x)
		}
	}
	return 0
}
func ctxIsAdmin(c *gin.Context) bool {
	v, _ := c.Get("is_admin")
	return v == true
}
func crud[T any](h H, preload ...string) (list, get, create, update, del gin.HandlerFunc) {
	apply := func(q *gorm.DB) *gorm.DB {
		for _, p := range preload {
			q = q.Preload(p)
		}
		return q
	}
	list = func(c *gin.Context) { var items []T; q := apply(h.DB); q.Find(&items); c.JSON(200, items) }
	get = func(c *gin.Context) {
		var item T
		if apply(h.DB).First(&item, id(c)).Error != nil {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.JSON(200, item)
	}
	create = func(c *gin.Context) {
		var item T
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(400, gin.H{"error": "bad json", "detail": err.Error()})
			return
		}
		h.DB.Create(&item)
		c.JSON(201, item)
	}
	update = func(c *gin.Context) {
		var item T
		if h.DB.First(&item, id(c)).Error != nil {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(400, gin.H{"error": "bad json", "detail": err.Error()})
			return
		}
		h.DB.Save(&item)
		c.JSON(200, item)
	}
	del = func(c *gin.Context) { var item T; h.DB.Delete(&item, id(c)); c.Status(204) }
	return
}
func (h H) CreateUser(c *gin.Context) {
	var u models.User
	if c.ShouldBindJSON(&u) != nil || u.Password == "" {
		c.JSON(400, gin.H{"error": "login and password required"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	u.PasswordHash = string(hash)
	h.DB.Create(&u)
	c.JSON(201, u)
}
func (h H) UpdateUser(c *gin.Context) {
	current, _ := c.Get("user_id")
	admin, _ := c.Get("is_admin")
	if admin != true && uint(current.(uint)) != id(c) {
		c.JSON(403, gin.H{"error": "cannot edit other users"})
		return
	}
	var u models.User
	if h.DB.First(&u, id(c)).Error != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var in models.User
	if c.ShouldBindJSON(&in) != nil {
		c.JSON(400, gin.H{"error": "bad json"})
		return
	}
	u.Login = in.Login
	u.Name = in.Name
	if admin == true {
		u.IsAdmin = in.IsAdmin
	}
	if in.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		u.PasswordHash = string(hash)
	}
	h.DB.Save(&u)
	c.JSON(200, u)
}

type payReq struct {
	CardNumber     string `json:"card_number"`
	Amount         int64  `json:"amount"`
	TerminalSerial string `json:"terminal_serial"`
}

func (h H) AuthorizePayment(c *gin.Context) {
	var r payReq
	if c.ShouldBindJSON(&r) != nil || r.Amount <= 0 {
		c.JSON(400, gin.H{"authorized": false, "message": "bad request"})
		return
	}
	var card models.Card
	if h.DB.Where("number=?", r.CardNumber).First(&card).Error != nil {
		c.JSON(404, gin.H{"authorized": false, "message": "card not found"})
		return
	}
	var term models.Terminal
	if h.DB.Where("serial_number=?", r.TerminalSerial).First(&term).Error != nil {
		c.JSON(404, gin.H{"authorized": false, "message": "terminal not found"})
		return
	}
	status := "approved"
	msg := "authorized"
	ok := true
	if card.Blocked {
		status = "declined"
		msg = "card blocked"
		ok = false
	} else if card.Balance < r.Amount {
		status = "declined"
		msg = "insufficient funds"
		ok = false
	} else {
		card.Balance -= r.Amount
		h.DB.Save(&card)
	}
	tx := models.Transaction{Amount: r.Amount, CardID: card.ID, TerminalID: term.ID, Status: status, Message: msg, CreatedAt: time.Now()}
	h.DB.Create(&tx)
	c.JSON(200, gin.H{"authorized": ok, "message": msg, "transaction_id": tx.ID, "balance": card.Balance})
}
func (h H) DownloadKeys(c *gin.Context) {
	var keys []models.CardKey
	h.DB.Find(&keys)
	c.JSON(200, keys)
}

func (h H) listCards(c *gin.Context) {
	var items []models.Card
	q := h.DB.Preload("Key")
	if !ctxIsAdmin(c) {
		q = q.Where("user_id = ?", ctxUserID(c))
	}
	q.Find(&items)
	c.JSON(200, items)
}
func (h H) getCard(c *gin.Context) {
	var item models.Card
	if err := h.DB.Preload("Key").First(&item, id(c)).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if !ctxIsAdmin(c) && item.UserID != ctxUserID(c) {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, item)
}
func (h H) createCard(c *gin.Context) {
	var item models.Card
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": "bad json", "detail": err.Error()})
		return
	}
	if !ctxIsAdmin(c) {
		item.UserID = ctxUserID(c)
	}
	h.DB.Create(&item)
	c.JSON(201, item)
}
func (h H) updateCard(c *gin.Context) {
	var existing models.Card
	if h.DB.First(&existing, id(c)).Error != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if !ctxIsAdmin(c) && existing.UserID != ctxUserID(c) {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var item models.Card
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": "bad json", "detail": err.Error()})
		return
	}
	item.ID = existing.ID
	if !ctxIsAdmin(c) {
		item.UserID = existing.UserID
	} else if item.UserID == 0 {
		item.UserID = existing.UserID
	}
	h.DB.Save(&item)
	c.JSON(200, item)
}
func (h H) deleteCard(c *gin.Context) {
	var item models.Card
	if h.DB.First(&item, id(c)).Error != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if !ctxIsAdmin(c) && item.UserID != ctxUserID(c) {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&models.Card{}, id(c))
	c.Status(204)
}

func (h H) listTransactions(c *gin.Context) {
	var items []models.Transaction
	q := h.DB.Preload("Card").Preload("Terminal")
	if !ctxIsAdmin(c) {
		uid := ctxUserID(c)
		q = q.Joins("JOIN cards ON cards.id = transactions.card_id").Where("cards.user_id = ?", uid)
	}
	q.Find(&items)
	c.JSON(200, items)
}
func (h H) getTransaction(c *gin.Context) {
	var item models.Transaction
	if err := h.DB.Preload("Card").Preload("Terminal").First(&item, id(c)).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if !ctxIsAdmin(c) && item.Card.UserID != ctxUserID(c) {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, item)
}
func (h H) createTransaction(c *gin.Context) {
	var item models.Transaction
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": "bad json", "detail": err.Error()})
		return
	}
	if !ctxIsAdmin(c) {
		var card models.Card
		if h.DB.First(&card, item.CardID).Error != nil || card.UserID != ctxUserID(c) {
			c.JSON(403, gin.H{"error": "forbidden"})
			return
		}
	}
	h.DB.Create(&item)
	c.JSON(201, item)
}
func (h H) updateTransaction(c *gin.Context) {
	var existing models.Transaction
	if err := h.DB.Preload("Card").First(&existing, id(c)).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if !ctxIsAdmin(c) && existing.Card.UserID != ctxUserID(c) {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var item models.Transaction
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": "bad json", "detail": err.Error()})
		return
	}
	if !ctxIsAdmin(c) {
		var newCard models.Card
		if h.DB.First(&newCard, item.CardID).Error != nil || newCard.UserID != ctxUserID(c) {
			c.JSON(403, gin.H{"error": "forbidden"})
			return
		}
	}
	item.ID = existing.ID
	h.DB.Save(&item)
	c.JSON(200, item)
}
func (h H) deleteTransaction(c *gin.Context) {
	var existing models.Transaction
	if err := h.DB.Preload("Card").First(&existing, id(c)).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if !ctxIsAdmin(c) && existing.Card.UserID != ctxUserID(c) {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&models.Transaction{}, id(c))
	c.Status(204)
}

func MountCRUD(api *gin.RouterGroup, h H, admin gin.HandlerFunc, authMW gin.HandlerFunc) {
	ag := api.Group("")
	ag.Use(authMW)
	l, g, cr, u, d := crud[models.Terminal](h)
	ag.GET("/terminals", l)
	ag.GET("/terminals/:id", g)
	ag.POST("/terminals", admin, cr)
	ag.PUT("/terminals/:id", admin, u)
	ag.DELETE("/terminals/:id", admin, d)
	ag.GET("/cards", h.listCards)
	ag.GET("/cards/:id", h.getCard)
	ag.POST("/cards", h.createCard)
	ag.PUT("/cards/:id", h.updateCard)
	ag.DELETE("/cards/:id", h.deleteCard)
	ag.GET("/transactions", h.listTransactions)
	ag.GET("/transactions/:id", h.getTransaction)
	ag.POST("/transactions", h.createTransaction)
	ag.PUT("/transactions/:id", h.updateTransaction)
	ag.DELETE("/transactions/:id", h.deleteTransaction)
	kg := ag.Group("/keys")
	kg.Use(admin)
	l, g, cr, u, d = crud[models.CardKey](h)
	kg.GET("", l)
	kg.GET("/:id", g)
	kg.POST("", cr)
	kg.PUT("/:id", u)
	kg.DELETE("/:id", d)
	ag.GET("/users", admin, func(c *gin.Context) { var users []models.User; h.DB.Find(&users); c.JSON(200, users) })
	ag.POST("/users", admin, h.CreateUser)
	ag.PUT("/users/:id", h.UpdateUser)
	ag.DELETE("/users/:id", admin, func(c *gin.Context) { var u models.User; h.DB.Delete(&u, id(c)); c.Status(204) })
	api.POST("/terminal/payments/authorize", h.AuthorizePayment)
	api.GET("/terminal/keys", h.DownloadKeys)
}

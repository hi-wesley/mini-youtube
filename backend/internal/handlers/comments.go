package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/firebase"
	"github.com/hi-wesley/mini-youtube/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Cloud Run manages TLS/Origins
}

// GET /v1/ws/comments?vid=<videoID>   (Upgrades to WS)
func CommentsSocket(c *gin.Context) {
	log.Println("CommentsSocket: new connection")
	vid := c.Query("vid")
	if vid == "" {
		log.Println("CommentsSocket: missing vid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing vid"})
		return
	}

	tokenStr := c.Query("token")
	if tokenStr == "" {
		log.Println("CommentsSocket: missing token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	token, err := firebase.Client.VerifyIDToken(c, tokenStr)
	if err != nil {
		log.Printf("CommentsSocket: invalid token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Store UID in context for this connection if needed later
	c.Set("uid", token.UID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("CommentsSocket: upgrade error: %v", err)
		return
	}
	log.Println("CommentsSocket: connection upgraded")

	// Basic pub‑sub: use a channel per video in memory
	hub := getHub(vid) // see below
	hub.register <- conn
	log.Println("CommentsSocket: connection registered")

	// read pump: discard messages (front‑end sends via REST)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("CommentsSocket: read error: %v", err)
			hub.unregister <- conn
			log.Println("CommentsSocket: connection unregistered")
			break
		}
	}
}

type wsHub struct {
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan models.Comment
	clients    map[*websocket.Conn]struct{}
}

var hubs = make(map[string]*wsHub)

func getHub(videoID string) *wsHub {
	if h, ok := hubs[videoID]; ok {
		return h
	}
	h := &wsHub{
		register: make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast: make(chan models.Comment, 16),
		clients: make(map[*websocket.Conn]struct{}),
	}
	go h.run()
	hubs[videoID] = h
	return h
}

func (h *wsHub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
		case c := <-h.unregister:
			delete(h.clients, c); _ = c.Close()
		case msg := <-h.broadcast:
			for cli := range h.clients {
				_ = cli.WriteJSON(msg)
			}
		}
	}
}

func GetComments(c *gin.Context) {
	var comments []models.Comment
	if err := db.Conn.Preload("User").Where("video_id = ?", c.Param("id")).Order("created_at asc").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusOK, comments)
}

// POST /v1/comments  {video_id, message}
func CreateComment(c *gin.Context) {
	uid := c.GetString("uid")
	var req struct {
		VideoID string `json:"video_id" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}
	comment := models.Comment{UserID: uid, VideoID: req.VideoID, Message: req.Message}
	if err := db.Conn.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"}); return
	}

	// Eager load user before broadcasting
	db.Conn.Preload("User").First(&comment, comment.ID)

	getHub(req.VideoID).broadcast <- comment
	c.JSON(http.StatusCreated, comment)
}

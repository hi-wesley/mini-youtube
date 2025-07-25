package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Cloud Run manages TLS/Origins
}

// GET /v1/ws/comments?vid=<videoID>   (Upgrades to WS)
func CommentsSocket(c *gin.Context) {
	vid := c.Query("vid")
	if vid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing vid"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil { return }

	// Basic pub‑sub: use a channel per video in memory
	hub := getHub(vid) // see below
	hub.register <- conn

	// read pump: discard messages (front‑end sends via REST)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			hub.unregister <- conn
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
	if err := db.Conn.Where("video_id = ?", c.Param("id")).Order("created_at desc").Find(&comments).Error; err != nil {
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
	getHub(req.VideoID).broadcast <- comment
	c.JSON(http.StatusCreated, comment)
}

func UploadVideo(c *gin.Context) {
	uid := c.GetString("uid")
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video required"}); return
	}
	// basic client‑side validated: size, mime type, etc.

	object := fmt.Sprintf("videos/%s/%d-%s", uid, time.Now().Unix(), header.Filename)
	writer := storageClient.Bucket(cfg.GcsBucket).Object(object).NewWriter(c)
	if _, err := io.Copy(writer, file); err != nil { … }
	_ = writer.Close()

	vid := models.Video{
		ID: uuid.NewString(), UserID: uid, Title: c.PostForm("title"),
		Description: c.PostForm("description"), ObjectName: object,
	}
	db.Create(&vid)

	// Async summarization (fire‑and‑forget)
	go ai.GenerateAndCacheSummary(vid.ID, "gs://"+cfg.GcsBucket+"/"+object)

	c.JSON(http.StatusCreated, vid)
}

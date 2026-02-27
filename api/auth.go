package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"surface-api/models"
	"time"

	"github.com/astaxie/beego/session"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// login authenticates a user and creates a session
func (app *App) login(c *gin.Context) {
	var req = &models.Login{}

	if err := c.BindJSON(req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256([]byte(req.Password))
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(dst, hash[:])

	if err := app.db.First(req, "username=?", req.Username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"error": "Invalid username/password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Password != string(dst) {
		c.JSON(http.StatusOK, gin.H{
			"error": "Invalid username/password",
		})
		return
	}

	s, _ := c.Get("sess")
	sess := s.(session.Store)
	sess.Set("username", req.Username)

	c.JSON(http.StatusOK, gin.H{
		"username": req.Username,
	})
}

// logout destroys the current session
func (app *App) logout(c *gin.Context) {
	app.sess.SessionDestroy(c.Writer, c.Request)
	c.Status(http.StatusOK)
}

// checkSession returns current session information
func (app *App) checkSession(c *gin.Context) {
	s, _ := c.Get("sess")
	sess := s.(session.Store)
	username := sess.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"username": username,
	})
}

// AuthMiddleware verifies session for protected routes
func (app *App) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s, err := app.sess.SessionStart(c.Writer, c.Request)
		if err != nil {
			log.Println("session error", err)
		}
		defer s.SessionRelease(c.Writer)

		url := c.Request.URL.String()
		if url != "/swagger/" && url != "/login" && url != "/logout" {
			if s.Get("username") == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Session expired",
				})
				return
			}
		}
		c.Set("sess", s)
		c.Next()
	}
}

// listUsers returns list of all users (passwords masked)
func (app *App) listUsers(c *gin.Context) {
	var users []models.Login

	if err := app.db.Find(&users).Error; err != nil {
		sendError(c, err)
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}

// addUser creates a new user
func (app *App) addUser(c *gin.Context) {
	var input models.CreateUserInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash := sha256.Sum256([]byte(input.Password))
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(dst, hash[:])

	user := models.Login{
		Username: input.Username,
		Password: string(dst),
	}

	if err := app.db.Create(&user).Error; err != nil {
		sendError(c, err)
		return
	}

	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

// deleteUser deletes a user and invalidates their sessions
func (app *App) deleteUser(c *gin.Context) {
	username := c.Param("username")

	// Get current logged in user
	s, _ := c.Get("sess")
	sess := s.(session.Store)
	currentUser := sess.Get("username")

	// Prevent self-deletion
	if currentUser != nil && currentUser.(string) == username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete currently logged in user"})
		return
	}

	var user models.Login

	if err := app.db.First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		sendError(c, err)
		return
	}

	usernameToDelete := user.Username

	if err := app.db.Delete(&user).Error; err != nil {
		sendError(c, err)
		return
	}

	go app.invalidateUserSessions(usernameToDelete)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// changePassword allows a logged-in user to change their own password
func (app *App) changePassword(c *gin.Context) {
	username := c.Param("username")

	s, _ := c.Get("sess")
	sess := s.(session.Store)
	currentUser := sess.Get("username")

	// Only allow changing own password
	if currentUser == nil || currentUser.(string) != username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only change own password"})
		return
	}

	// rate limiting: block if too many recent failed attempts
	now := time.Now()
	app.pwdChangeLock.Lock()
	attempts := app.pwdChangeAttempts[username]
	var pruned []time.Time
	for _, t := range attempts {
		if now.Sub(t) < pwdChangeWindow {
			pruned = append(pruned, t)
		}
	}
	if len(pruned) >= pwdChangeMaxAttempts {
		// compute remaining cooldown based on the oldest attempt in window
		earliest := pruned[0]
		remaining := pwdChangeWindow - now.Sub(earliest)
		if remaining < 0 {
			remaining = 0
		}
		app.pwdChangeLock.Unlock()
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many password change attempts, try again later", "attempts_left": 0, "cooldown_seconds": int(remaining.Seconds())})
		return
	}
	app.pwdChangeAttempts[username] = pruned
	app.pwdChangeLock.Unlock()

	var input struct {
		CurrentPassword string `json:"current_password"`
		Password        string `json:"password"`
		Confirm         string `json:"confirm"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// basic validations
	if input.Password == "" || input.Confirm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and confirm are required"})
		return
	}
	if input.Password != input.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and confirm do not match"})
		return
	}
	if len(input.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	// verify current password
	var user models.Login
	if err := app.db.First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
		sendError(c, err)
		return
	}

	currHash := sha256.Sum256([]byte(input.CurrentPassword))
	dstCurr := make([]byte, base64.StdEncoding.EncodedLen(len(currHash)))
	base64.StdEncoding.Encode(dstCurr, currHash[:])

	if user.Password != string(dstCurr) {
		// record failed attempt
		app.pwdChangeLock.Lock()
		app.pwdChangeAttempts[username] = append(app.pwdChangeAttempts[username], now)
		// compute attempts left and cooldown
		attemptsLeft := pwdChangeMaxAttempts - len(app.pwdChangeAttempts[username])
		if attemptsLeft < 0 {
			attemptsLeft = 0
		}

		var cooldownSeconds int
		if attemptsLeft == 0 {
			earliest := app.pwdChangeAttempts[username][0]
			remaining := pwdChangeWindow - now.Sub(earliest)
			if remaining < 0 {
				remaining = 0
			}

			cooldownSeconds = int(remaining.Seconds())
		} else {
			cooldownSeconds = 0
		}
		app.pwdChangeLock.Unlock()

		c.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect", "attempts_left": attemptsLeft, "cooldown_seconds": cooldownSeconds})
		return
	}

	// set new password
	newHash := sha256.Sum256([]byte(input.Password))
	dstNew := make([]byte, base64.StdEncoding.EncodedLen(len(newHash)))
	base64.StdEncoding.Encode(dstNew, newHash[:])

	if err := app.db.Model(&models.Login{}).Where("username = ?", username).Update("password", string(dstNew)).Error; err != nil {
		sendError(c, err)
		return
	}

	// invalidate other sessions in background and ensure current session remains set
	go app.invalidateUserSessions(username)
	// clear failed attempts on success
	app.pwdChangeLock.Lock()
	delete(app.pwdChangeAttempts, username)
	app.pwdChangeLock.Unlock()

	// UI should force re-login sess.Set("username", username)

	c.Status(http.StatusOK)
}

// invalidateUserSessions deletes all sessions for a given user
func (app *App) invalidateUserSessions(username string) {
	type sessionRecord struct {
		SessionKey  string `gorm:"column:session_key"`
		SessionData []byte `gorm:"column:session_data"`
	}

	var sessions []sessionRecord
	if err := app.db.Table("session").Find(&sessions).Error; err != nil {
		log.Println("Warning: Failed to fetch sessions:", err)
		return
	}

	var keysToDelete []string
	for _, session := range sessions {
		sessionDataStr := string(session.SessionData)
		if len(sessionDataStr) > 0 && (string(session.SessionData)[0] == 0x0D || session.SessionData[0] == 0x00) {
			if bytes.Contains(session.SessionData, []byte(username)) {
				keysToDelete = append(keysToDelete, session.SessionKey)
			}
		}
	}

	if len(keysToDelete) > 0 {
		if err := app.db.Table("session").Where("session_key IN ?", keysToDelete).Delete(nil).Error; err != nil {
			log.Println("Warning: Failed to delete sessions:", err)
		} else {
			log.Printf("Invalidated %d session(s) for user: %s", len(keysToDelete), username)
		}
	}
}

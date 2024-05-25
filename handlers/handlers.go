package handlers

import (
	. "clouderrors"
	. "db"
	"fmt"
	. "models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"utils"

	"github.com/gin-gonic/gin"
)

// Auth
func Register(c *gin.Context) {
	var data LoginData

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ErrInvalidData.Error())
		return
	}

	if !utils.ValidateEmail(data.Email) {
		c.JSON(http.StatusBadRequest, ErrInvalidEmail.Error())
		return
	}
	passwordErr := utils.ValidatePassword(data.Password)

	if passwordErr != nil {
		c.JSON(http.StatusBadRequest, passwordErr.Error())
		return
	}

	if DB.CheckIfUserWithEmailExists(data.Email) {
		c.JSON(http.StatusBadRequest, ErrEmailIsAlreadyTaken.Error())
		return
	}

	hash, err := utils.HashPassword(data.Password)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrRegistration.Error())
		return
	}

	create_err := DB.CreateUser(data.Email, hash)

	if create_err != nil {
		c.JSON(http.StatusInternalServerError, ErrRegistration.Error())
		return
	}

	user, err := DB.GetUser(data.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrRegistration.Error())
		return
	}

	token, err := utils.GenerateToken(user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrRegistration.Error())
		return
	}

	err = DB.SaveToken(user.ID, token)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrRegistration.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "uid": user.ID})
}

func Login(c *gin.Context) {
	var data LoginData

	// Check user credentials and generate a JWT token
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ErrInvalidData.Error())
		return
	}

	user, err := DB.GetUser(data.Email)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrInvalidCredentials.Error())
		return
	}

	if !utils.CheckPasswordHash(data.Password, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, ErrInvalidCredentials.Error())
		return
	}

	token, err := utils.GenerateToken(user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrLogin.Error())
		return
	}

	err = DB.SaveToken(user.ID, token)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrLogin.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "uid": user.ID})
}

func Logout(c *gin.Context) {
	var uid = c.GetInt64("user_id")
	var token = c.GetString("token")
	err := DB.DeleteToken(uid, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
	}
	c.JSON(http.StatusOK, gin.H{})
}

// User functions

func AddObject(c *gin.Context) {
	uid := c.GetInt64("user_id")
	path := "./" + c.Query("path")
	path = filepath.Clean(path)
	file, err := c.FormFile("file")
	if err != nil {
		fmt.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	name := file.Filename

	owner := c.Param("owner_id")
	ownerID, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	exists := DB.CheckIfUserWithIDExists(ownerID)
	if !exists {
		c.Status(http.StatusBadRequest)
		return
	}

	if uid != ownerID {
		if !DB.CheckCanWrite(uid, ownerID, path, name) {
			c.Status(http.StatusBadRequest)
			return
		}
	}
	
	dst := filepath.Join("/Users/usmanturkaev/cloud/", owner, path, name)

	_, err = DB.GetResource(uid, name, path)
	if err == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	path = filepath.Clean(path)
	err = c.SaveUploadedFile(file, dst)

	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	parts := strings.Split(path, string(os.PathSeparator))
	currentPath := filepath.Clean("")
	for i := 0; i < len(parts); i++ {
		if !(parts[i] == "." && currentPath == ".") {
			if !DB.CheckResourceExists(ownerID, parts[i], currentPath) {
				_ = DB.CreateResource(ownerID, parts[i], currentPath, "dir")
			}
		}
		currentPath = filepath.Join(currentPath, parts[i])
	}

	if err != nil {
		fmt.Print(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	err = DB.CreateResource(ownerID, file.Filename, path, "file")
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func GetResource(c *gin.Context) {
	uid := c.GetInt64("user_id")
	path := "./" + c.Query("path")
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	owner := c.Param("owner_id")
	ownerID, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	exists := DB.CheckIfUserWithIDExists(ownerID)
	if !exists {
		c.Status(http.StatusBadRequest)
		return
	}

	if uid != ownerID {
		if !DB.CheckCanRead(uid, ownerID, dir, name) {
			c.Status(http.StatusBadRequest)
			return
		}
	}

	if path == "." {
		content, err := DB.GetDirContent(ownerID, dir, name, false, false)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, content)
		return
	}
	resource, err := DB.GetResource(ownerID, name, dir)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if resource.Type == "dir" {
		content, err := DB.GetDirContent(ownerID, dir, name, false, false)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, content)
	} else {
		dst := filepath.Join("/Users/usmanturkaev/cloud/", owner, resource.Path, resource.Name)
		c.FileAttachment(dst, resource.Name)
		c.Status(http.StatusOK)
	}
}

func DeleteResource(c *gin.Context) {
	uid := c.GetInt64("user_id")
	path := "./" + c.Query("path")
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	owner := c.Param("owner_id")
	ownerID, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	exists := DB.CheckIfUserWithIDExists(ownerID)
	if !exists {
		c.Status(http.StatusBadRequest)
		return
	}

	if uid != ownerID {
		if !DB.CheckCanWrite(uid, ownerID, dir, name) {
			c.Status(http.StatusBadRequest)
			return
		}
	}

	if dir == "." && name == "." {
		c.Status(http.StatusForbidden)
		return
	}

	err = DB.DeleteResource(ownerID, dir, name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dst := filepath.Join("/Users/usmanturkaev/cloud/", owner, dir, name)
	err = utils.DeleteResource(dst)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func CreateDirectory(c *gin.Context) {
	uid := c.GetInt64("user_id")
	path := "./" + c.Query("path")
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	owner := c.Param("owner_id")
	ownerID, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		fmt.Println("parse")
		c.Status(http.StatusBadRequest)
		return
	}
	exists := DB.CheckIfUserWithIDExists(ownerID)
	if !exists {
		fmt.Println("exists")
		c.Status(http.StatusBadRequest)
		return
	}

	if uid != ownerID {
		if !DB.CheckCanWrite(uid, ownerID, dir, name) {
			fmt.Println("write")
			c.Status(http.StatusBadRequest)
			return
		}
	}

	_, err = DB.GetResource(ownerID, name, dir)
	if err == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	dst := filepath.Join("/Users/usmanturkaev/cloud/", owner, dir, name)
	utils.CreateDir(dst)

	parts := strings.Split(dir, string(os.PathSeparator))
	currentPath := filepath.Clean("")
	for i := 0; i < len(parts); i++ {
		if !(parts[i] == "." && currentPath == ".") {
			if !DB.CheckResourceExists(ownerID, parts[i], currentPath) {
				_ = DB.CreateResource(ownerID, parts[i], currentPath, "dir")
			}
		}
		currentPath = filepath.Join(currentPath, parts[i])
	}

	err = DB.CreateResource(ownerID, name, dir, "dir")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func ShareRights(c *gin.Context) {
	uid := c.GetInt64("user_id")
	email := c.Query("email")
	write := strings.ToLower(c.Query("write")) == "true"
	path := "./" + c.Query("path")
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	user, err := DB.GetUser(email)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if user.ID == uid {
		c.Status(http.StatusBadRequest)
		return
	}
	if DB.CheckCanRead(user.ID, uid, dir, name) {
		c.Status(http.StatusBadRequest)
		return
	}
	success := DB.ShareRights(user.ID, uid, dir, name, write)
	if success {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

func DeleteRights(c *gin.Context) {
	uid := c.GetInt64("user_id")
	email := c.Query("email")
	path := "./" + c.Query("path")
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	user, err := DB.GetUser(email)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if user.ID == uid {
		c.Status(http.StatusBadRequest)
		return
	}
	if !DB.CheckCanRead(user.ID, uid, dir, name) {
		c.Status(http.StatusBadRequest)
		return
	}
	success := DB.DeleteRights(user.ID, uid, dir, name)
	if success {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

func GetUsersSharedResources(c *gin.Context) {
	uid := c.GetInt64("user_id")

	content, err := DB.GetUserSharedResources(uid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, content)
}

func GetUsersWithAccess(c *gin.Context) {
	uid := c.GetInt64("user_id")
	path := "./" + c.Query("path")
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	owner := c.Param("owner_id")
	ownerID, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	exists := DB.CheckIfUserWithIDExists(ownerID)
	if !exists {
		c.Status(http.StatusBadRequest)
		return
	}

	if uid != ownerID {
		c.Status(http.StatusForbidden)
		return
	}

	users, err := DB.GetUsersWithAccess(ownerID, dir, name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, users)
}
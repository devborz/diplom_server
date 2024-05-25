package models

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type LoginData struct {
	Email    string
	Password string
}

type DirContent struct {
	Resources []Resource
}

type Resource struct {
	ID      int64
	Path    string
	Name    string
	OwnerID int64
	Created string
	Type    string
}

type SharedResource struct {
	Resource Resource
	ResourceID int64
	UserID int64
	CanWrite bool
	Created string
}

type UserWithAccessList struct {
	Users []UserWithAccess
}

type UserWithAccess struct {
	Email string
	Write bool
}
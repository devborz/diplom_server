module diplom_server

require github.com/gin-gonic/gin v1.10.0

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	golang.org/x/crypto v0.23.0 // indirect
)

replace github.com/devborz/diplom_server/server => ./github.com/devborz/diplom_server/server

replace github.com/devborz/diplom_server/db => ./github.com/devborz/diplom_server/db

go 1.21.3

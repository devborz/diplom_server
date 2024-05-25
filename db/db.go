package db

import (
	"database/sql"
	"fmt"
	"models"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "FkSo2W01dp"
	dbname   = "postgres"
)

type Database struct {
	db *sql.DB
}

func (d *Database) Connect() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	d.db = db
}

func (d *Database) CreateUser(email string, hash string) error {
	_, err := d.db.Exec("insert into users (email, password_hash) values($1, $2)", email, hash)
	return err
}

func (d *Database) GetUser(email string) (models.User, error) {
	user := models.User{}
	err := d.db.QueryRow("select ID, email, password_hash from users where email=$1", email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	return user, err
}

func (d *Database) GetUserByID(id int64) (models.User, error) {
	user := models.User{}
	err := d.db.QueryRow("select ID, email, password_hash from users where id=$1", id).Scan(&user.ID, &user.Email, &user.PasswordHash)
	return user, err
}

func (d *Database) CheckIfUserWithEmailExists(email string) bool {
	var count int
	err := d.db.QueryRow("select count(*) from users where email=$1", email).Scan(&count)
	if err != nil {
		return false
	}
	return count != 0
}

func (d *Database) CheckIfUserWithIDExists(id int64) bool {
	var count int
	err := d.db.QueryRow("select count(*) from users where id=$1", id).Scan(&count)
	if err != nil {
		return false
	}
	return count != 0
}

func (d *Database) CreateResource(uid int64, name string, path string, type_ string) error {
	_, err := d.db.Exec("insert into resources (name, path, owner_id, type) values($1, $2, $3, $4)", name, path, uid, type_)
	return err
}

func (d *Database) CheckResourceExists(uid int64, name string, path string) bool {
	var count int
	err := d.db.QueryRow("select count(*) from resources where name=$1 and path=$2 and owner_id=$3", name, path, uid).Scan(&count)
	if err != nil {
		return false
	}
	return count != 0
}

func (d *Database) GetResource(uid int64, name string, path string) (models.Resource, error) {
	r := models.Resource{}
	var time time.Time
	queryStr := "select ID, name, path, type, created from resources where owner_id=$1 and path=$2 and name=$3"
	err := d.db.QueryRow(queryStr, uid, path, name).Scan(&r.ID, &r.Name, &r.Path, &r.Type, &time)
	r.Created = time.String()
	r.OwnerID = uid
	return r, err
}

func (d *Database) DeleteResource(uid int64, path string, name string) error {
	path = filepath.Clean(path)
	fullpath := filepath.Join(path, name) + "/%"
	_, err := d.db.Exec("delete from resources where (path=$1 and name=$2) or (path like $3) and owner_id=$4", path, name, fullpath, uid)
	return err
}

func (d *Database) GetDirContent(uid int64, path string, name string, sortByName bool, ascending bool) (models.DirContent, error) {
	fullpath := filepath.Join(path, name)
	order := ""
	if sortByName {
		order = "name"
	} else {
		order = "created"
	}
	asc := "DESC"
	if ascending {
		asc = "ASC"
	}
	queryString := fmt.Sprintf("select id, name, path, owner_id, created, type from resources where path=$1 and owner_id=$2 order by %s %s", order, asc)
	rows, err := d.db.Query(queryString, fullpath, uid)
	if err != nil {
		fmt.Println(err.Error())
		return models.DirContent{}, err
	}
	defer rows.Close()

	resources := []models.Resource{}

	for rows.Next() {
		r := models.Resource{}
		var time time.Time
		err := rows.Scan(&r.ID, &r.Name, &r.Path, &r.OwnerID, &time, &r.Type)
		if err != nil {
			fmt.Println(err)
			continue
		}
		r.Created = time.String()
		resources = append(resources, r)
	}
	return models.DirContent{Resources: resources}, nil
}

func (d *Database) SaveToken(uid int64, token string) error {
	_, err := d.db.Exec("insert into authtokens (user_id, token) values($1, $2)", uid, token)
	return err
}

func (d *Database) CheckToken(uid int64, token string) bool {
	var count int
	err := d.db.QueryRow("select count(*) from authtokens where user_id=$1 and token=$2", uid, token).Scan(&count)
	if err != nil {
		return false
	}

	return count != 0
}

func (d *Database) DeleteToken(uid int64, token string) error {
	_, err := d.db.Exec("delete from authtokens where user_id=$1 and token=$2", uid, token)
	return err
}

func (d *Database) CheckCanRead(uid int64, ownerID int64, path string, name string) bool {
	var fullpath = filepath.Join(path, name)
	var count int
	err := d.db.QueryRow("select count(*) from sharedresources where owner_id=$1 and user_id=$2 and $3 like CONCAT(fullpath, '%')", ownerID, uid, fullpath).Scan(&count)
	if err != nil {
		return false
	}
	return count != 0
}

func (d *Database) CheckCanWrite(uid int64, ownerID int64, path string, name string) bool {
	var fullpath = filepath.Join(path, name)
	var count int
	fmt.Println(fullpath)
	err := d.db.QueryRow("select count(*) from sharedresources where owner_id=$1 and user_id=$2 and $3 like CONCAT(fullpath, '%') and can_write=$4", ownerID, uid, fullpath, true).Scan(&count)
	if err != nil {
		return false
	}
	return count != 0
}

func (d *Database) ShareRights(uid int64, ownerID int64, path string, name string, write bool) bool {
	var dirID string
	err := d.db.QueryRow("select ID from resources where path=$1 and name=$2 and owner_id=$3 and type=$4", path, name, ownerID, "dir").Scan(&dirID)
	if dirID == "" || err != nil {
		return false
	}
	fullpath := filepath.Join(path, name)
	_, err = d.db.Exec("insert into sharedresources (owner_id, fullpath, path, name, user_id, can_write) values($1, $2, $3, $4, $5, $6)", ownerID, fullpath, path, name, uid, write)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (d *Database) DeleteRights(uid int64, ownerID int64, path string, name string) bool {
	fullpath := filepath.Join(path, name)
	_, err := d.db.Exec("delete from sharedresources where owner_id=$1 and $2 like CONCAT(fullpath, '%') and user_id=$3", ownerID, fullpath, uid)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (d *Database) GetUsersWithAccess(ownerID int64, path string, name string) (models.UserWithAccessList, error) {
	users := []models.UserWithAccess{}
	fullpath := filepath.Join(path, name)
	rows, err := d.db.Query("select email, can_write from sharedresources inner join users on sharedresources.user_id=users.id where owner_id=$1 and $2 like CONCAT(fullpath, '%')", ownerID, fullpath)
	if err != nil {
		return models.UserWithAccessList{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		var write bool
		err := rows.Scan(&email, &write)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		users = append(users, models.UserWithAccess{Email: email, Write: write})
	}
	return models.UserWithAccessList{Users: users}, nil
}

func (d *Database) GetUserSharedResources(uid int64) (models.DirContent, error) {
	resources := []models.Resource{}
	rows, err := d.db.Query("select resources.id, resources.owner_id, resources.path, resources.name, resources.created, resources.type from sharedresources inner join resources on sharedresources.owner_id=resources.owner_id and sharedresources.path=resources.path and sharedresources.name=resources.name where user_id=$1", uid)
	if err != nil {
		fmt.Println(err.Error())
		return models.DirContent{Resources: resources}, nil
	}
	defer rows.Close()

	for rows.Next() {
		r := models.Resource{}
		var time time.Time
		err := rows.Scan(&r.ID, &r.OwnerID, &r.Path, &r.Name, &time, &r.Type)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		r.Created = time.String()
		resources = append(resources, r)
	}
	return models.DirContent{Resources: resources}, nil
}

var DB Database = Database{}

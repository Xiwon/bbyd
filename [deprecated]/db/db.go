package db

import(
	"os"
	"fmt"
	"crypto/sha256"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserProfile struct {
	Uid uint
	Username string
	Email string
	Auth string
}
type userDetails struct {
	gorm.Model
	Username string
	Secret string
	Email string
	Auth string
}

var dsn = "host=localhost port=5432 user=postgres dbname=byddb sslmode=disable"
var dbase, start_err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

func DatabaseStart() {
	if start_err != nil {
		fmt.Println("bbyd: Database connect error")
		fmt.Println(start_err)
		os.Exit(1)
	} else {
		fmt.Println("bbyd: Database connect succeed")
	}

	dbase.AutoMigrate(&userDetails{})

	TryChangeAuth("root", "admin")

	var usrDatas []userDetails
	dbase.Find(&usrDatas)
	fmt.Println("database status:\n", usrDatas)
}

func GetUsrByName(name string) (UserProfile, int) {
	var user userDetails
	result := dbase.First(&user, "username = ?", name)
	if result.RowsAffected > 0 {
		prof := UserProfile{
			Uid: user.ID,
			Username: user.Username,
			Email: user.Email,
			Auth: user.Auth,
		}
		return prof, 1
	} else {
		return UserProfile{}, 0
	}
}

func GetSecretByName(name string) (string, int) {
	var user userDetails
	result := dbase.First(&user, "username = ?", name)
	if result.RowsAffected > 0 {
		return user.Secret, 1
	} else {
		return "", 0
	}
}

func TryRegister(name string, passwd string, email string) int {
	var user userDetails
	result := dbase.First(&user, "username = ?", name)
	if result.RowsAffected == 0 {
		sec_sum := sha256.Sum256([]byte(passwd))
		sec := fmt.Sprintf("%x", sec_sum)
		var newusr = userDetails{
			Username: name,
			Secret: sec,
			Email: email,
			Auth: "normal",
		}
		dbase.Create(&newusr)
		return 1
	} else {
		return 0
	}
}

func TryChangeAuth(name string, auth string) int {
	var user userDetails
	result := dbase.First(&user, "username = ?", name)

	if result.RowsAffected > 0 {
		user.Auth = auth
		dbase.Save(user)
		return 1
	} else {
		return 0
	}
}

func TryChangeInfo(name string, passwd string, email string) (string, int) {
	var user userDetails
	result := dbase.First(&user, "username = ?", name)

	if result.RowsAffected == 0 {
		return "user " + name + " not found", 0
	} else {
		var retmsg string
		
		if passwd != "" {
			sec_sum := sha256.Sum256([]byte(passwd))
			sec := fmt.Sprintf("%x", sec_sum)
			user.Secret = sec
			
			retmsg += "passwd changed"
		}
		if email != "" {
			user.Email = email

			if retmsg == "" {
				retmsg = "email changed to " + email
			} else {
				retmsg += ", email changed to " + email
			}
		}

		if retmsg == "" {
			retmsg = "nothing changed"
		} else {
			dbase.Save(&user)
		}

		return retmsg, 1
	}
}

func TryDeleteByName(name string) (string, int) {
	var user userDetails
	result := dbase.First(&user, "username = ?", name)

	if result.RowsAffected == 0 {
		return "user " + name + " not found", 0
	} else {
		dbase.Delete(&user)
		return "delete user " + name, 1
	}
}
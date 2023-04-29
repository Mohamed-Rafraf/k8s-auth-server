package pkg

import (
	"log"
	"os"

	"github.com/caarlos0/env/v8"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type config struct {
	Home          string `env:"HOME"`
	Port          string `env:"PORT" envDefault:"8080"`
	AdminName     string `env:"K8S_AUTH_ADMIN_NAME" envDefault:"Mohamed"`
	AdminFullName string `env:"K8S_AUTH_ADMIN_FULLNAME"envDefault:"Mohamed Rafraf"`
	AdminMail     string `env:"K8S_AUTH_ADMIN_MAIL"envDefault:"mohamedrafraf99@gmail.com"`
	ClientID      string `env:"OAUTH2_CLIENT_ID",required`
	ClientSecret  string `env:"OAUTH2_CLIENT_SECRET",required`
	RedirectURL   string `env:"OAUTH2_REDIRECT_URL",required`
}

var GoogleOauthConfig *oauth2.Config
var Config config
var DB *gorm.DB

func init() {
	_ = os.Mkdir("./clusters", 0775)
	err := InitConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	InitDB()
	if !AdminExists(Config.AdminMail) {
		err = AddAdmin(Config.AdminName, Config.AdminFullName, Config.AdminMail)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  Config.RedirectURL,
		ClientID:     Config.ClientID,
		ClientSecret: Config.ClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	log.Println(Config)
}

func InitConfig() error {
	err := env.Parse(&Config)
	if err != nil {
		return err
	}
	return nil
}

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	DB.AutoMigrate(&Admin{})
	DB.AutoMigrate(&Cluster{})
	DB.AutoMigrate(&ClusterUser{})
	DB.AutoMigrate(&UserCluster{})

	log.Println("Database initialized")
}

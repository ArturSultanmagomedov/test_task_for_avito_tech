package main

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"
)

func main() {

	// Trying to load all configs
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env variables: %s", err.Error())
	}

	address := fmt.Sprintf("%s:%s", viper.GetString("host"), viper.GetString("port"))

	// Update currencies quotes every 6 hours
	go func() {
		UpdateCurrencyJson()
		for range time.Tick(time.Hour * 6) {
			UpdateCurrencyJson()
		}
	}()

	// Initialize database
	postgres, err := NewPostgresDB(Config{
		Username: viper.GetString("db.username"),
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
		return
	}
	db := NewDatabase(postgres)
	if err := db.CreateUsersTable(); err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
		return
	}

	// Initialize Echo
	e := echo.New()

	// Set custom json validator
	e.Validator = &Validator{v: validator.New()}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/add_funds", func(c echo.Context) error {
		s := &struct {
			UserId int     `json:"id" validate:"required"`
			Sum    float32 `json:"sum" validate:"required"`
		}{}
		if err := c.Bind(s); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		if err = c.Validate(s); err != nil {
			return err
		}
		if s.Sum <= 0 {
			return c.JSON(http.StatusBadRequest, "sum can't be negative or 0")
		}

		ex, err := db.IsUserExist(s.UserId)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		if !ex {
			err := db.CreateUser(s.UserId, s.Sum)
			if err != nil {
				log.Fatalf(err.Error())
				return err
			}
		} else {
			_, err := db.UpdateBalance(s.UserId, s.Sum)
			if err != nil {
				log.Fatalf(err.Error())
				return err
			}
		}

		return c.NoContent(http.StatusOK)
	})
	e.GET("/write_off_funds", func(c echo.Context) error {
		s := &struct {
			UserId int     `json:"id" validate:"required"`
			Sum    float32 `json:"sum" validate:"required"`
		}{}
		if err := c.Bind(s); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		if err = c.Validate(s); err != nil {
			return err
		}
		if s.Sum <= 0 {
			return c.JSON(http.StatusBadRequest, "sum can't be negative or 0")
		}

		ex, err := db.IsUserExist(s.UserId)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		if !ex {
			s := fmt.Sprintf("user %d does not exist", s.UserId)
			log.Printf(s)
			return c.JSON(http.StatusNotFound, s)
		}

		user, err := db.GetUser(s.UserId)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		if user.Balance < s.Sum {
			e := fmt.Sprintf("user %d has insufficient funds", s.UserId)
			log.Printf(e)
			return c.JSON(http.StatusPreconditionFailed, e)
		}

		if _, err := db.UpdateBalance(s.UserId, user.Balance-s.Sum); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	})
	e.GET("/funds_transfer", func(c echo.Context) error {
		s := &struct {
			Id1 int     `json:"id1" validate:"required"`
			Id2 int     `json:"id2" validate:"required"`
			Sum float32 `json:"sum" validate:"required"`
		}{}
		if err := c.Bind(s); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		if err = c.Validate(s); err != nil {
			return err
		}
		if s.Sum <= 0 {
			return c.JSON(http.StatusBadRequest, "sum can't be negative or 0")
		}

		// Проверить существует ли первый юзер (если не существует - вернуть ошибку)
		ex, err := db.IsUserExist(s.Id1)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		if !ex {
			s := fmt.Sprintf("user %d does not exist", s.Id1)
			log.Printf(s)
			return c.JSON(http.StatusNotFound, s)
		}

		// Проверить достаточно ли средств у первого юзера (если нет - вернуть ошибку)
		user, err := db.GetUser(s.Id1)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		if user.Balance < s.Sum {
			e := fmt.Sprintf("user %d has insufficient funds", s.Id1)
			log.Printf(e)
			return c.JSON(http.StatusPreconditionFailed, e)
		}

		// Проверить существует ли второй юзер (если не существует - создать)
		ex, err = db.IsUserExist(s.Id2)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		if !ex {
			err := db.CreateUser(s.Id2, 0)
			if err != nil {
				log.Fatalf(err.Error())
				return err
			}
		}

		err = db.CreateFundsTransaction(s.Id1, s.Id2, s.Sum)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}

		return c.NoContent(http.StatusOK)
	})
	e.GET("/get_balance", func(c echo.Context) error {
		s := &struct {
			UserId int `json:"id" validate:"required"`
		}{}
		if err := c.Bind(s); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		if err := c.Validate(s); err != nil {
			return err
		}

		currency := c.QueryParam("currency")

		ex, err := db.IsUserExist(s.UserId)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}
		if !ex {
			s := fmt.Sprintf("user %d does not exist", s.UserId)
			log.Printf(s)
			return c.JSON(http.StatusNotFound, s)
		}

		user, err := db.GetUser(s.UserId)
		if err != nil {
			log.Fatalf(err.Error())
			return err
		}

		if currency == "" {
			return c.JSON(http.StatusOK, user.Balance)
		} else {
			if tt, err := ConvertFromRubTo(currency, user.Balance); err != nil {
				return err
			} else {
				return c.JSON(http.StatusOK, tt)
			}
		}
	})

	// Start server
	e.Logger.Fatal(e.Start(address))
}

// Configs from ./configs/config.yaml
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

// Validator validates json requests
type Validator struct {
	v *validator.Validate
}

func (cv *Validator) Validate(i interface{}) error {
	if err := cv.v.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return nil
}

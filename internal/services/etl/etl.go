// etl implements Service for Extract Transfrom Load pipeline.
package etl

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	"github.com/gin-gonic/gin"
	gsdmysql "github.com/go-sql-driver/mysql"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/spf13/viper"
)

type DatasetFileMetadata struct {
	Directory string
	Type      string
	IsRead    bool
	CreatedAt string
	ReadAt    string
}

type Etl struct {
	services.Service
}

func NewEtl() (services.IService, error) {
	e := Etl{}
	l := log.New(os.Stdout, "Etl: ", log.LstdFlags)
	e.Logger = l
	e.Router = gin.Default()

	v, err := e.SetConfig()
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.ConfigReader = nil
		e.State = services.StateIdle
		return &e, err
	}
	e.ConfigReader = v

	var ci services.ConnInfo
	if err := e.ConfigReader.UnmarshalKey("ConnInfo", &ci); err != nil {
		l.Fatalln(err)
	}

	e.ConnInfo = &ci
	cfg := gsdmysql.NewConfig()
	cfg.User = ci.Username
	cfg.Passwd = ci.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", ci.Ip, ci.Port)
	cfg.DBName = ci.Name

	db, err := sql.Open(ci.Type, cfg.FormatDSN())
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.State = services.StateIdle
		return &e, err
	}
	e.DB = db
	return &e, nil
}

func (e *Etl) IsDataExtracted() ([]*DatasetFileMetadata, error) {
	rows, err := e.DB.Query(`SELECT directory, data_type, read, created_at, read_at FROM schema_reads`)
	if err != nil {
		e.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	dfms := []*DatasetFileMetadata{}
	for rows.Next() {
		dfm := &DatasetFileMetadata{}
		if err := rows.Scan(&dfm.Directory, &dfm.Type, &dfm.IsRead,
			&dfm.CreatedAt, &dfm.ReadAt); err != nil {
			e.Logger.Printf("Error occured: %v.\n", err)
			continue
		}
		if !dfm.IsRead {
			dfms = append(dfms, dfm)
		}
	}
	return dfms, nil
}

func (e *Etl) Extract(...func()) error { panic("Not implemented") }

func (e *Etl) Start() error {
	dfms, err := e.IsDataExtracted()
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.State = services.StateIdle
	}

	if dfms != nil {
		// Extract data here
		e.Logger.Printf("Extracting data for the first ever setup...")
		panic("Not implemented")
	}
	// v1 of api.
	{
		v1 := e.Router.Group("/v1")
		v1.POST("api/sync", func(ctx *gin.Context) { panic("Not implemented") })
		v1.POST("api/", func(ctx *gin.Context) { panic("Not implemented") })
	}

	go func() {
		if err := e.Router.Run(":9998"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("Router failed: %v\n", err)
		}
	}()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	sig := <-kill
	e.Logger.Printf("Gracefully shutting down the server: %v\n.", sig)
	e.State = services.StateDown
	e.DB.Close()
	return fmt.Errorf("server closed")
}

func (e *Etl) SetConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("EtlConfig")
	v.SetConfigType("toml")
	v.AddConfigPath(os.Getenv("ETL_CONFIG_DIR_PATH"))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

// ExposeConnInfo exposes configuration.
func (e *Etl) ExposeConnInfo() *services.ConnInfo {
	return e.ConnInfo
}

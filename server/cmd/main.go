package main

import (
	httpDownloadServer "HTTP-download-server/server"
	"HTTP-download-server/server/models"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

var GitCommit string
var BuildTime string

func main() {
	var addr string = carrot.GetEnv("ADDR")
	var logFile string = carrot.GetEnv("LOG_FILE")
	var runMigration bool
	var logerLevel string = carrot.GetEnv("LOG_LEVEL")
	var dbDriver string = carrot.GetEnv(carrot.ENV_DB_DRIVER)
	var dsn string = carrot.GetEnv(carrot.ENV_DSN)
	var traceSql bool = carrot.GetEnv("TRACE_SQL") != ""

	var superUserEmail string
	var superUserPassword string

	log.Default().SetFlags(log.LstdFlags | log.Lshortfile)

	if addr == "" {
		addr = ":8000"
	}

	flag.StringVar(&superUserEmail, "superuser", "", "Create an super user with email")
	flag.StringVar(&superUserPassword, "password", "", "Super user password")
	flag.StringVar(&addr, "addr", addr, "HTTP Serve address")
	flag.StringVar(&logFile, "log", logFile, "Log output file name, default is os.Stdout")
	flag.StringVar(&logerLevel, "level", logerLevel, "Log level debug|info|warn|error")
	flag.BoolVar(&runMigration, "m", false, "Run migration only")
	flag.StringVar(&dbDriver, "db", dbDriver, "DB Driver, sqlite|mysql")
	flag.StringVar(&dsn, "dsn", dsn, "DB DSN")
	flag.BoolVar(&traceSql, "tracesql", traceSql, "Trace sql execution")
	flag.Parse()

	var lw io.Writer = os.Stdout
	var err error
	if logFile != "" {
		lw, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("open %s fail, %v\n", logFile, err)
		} else {
			log.SetOutput(lw)
		}
	} else {
		logFile = "console"
	}

	if logerLevel != "" {
		ll := carrot.LevelInfo
		switch logerLevel {
		case "debug":
			ll = carrot.LevelDebug
		case "info":
			ll = carrot.LevelInfo
		case "warn":
			ll = carrot.LevelWarning
		case "error", "fatal", "panic":
			ll = carrot.LevelError
		default:
			ll = carrot.LevelInfo
		}
		carrot.SetLogLevel(ll)
	}

	fmt.Println("GitCommit   =", GitCommit)
	fmt.Println("BuildTime   =", BuildTime)
	fmt.Println("addr        =", addr)
	fmt.Println("logfile     =", logFile)
	fmt.Println("logerLevel  =", logerLevel)
	fmt.Println("DB Driver   =", dbDriver)
	fmt.Println("DSN         =", dsn)
	fmt.Println("traceSql    =", traceSql)
	fmt.Println("migration   =", runMigration)

	db, err := carrot.InitDatabase(lw, dbDriver, dsn)
	if err != nil {
		carrot.Warning("init database failed", err)
		return
	}

	if traceSql {
		db = db.Debug()
	}

	if err = carrot.InitMigrate(db); err != nil {
		panic(err)
	}

	if err = models.Migration(db); err != nil {
		panic(err)
	}

	if runMigration {
		log.Println("migration done")
		return
	}

	if superUserEmail != "" && superUserPassword != "" {
		u, err := carrot.GetUserByEmail(db, superUserEmail)
		if err == nil && u != nil {
			carrot.SetPassword(db, u, superUserPassword)
			log.Println("Update super with new password")
		} else {
			u, err = carrot.CreateUser(db, superUserEmail, superUserPassword)
			if err != nil {
				panic(err)
			}
		}
		u.IsStaff = true
		u.Activated = true
		u.Enabled = true
		u.IsSuperUser = true
		db.Save(u)
		log.Println("Create super user")
		return
	}

	r := gin.New()

	server := NewDownloadServer(db)

	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    lw,
		Formatter: GinLoggerWithUserIdFormat,
	}), gin.Recovery())

	if err = server.Prepare(r); err != nil {
		log.Panic("prepare http-download-server failed", err)
		return
	}

	carrot.Warning("http-download-server is running on", addr)
	r.Run(addr)
}

func GinLoggerWithUserIdFormat(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	var userid string = "-"
	if user, ok := param.Keys[carrot.UserField]; ok && user != nil {
		if u, ok := user.(*carrot.User); ok {
			userid = u.Email
		}
	}

	return fmt.Sprintf("[HTTP] %v | %s |%s %3d %s| %s | %s | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		userid,
		statusColor, param.StatusCode, resetColor,
		httpDownloadServer.FormatSizeHuman(float64(param.BodySize)),
		param.Latency.Round(time.Millisecond),
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

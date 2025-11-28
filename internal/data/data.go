package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"tenant-service/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewRedis,
	NewTenantRepo,
	NewQuotaRepo,
	NewProductRepo,
)

// Data ..
type Data struct {
	db    *gorm.DB
	redis *redis.Client
}

// GormWriter u81eau5b9au4e49 GORM u65e5u5fd7u5199u5165u5668
type GormWriter struct {
	helper *log.Helper
}

// Printf u5b9eu73b0 logger.Writer u63a5u53e3
func (w *GormWriter) Printf(format string, args ...interface{}) {
	w.helper.Infof(format, args...)
}

// NewDB creates a new database connection.
func NewDB(conf *conf.Data, l log.Logger) *gorm.DB {
	logHelper := log.NewHelper(l)

	// u521bu5efa GORM u65e5u5fd7u914du7f6e
	writer := &GormWriter{helper: logHelper}
	gormLogger := logger.New(
		writer,
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(conf.Database.Source), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // u4f7fu7528u5355u6570u8868u540d
		},
	})
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get db: %v", err)
	}

	sqlDB.SetMaxIdleConns(int(conf.Database.MaxIdleConns))
	sqlDB.SetMaxOpenConns(int(conf.Database.MaxOpenConns))
	sqlDB.SetConnMaxLifetime(conf.Database.ConnMaxLifetime.AsDuration())

	logHelper.Info("database connected")
	return db
}

// NewRedis creates a new redis client.
func NewRedis(conf *conf.Data, l log.Logger) *redis.Client {
	logHelper := log.NewHelper(l)

	client := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           int(conf.Redis.Db),
		DialTimeout:  conf.Redis.DialTimeout.AsDuration(),
		ReadTimeout:  conf.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: conf.Redis.WriteTimeout.AsDuration(),
		PoolSize:     int(conf.Redis.PoolSize),
		MinIdleConns: int(conf.Redis.MinIdleConns),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logHelper.Fatalf("failed to connect redis: %v", err)
	}

	logHelper.Info("redis connected")
	return client
}

// NewData .
func NewData(db *gorm.DB, redis *redis.Client, l log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(l)
	logHelper.Info("creating data resources")

	d := &Data{
		db:    db,
		redis: redis,
	}

	return d, func() {
		logHelper.Info("closing data resources")
		if redis != nil {
			if err := redis.Close(); err != nil {
				logHelper.Errorf("redis close error: %v", err)
			}
		}
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				logHelper.Errorf("db connection error: %v", err)
				return
			}
			if err := sqlDB.Close(); err != nil {
				logHelper.Errorf("db close error: %v", err)
			}
		}
	}, nil
}

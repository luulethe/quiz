package db

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	// Maximum connections to DB
	MaxDBConns = 50
	// Maximum lifetime of a connection
	MaxDBConnLifeTime = time.Minute * 4
)

// Config defines the database connection settings
// DBAddress: the server address of database connection, default to primary database
// Replica: <optional>, a list of replica server addresses
// DBUsername: database connection user, for both primary db and replica db
// DBPassword: database connection password, for both primary db and replica db
// DBName: <optional>, for sharding use only, a list of sharding database names under the DBAddress
type Config struct {
	DBAddress       string        `yaml:"address" xml:"address"`
	Replica         []string      `yaml:"replica" xml:"replica"`
	DBUsername      string        `yaml:"username" xml:"username"`
	DBPassword      string        `yaml:"password" xml:"password"`
	DBName          string        `yaml:"name" xml:"name"`
	MaxOpenConns    int           `yaml:"max_open_conns" xml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" xml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" xml:"conn_max_lifetime"`
	DriverName      string        `yaml:"-" xml:"-"`
}

func dbConnect(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&interpolateParams=true&timeout=10s&readTimeout=10s",
		config.DBUsername, config.DBPassword, config.DBAddress, config.DBName)

	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = MaxDBConns
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = MaxDBConns
	}
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = MaxDBConnLifeTime
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true},
	})
	if err != nil {
		return nil, err
	}

	// init connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return db, err

}

func dbConnectToReplica(config Config) ([]*gorm.DB, error) {
	replicas := []*gorm.DB{}
	for _, replica := range config.Replica {
		config.DBAddress = replica
		replica, err := dbConnect(config)
		if err != nil {
			return nil, err
		}
		replicas = append(replicas, replica)
	}
	return replicas, nil
}

func ReadConfigs(rd io.Reader) (conf []Config, err error) {
	decoder := xml.NewDecoder(rd)
	for {
		var config Config
		err = decoder.Decode(&config)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		conf = append(conf, config)
	}
	return
}

type Stater interface {
	MasterStats() *sql.DBStats
	SlaveStats() []*sql.DBStats
}
type dbStatsCollector struct {
	ds   Stater
	desc *prometheus.Desc
}

func NewDBStatsCollector(managerName string, ds Stater) prometheus.Collector {
	desc := prometheus.NewDesc(
		"database_connection",
		"database connection statistics",
		[]string{"replica", "attribute"},
		prometheus.Labels{"manager": managerName},
	)
	return &dbStatsCollector{ds: ds, desc: desc}
}

func (c *dbStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

// Collect returns the current state of all metrics of the collector.
func (c *dbStatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.ds.MasterStats()
	if stats != nil {
		dbStatsCollect(ch, c.desc, *stats, "master")
	}

	for i, stats := range c.ds.SlaveStats() {
		dbStatsCollect(ch, c.desc, *stats, fmt.Sprintf("slave_%d", i))
	}
}

func dbStatsCollect(ch chan<- prometheus.Metric, desc *prometheus.Desc, stats sql.DBStats, replica string) {
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.MaxOpenConnections), replica, "MaxOpenConnections")
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.OpenConnections), replica, "OpenConnections")
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.InUse), replica, "InUse")
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.Idle), replica, "Idle")
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.WaitCount), replica, "WaitCount")
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.WaitDuration), replica, "WaitDuration")
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.MaxIdleClosed), replica, "MaxIdleClosed")
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue,
		float64(stats.MaxLifetimeClosed), replica, "MaxLifetimeClosed")
}

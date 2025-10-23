package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/v1adis1av28/level3/shortener/internal/config"
	"github.com/v1adis1av28/level3/shortener/internal/handlers/analytic"

	"github.com/wb-go/wbf/dbpg"
)

type Storage struct {
	DB    *dbpg.DB
	Mutex *sync.Mutex
}

func New(dbConf *config.DBConfig) (*Storage, error) {
	dbOpt := &dbpg.Options{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
	}
	db, err := dbpg.New(dbConf.MasterDbUrl, dbConf.SlaverUrl, dbOpt)
	if err != nil {
		log.Fatal("error on creating new storage %v", err)
		os.Exit(1)
	}

	stmt, err := db.Master.Prepare(`
		CREATE TABLE IF NOT EXISTS URL(
			ID SERIAL PRIMARY KEY,
			URL VARCHAR(255) NOT NULL,
			ALIAS VARCHAR(128) UNIQUE
		);
`)
	if err != nil {
		log.Fatal("error on initializing url tables, err: %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	stmt, err = db.Master.Prepare(`
		CREATE TABLE IF NOT EXISTS ANALYTIC(
			ID SERIAL PRIMARY KEY,
			ALIAS VARCHAR(128),
			USER_AGENT VARCHAR(255),
			REQUEST_TIME TIMESTAMP,
			CONSTRAINT ALIAS_NAME FOREIGN KEY(ALIAS) REFERENCES URL(ALIAS)
		);
	`)

	if err != nil {
		log.Fatal("error on initializing analytic tables, err: %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	return &Storage{DB: db, Mutex: &sync.Mutex{}}, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	stmt, err := s.DB.Master.Prepare("Select U.URL from URL as U where alias = $1")
	if err != nil {
		return "", fmt.Errorf("error while getting url! error : %v", err)
	}
	var result string
	err = stmt.QueryRow(alias).Scan(&result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("no url using this alias")
		}

		return "", err
	}

	return result, nil
}

func (s *Storage) ShortenURL(url, alias string) error {
	_, err := s.DB.Master.Exec("INSERT INTO URL (url, alias) values($1,$2);", url, alias)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { //unique constraint error handler
				return fmt.Errorf("this alias already present")
			}
		}
		return fmt.Errorf("error on prepare statemen inserting short url, err %v", err)
	}
	return nil
}

func (s *Storage) UpdateStats(alias, UA string) error {
	_, err := s.DB.Master.Exec("INSERT INTO ANALYTIC (ALIAS,USER_AGENT,REQUEST_TIME) VALUES($1,$2,$3);", alias, UA, time.Now())
	if err != nil {
		return fmt.Errorf("error on updating analytic")
	}
	return nil
}

func (s *Storage) GetAnalytic(alias string) (*analytic.AnalyticResponse, error) {
	result := &analytic.AnalyticResponse{}
	//currentdayGroup ID SERIAL PRIMARY KEY,
	// ALIAS VARCHAR(128),
	// USER_AGENT VARCHAR(255),
	// REQUEST_TIME TIMESTAMP,
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	stmt, err := s.DB.Master.Prepare(`SELECT A.ALIAS, A.USER_AGENT, A.REQUEST_TIME FROM
	 ANALYTIC AS A WHERE A.ALIAS = $1 AND REQUEST_TIME >= NOW() - INTERVAL '1 day';`)
	if err != nil {
		return nil, fmt.Errorf("error while fetching analytic for alias: %s, error: %v", alias, err)
	}
	rows, err := stmt.Query(alias)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var unit analytic.AnalyticUnit
		if err := rows.Scan(&unit.Alias, &unit.UA, &unit.RequestTime); err != nil {
			return nil, fmt.Errorf("error while scanning currrent day rows error:%v", err)
		}
		result.CurrentDayGroup = append(result.CurrentDayGroup, unit)
	}

	// LastMonthGroup
	stmt, err = s.DB.Master.Prepare(`SELECT A.ALIAS, A.USER_AGENT, A.REQUEST_TIME FROM
	 ANALYTIC AS A WHERE A.ALIAS = $1 AND REQUEST_TIME >= NOW() - INTERVAL '1 month';`)
	if err != nil {
		return nil, fmt.Errorf("error while fetching analytic for alias: %s, error: %v", alias, err)
	}
	rows, err = stmt.Query(alias)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var unit analytic.AnalyticUnit
		if err := rows.Scan(&unit.Alias, &unit.UA, &unit.RequestTime); err != nil {
			return nil, fmt.Errorf("error while scanning last month rows error:%v", err)
		}
		result.LastMonthGroup = append(result.LastMonthGroup, unit)
	}
	//UAGroup
	stmt, err = s.DB.Master.Prepare(`SELECT A.USER_AGENT,A.ALIAS,COUNT(*) as REDIRECT_COUNT,
    STRING_AGG(TO_CHAR(A.REQUEST_TIME, 'YYYY-MM-DD HH24:MI:SS'), '|' ORDER BY A.REQUEST_TIME DESC) as REQUEST_TIMES
	FROM ANALYTIC AS A 	WHERE A.ALIAS = $1 GROUP BY A.USER_AGENT, A.ALIAS
	ORDER BY REDIRECT_COUNT DESC;`)
	if err != nil {
		return nil, fmt.Errorf("error while fetching analytic for alias: %s, error: %v", alias, err)
	}
	rows, err = stmt.Query(alias)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var unit analytic.UAAnaliticUnit
		var timesStr string

		if err := rows.Scan(&unit.UA, &unit.Alias, &unit.RedirectCount, &timesStr); err != nil {
			return nil, fmt.Errorf("error while scanning last month rows error:%v", err)
		}

		if timesStr != "" {
			timeStrs := strings.Split(timesStr, "|")
			for _, ts := range timeStrs {
				parsedTime, err := time.Parse("2006-01-02 15:04:05", ts)
				if err == nil {
					unit.RequestTimes = append(unit.RequestTimes, parsedTime)
				}
			}
		}

		result.UserAgentGroup = append(result.UserAgentGroup, unit)
	}
	//todo декомпозировать функции
	fmt.Println(result.UserAgentGroup)
	return result, nil
}

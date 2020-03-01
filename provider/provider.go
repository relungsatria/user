package provider

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func initDB(connectionString string) *sql.DB {
	dbConnection, err := sql.Open(`mysql`, connectionString)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = dbConnection.Ping()
	if err != nil {
		log.Println(err)
		return nil
	}
	return dbConnection
}

func initRedis(address string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: address,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println(pong)
	return client
}


func queryTemplate(ctx context.Context, db *sql.DB, queryParam string, argsParam ...interface{}) (rows *sql.Rows, err error) {
	if db == nil {
		err = errors.New("db not initiated")
		log.Print(err)
		return
	}
	statement, err := db.Prepare(queryParam)
	if err != nil {
		log.Print(err)
		return
	}
	defer statement.Close()

	rows, err = statement.QueryContext(ctx, argsParam...)
	if err != nil {
		log.Print(err)
	}
	return
}

func mutationTemplate(ctx context.Context, db *sql.DB, queryParam string, argsParam ...interface{}) (err error) {
	if db == nil {
		err = errors.New("db not initiated")
		log.Print(err)
		return
	}
	statement, err := db.Prepare(queryParam)
	if err != nil {
		log.Print(err)
		return
	}
	defer statement.Close()

	_, err = statement.ExecContext(ctx, argsParam...)
	if err != nil {
		log.Print(err)
	}
	return
}
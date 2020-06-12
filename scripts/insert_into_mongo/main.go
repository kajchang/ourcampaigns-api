package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertIntoMongo(dumpPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	cancel()
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %s", err)
	}

	db := client.Database("ourcampaigns")

	fileInfos, err := ioutil.ReadDir(dumpPath)
	if err != nil {
		log.Fatalf("failed to read files from dump folder: %s", err)
	}

	for _, fileInfo := range fileInfos {
		fmt.Println(fileInfo.Name())
		col := db.Collection(strings.Split(fileInfo.Name(), "-")[0])

		file, err := os.Open(path.Join(dumpPath, fileInfo.Name()))
		if err != nil {
			log.Fatalf("failed to open file %s: %s", file.Name(), err)
		}

		scanner := bufio.NewScanner(file)

		const maxCapacity = 512 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		scanner.Scan()
		headers := strings.Split(scanner.Text(), "\t")

		docs := make([]interface{}, 0)

		for scanner.Scan() {
			doc := make(map[string]interface{})
			values := strings.Split(scanner.Text(), "\t")
			for j := range headers {
				header := headers[j]
				value := values[j]
				i, err := strconv.Atoi(value)
				if err == nil {
					doc[header] = i
					continue
				}
				f, err := strconv.ParseFloat(value, 64)
				if err == nil {
					doc[header] = f
					continue
				}
				d, err := time.Parse("2006-01-02 15:04:05", value)
				if err == nil {
					doc[header] = d
					continue
				}
				d, err = time.Parse("2006-01-02", value)
				if err == nil {
					doc[header] = d
					continue
				}
				if value == "0000-00-00 00:00:00" {
					doc[header] = time.Time{}
					continue
				}
				doc[header] = value
			}

			docs = append(docs, doc)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = col.InsertMany(ctx, docs)
		cancel()
		if err != nil {
			log.Fatalf("failed to insert into %s: %s", col.Name(), err)
		}

		file.Close()
	}
}

func main() {
	InsertIntoMongo("ourcampaigns-dump")
}

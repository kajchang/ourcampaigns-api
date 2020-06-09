package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertIntoMongo(dumpPath string) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("ourcampaigns")

	fileInfos, err := ioutil.ReadDir(dumpPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, fileInfo := range fileInfos {
		fmt.Println(fileInfo.Name())
		col := db.Collection(strings.Split(fileInfo.Name(), "-")[0])

		file, err := os.Open(path.Join(dumpPath, fileInfo.Name()))
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(file)

		scanner.Scan()
		headers := strings.Split(scanner.Text(), "\t")

		docs := make([]interface{}, 0)

		for scanner.Scan() {
			doc := make(map[string]string)

			values := strings.Split(scanner.Text(), "\t")
			for i := range headers {
				doc[headers[i]] = values[i]
			}

			docs = append(docs, doc)
		}

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = col.InsertMany(ctx, docs)
		if err != nil {
			log.Fatal(err)
		}

		file.Close()
	}
}

func main() {
	InsertIntoMongo("ourcampaigns-dump")
}

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
)

var DB_NAMES []string = []string{"User", "Container", "Race", "RaceMember", "Candidate", "Poll", "PollResult", "PollingFirm", "Endorsement", "FinanceReport", "Issue", "Party", "News", "Book", "InfoLink", "Leaning", "Prediction"}

func DownloadSiteData(dumpPath string) {
	err := os.Mkdir(dumpPath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("failed to create dump folder: %s", err)
	}

	for _, dbName := range DB_NAMES {
		for page := 0; true; page++ {
			res, err := http.Get(fmt.Sprintf("https://www.ourcampaigns.com/cgi-bin/datadownload.cgi?WhichDB=%s&WhichPage=%d", dbName, page))
			if err != nil {
				log.Fatalf("failed to request from ourcampaigns.com: %s", err)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)
			res.Body.Close()
			println(buf.Len())
			if buf.Len() <= 306 {
				break
			}

			downloadPath := path.Join(dumpPath, fmt.Sprintf("%s-%d.tsv", dbName, page))
			out, err := os.Create(downloadPath)
			fmt.Printf("Downloading %s...\n", downloadPath)
			out.Write(buf.Bytes())

			out.Close()
		}
	}
}

func main() {
	DownloadSiteData("ourcampaigns-dump")
}

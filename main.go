package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/robfig/cron"

	"github.com/zelenin/go-tdlib/client"
)

// Giveaways is our object that holds the content parsed from the groups file
type Giveaways struct {
	GroupID int64
	counter int64
}

var (
	filename       = flag.String("filename", "", "filename to read groups from.")
	every          = flag.String("every", "12h", "start a new round of giveaways randomly, every X times. default: 12h (12 hours).")
	totalgiveaways = flag.Int64("giveaways", 3, "number of total giveaways to make.")
	amount         = flag.Float64("amount", 0.00042, "Amount to give away per each giveaway.")
	apiID          = flag.Int64("apiID", 0, "your Telegram API ID.")
	apiHash        = flag.String("apiHASH", "", "your Telegram API HASH.")
	giveaways      []*Giveaways
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	flag.Parse()
	client.SetLogVerbosityLevel(1)

	// wait channel. a never ending story.
	wait := make(chan bool)
	// client authorizer
	authorizer := client.ClientAuthorizer()
	go client.CliInteractor(authorizer)

	authorizer.TdlibParameters <- &client.TdlibParameters{
		UseTestDc:              false,
		DatabaseDirectory:      filepath.Join(".tdlib", "database"),
		FilesDirectory:         filepath.Join(".tdlib", "files"),
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		UseSecretChats:         false,
		ApiId:                  int32(*apiID),
		ApiHash:                *apiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "Server",
		SystemVersion:          "1.0.0",
		ApplicationVersion:     "1.0.0",
		EnableStorageOptimizer: true,
		IgnoreFileNames:        false,
	}

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}

	me, err := tdlibClient.GetMe()
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	if len(*filename) == 0 {
		log.Printf("Me: %s %s [%s] %d\n", me.FirstName, me.LastName, me.Username, me.Id)

		chats, err := tdlibClient.GetChats(&client.GetChatsRequest{OffsetOrder: 9223372036854775807, Limit: 10000})
		if err != nil {
			log.Fatalf("GetChats error: %s", err)
		}

		for _, chat := range chats.ChatIds {
			chat, _ := tdlibClient.GetChat(&client.GetChatRequest{ChatId: chat})
			log.Printf("ID: %d - Name: %s", chat.Id, chat.Title)
		}

		fmt.Printf("\nChoose one or more groups from the list above and write their IDs line by line into a file.\n")
		fmt.Printf("If you are ready, you can run this app again with the filename as the first argument.\n")
		return
	}

	scanfile()

	c := cron.New()

	c.AddFunc(fmt.Sprintf("@every %s", *every), func() {
		mytotalgiveaways := *totalgiveaways
		if len(giveaways) == 0 {
			scanfile()
		}

		for mytotalgiveaways != 0 {
			for i, giveaway := range giveaways {
				logrus.Infof("%#v", giveaway)
				parsetime, err := time.ParseDuration(*every)
				if err != nil {
					check(err)
				}

				logrus.Info(int(parsetime.Seconds() / float64(*totalgiveaways)))
				rand.Seed(time.Now().UnixNano())
				n := rand.Intn(int(parsetime.Seconds() / float64(*totalgiveaways)))
				for n == 0 {
					logrus.Info(int(parsetime.Seconds() / float64(*totalgiveaways)))
					rand.Seed(time.Now().UnixNano())
					n = rand.Intn(int(parsetime.Seconds() / float64(*totalgiveaways)))
					logrus.Warn("n == 0")
					if n != 0 {
						break
					}
				}

				logrus.Infof("Sleeping for %s.", time.Duration(n)*time.Second)
				time.Sleep(time.Duration(n) * time.Second)

				msg, err := tdlibClient.SendMessage(&client.SendMessageRequest{
					ChatId: giveaway.GroupID,
					InputMessageContent: &client.InputMessageText{
						Text: &client.FormattedText{
							Text: fmt.Sprintf("/giveaway %v", *amount),
						},
					}})
				if err != nil {
					check(err)
				}
				giveaway.counter--
				logrus.Infof("%#v", msg)

				// also track totalgiveaways. user might have predefined giveaways per group
				// but doesn't want to giveaway more than totalgiveaways
				time.Sleep(time.Second * 1)
				mytotalgiveaways--
				logrus.Warn("totalgiveaways: ", mytotalgiveaways)
				if mytotalgiveaways == 0 {
					logrus.Warn("totalgiveaways depleted.")
					break
				}
				if giveaway.counter == 0 {
					logrus.Info("Group has no counter left. Skipping.")
					if len(giveaways) == 1 {
						break
					}
					giveaways = append(giveaways[:i], giveaways[i+1:]...)
					continue
				}
				giveaways[i] = giveaway
			}
		}
	})
	c.Start()
	<-wait
}

func scanfile() {
	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, ":")
		if len(split) < 2 {
			logrus.Errorf("Skipping wrong format of line: %s.", line)
			continue
		}

		chatid, err := strconv.ParseInt(split[0], 10, 64)
		if err != nil {
			check(err)
		}
		counter, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			logrus.Error(err)
			continue
		}
		giveaways = append(giveaways, &Giveaways{GroupID: chatid, counter: counter})
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// shuffle random groups from slice
	shuffle(giveaways)

	logrus.Warn("GROUPS SHUFFLED")
	for _, giveaway := range giveaways {
		logrus.Infof("%#v", giveaway)
	}
}

func shuffle(vals []*Giveaways) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(vals) > 0 {
		n := len(vals)
		randIndex := r.Intn(n)
		vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
		vals = vals[:n-1]
	}
}

package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v8"
)

const colorReset = "\033[0m"

const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorYellow = "\033[33m"
const colorBlue = "\033[34m"
const colorPurple = "\033[35m"
const colorCyan = "\033[36m"
const colorWhite = "\033[37m"

// to store command line arguments.
type cmdLineArgs struct {
	userFile         string
	processUserFile  bool
	threads          int
	channelBuffer    int
	clientId         string
	clientSecret     string
	clientRealm      string
	destinationRealm string
	url              string
}

//
// The structure of user details.
//
type simple_detail struct {
	user_name  string
	first_name string
	last_name  string
	password   string
	email      string
}

type KeycloakUser struct {
	FirstName string
	LastName  string
	Email     string
	Enabled   bool
	Username  string
}

func main() {

	// Read in the command line arguments
	var cmdArgs = processCommandLine()
	// Display the command line arguments back to the user.
	printCmdLineArgs(cmdArgs)
	// log the command line arguments to the log file.

	startTimeString := strconv.FormatInt(time.Now().Unix(), 10)

	startTime := makeTimestamp()
	startTime2 := nowAsUnixMilliseconds()

	log.Println(startTime, startTime2)

	f, err := os.OpenFile(startTimeString+"-testlogfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	rand.Seed(time.Now().UnixNano())
	log.SetFlags(0)
	logCmdLineArgs(cmdArgs)

	log.Println("STARTING")

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(cmdArgs.threads)

	usersChannel := make(chan []string, cmdArgs.channelBuffer)
	resultsChannel := make(chan string, cmdArgs.channelBuffer)

	go readUserFile(usersChannel, cmdArgs.userFile)
	go writeLog(resultsChannel)

	for i := 0; i < cmdArgs.threads; i++ {
		go createUserWorker(i, cmdArgs.clientRealm, cmdArgs.clientId, cmdArgs.clientSecret, cmdArgs.destinationRealm, cmdArgs.url, usersChannel, resultsChannel, &wgReceivers)
	}

	wgReceivers.Wait()
	log.Println("Total Time: ", makeTimestamp()-startTime)

	duration := makeTimestamp() - startTime
	println("Took " + strconv.FormatInt(duration, 10) + " time")
}

//
// reads file and adds data it to the channel
func readUserFile(jobs chan []string, fileName string) {
	log.Println("[START]: reading file ", fileName, "*******************************************")
	openedFileOk := false
	csvfile, err := os.Open(fileName)

	if err != nil {
		fmt.Println("[ERROR] Couldn't open the tsv file [", fileName, "]")
		log.Fatalln("Couldn't open the tsv file", fileName, err)
	} else {
		openedFileOk = true
	}
	defer csvfile.Close()

	if openedFileOk {
		reader := csv.NewReader(csvfile)

		reader.Comma = '\t' // Use tab-delimited instead of comma <---- here!

		reader.FieldsPerRecord = -1 // don't check

		csvData, err := reader.ReadAll()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		counter := 0

		for _, each := range csvData {
			counter++
			if (counter % 1000) == 0 {
				log.Println("Read ", counter, " rows")
			}
			jobs <- each
		}
		log.Println("END: reading file *******************************************")
	} else {
		log.Println("Failed to open file.")
	}
	close(jobs)
}

func writeLog(results chan string) {
	for j := range results {
		log.Println("RSLT: ", j)
	}
}

func createUserWorker(id int, realmName string, clientId string, clientSecret string, targetRealm string, url string, jobs <-chan []string, results chan<- string, wg *sync.WaitGroup) {

	defer func() {
		if r := recover(); r != nil {
			println("panic:" + r.(string))
		}
	}()

	successCounter := 0
	defaultPassword := "Letmein123"
	log.Println(id, " Bulk User Tool Starting")

	client := gocloak.NewClient(url)
	ids := strconv.Itoa(id)
	ctx := context.Background()
	log.Println(ids, "] logging into keycloak")
	token, err := client.LoginClient(ctx, clientId, clientSecret, realmName)

	if err != nil {
		log.Println(ids, "] ", token)
		log.Println(ids, "] ", err)
		panic(ids + "] Something wrong with the credentials or url : Error: " + err.Error())
	} else {
		log.Println(ids, "] Login Success", token)
	}
	//USER_ID	FIRST_NAME	LAST_NAME	PASSWORD	EMAIL
	for channelData := range jobs {
		ok := true
		user := gocloak.User{
			FirstName: gocloak.StringP(channelData[1]),
			LastName:  gocloak.StringP(channelData[2]),
			Email:     gocloak.StringP(channelData[4]),
			Enabled:   gocloak.BoolP(true),
			Username:  gocloak.StringP(channelData[0])}

		createdUser, err := client.CreateUser(ctx, token.AccessToken, targetRealm, user)
		if err != nil {
			log.Println(ids, "] create user error : ", err.Error())
			ok = false
			//panic("Oh no!, failed to create user :(")
			// set the return error code.
			results <- "[" + ids + "] " + channelData[3] + err.Error()
		} else {
			// if we need more logging.
			//log.Println(ids, "]created user success : ", createdUser)
		}
		// if it is still ok, then set the password, if it is not ok, then the password creation failed.
		if ok {
			err = client.SetPassword(ctx, token.AccessToken, createdUser, targetRealm, defaultPassword, false)
			if err != nil {
				log.Println(ids, "] SetPassword error: ", err.Error())
				results <- ids + "] " + channelData[0] + " can't set password" + err.Error()
			}
		}
		// if we get to the end, and it is still ok, then we assume the user is created
		// and we add the return message.
		if ok {
			successCounter++
			results <- "[" + ids + "] " + channelData[0] + " successfully created"
		}
		//log.Println(ids, "] ", channelData[0], " Created user and set password successfully")
	}
	defer wg.Done()

	log.Println("Worker [", ids, "] created ", successCounter, " users")
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func nowAsUnixMilliseconds() int64 {
	return time.Now().Round(time.Millisecond).UnixNano() / 1e6
}

func processCommandLine() cmdLineArgs {
	userFile := flag.String("userFile", "example-user-file.tsv", "The file name of a user details file.")
	processUserFile := flag.Bool("processUserFile", true, "Process user file")

	threads := flag.Int("thread", 10, "the number of threads to run the keycloak import")
	channelBuffer := flag.Int("channelBuffer", 10000, "the number of buffered spaces in the channel buffer")

	clientId := flag.String("clientId", "admin-cli", "The API user that will execute the calls.")
	clientSecret := flag.String("clientSecret", "16dbc557-4de1-46b5-973b-8e06e104c87e", "The secret for the keycloak user defined by `clientId`")
	clientRealm := flag.String("clientRealm", "master", "The realm in which the `client_id` exists")

	destinationRealm := flag.String("destinationRealm", "delete", "The realm in keycloak where the users are to be created. This may or may not be the same as the `clientRealm`")

	url := flag.String("url", "http://localhost:8080/", "The URL of the keycloak server.")

	flag.Parse()

	if len(flag.Args()) > 0 {
		log.Println("unknown options specified:", flag.Args())
		fmt.Println("unknown options specified:", flag.Args())
		fmt.Println(string(colorRed), "[ERROR]", string(colorReset), "exiting")
		os.Exit(1)
	}
	return cmdLineArgs{userFile: *userFile, processUserFile: *processUserFile, channelBuffer: *channelBuffer,
		clientId: *clientId, clientSecret: *clientSecret, clientRealm: *clientRealm, destinationRealm: *destinationRealm, url: *url, threads: *threads}
}

// Print the command line arguments on the screen.
func printCmdLineArgs(cLA cmdLineArgs) {
	fmt.Println("[KeyCloak Bulk Import via API]")

	fmt.Println("userfile:", cLA.userFile)
	fmt.Println("channelBuffer", cLA.channelBuffer)
	fmt.Println("clientId", cLA.clientId)
	fmt.Println("clientSecret", cLA.clientSecret)
	fmt.Println("clientRealm", cLA.clientRealm)
	fmt.Println("destinationRealm", cLA.destinationRealm)
	fmt.Println("url", cLA.url)
	fmt.Println("threads:", cLA.threads)
}

// Send the command line arguments to the log file
func logCmdLineArgs(cLA cmdLineArgs) {
	log.Println("[KeyCloak Bulk Import via API]")
	log.Println("userfile:", cLA.userFile)
	log.Println("channelBuffer", cLA.channelBuffer)
	log.Println("clientId", cLA.clientId)
	log.Println("clientSecret", cLA.clientSecret)
	log.Println("clientRealm", cLA.clientRealm)
	log.Println("destinationRealm", cLA.destinationRealm)
	log.Println("url", cLA.url)
	log.Println("threads:", cLA.threads)
}

func testColour() {
	fmt.Println(string(colorRed), "test", string(colorReset))
	fmt.Println(string(colorGreen), "test", string(colorReset))
	fmt.Println(string(colorYellow), "test", string(colorReset))
	fmt.Println(string(colorBlue), "test", string(colorReset))
	fmt.Println(string(colorPurple), "test", string(colorReset))
	fmt.Println(string(colorWhite), "test", string(colorReset))
	fmt.Println(string(colorCyan), "test", string(colorReset))
	fmt.Println("next")
}

//user := gocloak.User{
//	FirstName: gocloak.StringP(channelData[1]),
//	LastName:  gocloak.StringP(channelData[2]),
//	Email:     gocloak.StringP(channelData[7]),
//	Enabled:   gocloak.BoolP(true),
//	Username:  gocloak.StringP(channelData[0])}

//func readPropertiesFile() {
//	p, err := Load([]byte("key=value\nkey2=${key}"), ISO_8859_1)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Dump the expanded key/value pairs of the Properties
//	fmt.Println("Expanded key/value pairs")
//	fmt.Println(p)
//}

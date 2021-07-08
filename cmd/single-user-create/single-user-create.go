package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Nerzal/gocloak/v8"
)

// to store command line arguments.
type cmdLineArgs struct {
	clientId         string
	clientSecret     string
	clientRealm      string
	destinationRealm string
	url              string
	user_name        string
	first_name       string
	last_name        string
	password         string
	email            string
	groupId          string
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
	//printCmdLineArgs(cmdArgs)
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

	log.Println("STARTING")

	createUser(cmdArgs.clientId,
		cmdArgs.clientSecret,
		cmdArgs.clientRealm,
		cmdArgs.destinationRealm,
		cmdArgs.url,
		cmdArgs.user_name,
		cmdArgs.first_name,
		cmdArgs.last_name,
		cmdArgs.password,
		cmdArgs.email,
		cmdArgs.groupId)

	log.Println("Total Time: ", makeTimestamp()-startTime)

	duration := makeTimestamp() - startTime
	println("Took " + strconv.FormatInt(duration, 10) + " time")
}

func createUser(clientId string,
	clientSecret string,
	clientRealm string,
	targetRealm string,
	url string,
	user_name string,
	first_name string,
	last_name string,
	password string,
	email string,
	groupId string) {

	ok := true
	client := gocloak.NewClient(url)
	log.Println(" connecting to:", url)
	ctx := context.Background()
	log.Println(" logging into keycloak:", clientId, clientSecret, clientRealm)
	token, err := client.LoginClient(ctx, clientId, clientSecret, clientRealm)

	if err != nil {
		log.Println("token:", token)
		log.Println("err:", err)
		panic(" Something wrong with the credentials or url : Error: " + err.Error())
	} else {
		log.Println("] Login Success", token)
	}

	user := gocloak.User{
		FirstName: gocloak.StringP(first_name),
		LastName:  gocloak.StringP(last_name),
		Email:     gocloak.StringP(email),
		Enabled:   gocloak.BoolP(true),
		Username:  gocloak.StringP(user_name)}

	createdUser, err := client.CreateUser(ctx, token.AccessToken, targetRealm, user)
	if err != nil {
		log.Println(" create user error : ", err.Error())
		ok = false
		//panic("Oh no!, failed to create user :(")
	} else {
		// if we need more logging.
		log.Println("created user success : ", createdUser)
	}
	// if it is still ok, then set the password, if it is not ok, then the password creation failed.
	if ok {
		err = client.SetPassword(ctx, token.AccessToken, createdUser, targetRealm, password, false)
		if err != nil {
			log.Println("] SetPassword error: ", err.Error())
		}
	}
	// if the user was created, then try to add to group.
	if ok {
		// If a group was passed in, then add the user to the group.
		if groupId != "" {
			log.Println("add user to group:", groupId)
			err = client.AddUserToGroup(ctx, token.AccessToken, targetRealm, createdUser, groupId)
			if err != nil {
				log.Println("] AddUserToGroup error: ", err.Error())
			}
		}
	}
	// if we get to the end, and it is still ok, then we assume the user is created
	// and we add the return message.

	//log.Println( "] ", channelData[0], " Created user and set password successfully")
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func nowAsUnixMilliseconds() int64 {
	return time.Now().Round(time.Millisecond).UnixNano() / 1e6
}

func processCommandLine() cmdLineArgs {

	clientId := flag.String("clientId", "admin-cli", "The API user that will execute the calls.")
	clientSecret := flag.String("clientSecret", "16dbc557-4de1-46b5-973b-8e06e104c87e", "The secret for the keycloak user defined by `clientId`")
	clientRealm := flag.String("clientRealm", "master", "The realm in which the `client_id` exists")

	destinationRealm := flag.String("destinationRealm", "delete", "The realm in keycloak where the users are to be created. This may or may not be the same as the `clientRealm`")

	url := flag.String("url", "http://127.0.0.1:8080/", "The URL of the keycloak server.")

	groupId := flag.String("forceGroup", "", "Add all users for this import to this group")

	// Column Definitions
	user_name := flag.String("userName", "", "The username value")
	first_name := flag.String("firstName", "", "The first name value")
	last_name := flag.String("lastName", "", "The last name value")
	password := flag.String("password", "", "The password value")
	email := flag.String("email", "", "The email value")

	flag.Parse()

	return cmdLineArgs{clientId: *clientId,
		clientSecret:     *clientSecret,
		clientRealm:      *clientRealm,
		destinationRealm: *destinationRealm,
		url:              *url,
		user_name:        *user_name,
		first_name:       *first_name,
		last_name:        *last_name,
		password:         *password,
		email:            *email,
		groupId:          *groupId}
}

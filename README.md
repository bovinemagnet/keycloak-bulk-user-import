# keycloak-bulk-user-import

Simple go application to bulk load users from a CSV or TSV file.

Multi-threadded, allowing the parrallel import of a lot of users into keycloak via API.

## Stages

This is my plan for this little tool.

1. Import custom simple file, consisting of `USER_ID`	`FIRST_NAME`	`LAST_NAME`	`PASSWORD`	`EMAIL`
2. Import TSV/CSV utilising header rows to determine the columns needed for importation.
3. Large file/long time support, currently a large file will timeout, as the keycloak session is only active for a short period by default.

## Usage

To execute the program, utilise the following command line such as `bulk-user-create -help`

```bash
$ ./bulk-user-create -help 
Usage of ./bulk-user-create:
  -channelBuffer int
        the number of buffered spaces in the channel buffer (default 10000)
  -clientId string
        The API user that will execute the calls. (default "admin-cli")
  -clientRealm client_id
        The realm in which the client_id exists (default "master")
  -clientSecret clientId
        The secret for the keycloak user defined by clientId (default "16dbc557-4de1-46b5-973b-8e06e104c87e")
  -destinationRealm clientRealm
        The realm in keycloak where the users are to be created. This may or may not be the same as the clientRealm (default "delete")
  -processUserFile
        Process user file (default true)
  -thread int
        the number of threads to run the keycloak import (default 10)
  -url string
        The URL of the keycloak server. (default "http://localhost:8080/")
  -userFile string
        The file name of a user details file. (default "example-user-file.tsv")
```

## Acknowledgements

This package makes use of the following libraries and packages.

* https://github.com/Nerzal/gocloak/

### Release Management

Handled by

* https://github.com/goreleaser/goreleaser
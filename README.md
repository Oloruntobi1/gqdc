# Wallet Service

## Architecture

```
internal
├── database
│   ├── database.go
│   ├── mysql.go
│   └── filesystem.go
├── server
│   ├── server.go
│   └── wallet_handler.go
├── middleware
│   ├── auth.go
│   └── cache.go

```

The above folder structure gives an overview of the core abstraction layers of the application. Let's talk about each layer.

### Database

This is where you define all database engines used by the application. To add a new database enigine, simply add a new file to the `database` folder and define the interface specified in the `database.go` file.

To use InMemory database or any storage mechanism of your choice, just update the
`CURRENT_STORAGE` env variable. For now `MYSQL` and `FileSystem`(which i did not fully
implement) are the acceptable ones with it defaulting to `InMemory`(which i fully implemented) if the env variable isn't set.

### Server

This layer is responsible for handling all HTTP requests. It usually contains a Service and/or a Repository/Database. The Service is responsible for handling the business logic and the Repository/Database is responsible for retrieving data from the database.

### Middleware

The middleware folder houses all the files that tackles any sort of middlware you wish to implement. In this submission I have just three:

i) `auth.go` -- for authentication,
ii) `cache.go` -- for caching,
ii) `logger.go` -- for logging

## Running the Application

Once you have docker and docker-compose installed, you can run the `$ make up` to bring up the services locally.

Then make a `GET` request to `http://localhost:8080/api/v1/users` which returns all users (again with the password in clear text for sake of this assignment), after this login with the password and email to get a token which can now be used to perform other
operations as contained in the API Documentation.

## Running Tests

All tests can be run with `$ make test`.

## Useful Make Commands

Run any of the following command prefixing `$ make <command>`:

`

help:  Output available commands
setup:  Builds the web container
start:  Start all services
dev:  Run the web server in dev mode without using docker
mysql:  Starts the mysql server
stop:  Stop all services
destroy:  Remove all containers and images. Also, destroy all volumes
test:  Run all tests
hot-reload:  Enables hot reload for the web service
go-format:  Run go fmt ./... on all go files
`

## Documentation

[Click](https://documenter.getpostman.com/view/9095594/UVyxRZXy)

## Other things to note

a) I added a tiny bit extra security on the `GetBalance -- GET /api/v1/wallets/{wallet_id}/balance` which prevents a logged in user from getting someone else's
wallet balance. I did not add this to get `Credit` and `Debit` endpoints however,
because this business rule is not originally part of the assignment in the first
place.

b) Please note that in the handler for creating a user, in the `internal/server/user.go` file, I put in a comment there where i imagined a wallet being created for each user
in the background while signing up.

c) Also, I am saving the user password in clear text for the sake of testing only.
(What i mean is, say you want to quickly login and test with one of the already seeded users who actually have wallets created for them, you would need to supply the correct password, which is why i saved it in clear text in the database. Check files `internal/models/user` and `cmd/web/main.go` for context)

d) Normally the .env file will be in gitignore

In addition these other tools inclduing Prometheus, Grafana, Jaeger, BlackExporters, AlertManager etc. can be further utilized to achieve more metric and have proper understanding of the overall application's health at every point in time, but i guess I should get in first before i start demonstrating these skills.

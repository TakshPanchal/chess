# Chess 

A Platform where folks can 
1. Join as guest
2. Play or watch a match
3. Have a video or audio chat while playing

### Tech Stack 
1. React for frontend 
2. Typescript lang
3. Golang for backend using std libraries and `gorilla/mux`
4. Storing game in Postgres DB
5. Using Redis as a queue


### Setup 
Make sure golang 1.21 is installed on the system. 
From the root directory run.
```
go mod tidy
```
To start the postgres and reddis docker containers run from the root. Make sure there will be `.env` file with the required variables as per `sample.env` file.
```
docker compose up docker-compose -d
```
To run the server app run. The server will be listing to `8080` port. 
```
go run src/main/*
```
To run the start the frontend app run 
```
cd frontend
npm install
npm run dev 
```


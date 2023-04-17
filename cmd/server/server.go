package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/LucianTavares/comunicacao_entre_sistemas/graphql/graph"
	"github.com/LucianTavares/comunicacao_entre_sistemas/graphql/graph/dataloader"
	"github.com/LucianTavares/comunicacao_entre_sistemas/graphql/graph/generated"
	"github.com/LucianTavares/comunicacao_entre_sistemas/graphql/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

const defaultPort = "8080"

func main() {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	categoryDb := database.NewCategory(db)
	courseDb := database.NewCourse(db)

	loader := dataloader.NewDataLoader(categoryDb)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		CourseDB:   courseDb,
		CategoryDB: categoryDb,
	}}))

	dataloaderSrv := dataloader.Middleware(loader, srv)

	http.Handle("/query", dataloaderSrv)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

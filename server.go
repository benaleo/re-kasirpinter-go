package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"re-kasirpinter-go/config"
	"re-kasirpinter-go/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		log.Printf("Warning: Failed to load environment variables: %v", err)
	}

	// Initialize database
	db, err := config.InitDb()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize email queue
	graph.GetEmailQueue()
	log.Println("Email queue initialized with background workers")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers:  &graph.Resolver{DB: db},
		Directives: graph.DirectiveRoot{Auth: graph.AuthDirective},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	http.HandleFunc("/kaitheathcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	http.Handle("/graphql", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", graph.AuthMiddleware(srv))

	fmt.Println("Listening on 0.0.0.0:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"re-kasirpinter-go/config"
	"re-kasirpinter-go/graph"
	"re-kasirpinter-go/service"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		origin := r.Header.Get("Origin")

		// If ALLOWED_ORIGINS is set to "*", allow all origins
		if allowedOrigins == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if allowedOrigins != "" && origin != "" {
			// Check if the origin is in the allowed list
			allowedList := strings.Split(allowedOrigins, ",")
			for _, allowed := range allowedList {
				allowed = strings.TrimSpace(allowed)
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// Set other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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

	// Initialize R2 service
	r2Service, err := service.NewR2Service()
	if err != nil {
		log.Printf("Warning: Failed to initialize R2 service: %v", err)
	}

	// Initialize user service
	userService, err := service.NewUserService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize user service: %v", err)
	}

	authService := service.NewAuthService(db)

	// Initialize role service
	roleService, err := service.NewRoleService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize role service: %v", err)
	}

	// Initialize ingredient category service
	ingredientCategoryService, err := service.NewIngredientCategoryService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize ingredient category service: %v", err)
	}

	// Initialize ingredient service
	ingredientService, err := service.NewIngredientService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize ingredient service: %v", err)
	}

	// Initialize ingredient stock service
	ingredientStockService, err := service.NewIngredientStockService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize ingredient stock service: %v", err)
	}

	// Initialize product category service
	productCategoryService, err := service.NewProductCategoryService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize product category service: %v", err)
	}

	// Initialize product service
	productService, err := service.NewProductService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize product service: %v", err)
	}

	// Initialize discount service
	discountService, err := service.NewDiscountService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize discount service: %v", err)
	}

	// Initialize product variant service
	productVariantService, err := service.NewProductVariantService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize product variant service: %v", err)
	}

	// Initialize product ingredient service
	productIngredientService := service.NewProductIngredientService(db)

	// Initialize product extra service
	productExtraService, err := service.NewProductExtraService(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize product extra service: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers:  &graph.Resolver{DB: db, R2Service: r2Service, UserService: userService, AuthService: authService, RoleService: roleService, IngredientCategoryService: ingredientCategoryService, IngredientService: ingredientService, IngredientStockService: ingredientStockService, ProductCategoryService: productCategoryService, ProductService: productService, DiscountService: discountService, ProductVariantService: productVariantService, ProductIngredientService: productIngredientService, ProductExtraService: productExtraService},
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
	srv.Use(&graph.LoggingInterceptor{})

	// Wrap handlers with CORS middleware
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	mux.HandleFunc("/kaitheathcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	mux.HandleFunc("/graphql", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", graph.AuthMiddleware(db)(srv))

	// Apply CORS middleware to all routes
	handler := corsMiddleware(mux)

	fmt.Println("Listening on 0.0.0.0:" + port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, handler))
}

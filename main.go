package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/yuheng-liu/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
	jwtSecret      string
	polkaKey       string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// by default, godotenv will look for a file named .env in the current directory
	godotenv.Load()
	// retrieve the env value for jwt secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	// retrieve the env value for authorization key
	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}
	// creates a new .json db with file name "database.json"
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}
	// checks if it's in debug mode, will delete database if yes
	dbg := flag.Bool("debug", false, "Enable debug mode")
	// parse flags entered via command line and saved result to dbg pointer
	flag.Parse()
	if dbg != nil && *dbg {
		// deletes previous database.json file
		err := db.ResetDB()
		if err != nil {
			log.Fatal(err)
		}
	}
	// init apiConfig struct
	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	// api common
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
	// chirps
	apiRouter.Post("/chirps", apiCfg.handlerChirpsCreate)
	apiRouter.Get("/chirps", apiCfg.handlerChirpsRetrieve)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.handlerChirpsGet)
	apiRouter.Delete("/chirps/{chirpID}", apiCfg.handlerChirpsDelete)
	// users
	apiRouter.Post("/login", apiCfg.handlerLogin)
	apiRouter.Post("/refresh", apiCfg.handlerRefresh)
	apiRouter.Post("/revoke", apiCfg.handlerRevoke)
	apiRouter.Post("/users", apiCfg.handlerUsersCreate)
	apiRouter.Put("/users", apiCfg.handlerUsersUpdate)
	// polka webhook
	apiRouter.Post("/polka/webhooks", apiCfg.handlerWebhook)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)
	router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

	server := &http.Server{
		Handler: corsMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

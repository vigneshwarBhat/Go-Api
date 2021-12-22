package app

import (
	"Product-Api/config"
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
	Config *config.Configuration
	Server *http.Server
}

func (a *App) Initialize() {
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	errInReading := viper.ReadInConfig()
	if errInReading != nil {
		log.Fatal("error in reading environment file", errInReading)
	}

	err1 := viper.Unmarshal(&a.Config)
	if err1 != nil {
		log.Fatal(err1)
		return
	}
	var err error
	a.DB, err = sql.Open("mssql", a.Config.ConnectionString)
	if err != nil {
		log.Fatal(err)
		return
	}

	pingErr := a.DB.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
		return
	}

	a.Router = mux.NewRouter()
	a.InitRoutes()
}

func (a *App) Run(addr string) {
	a.Server = &http.Server{Addr: ":8010", Handler: a.Router}
	a.Server.ListenAndServe()
	log.Printf("server started running on port 8010")
}

func (a *App) InitRoutes() {
	a.Router.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {

		a.Server.Shutdown(context.Background())
	})
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/product/Create", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/Update/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/Delete/{id:[0-9]+}", a.updateProduct).Methods("DELETE")
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	if count < 10 || count < 1 {
		count = 10
	}
	products, err := getProducts(a.DB, count)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, products)
}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid product id")
		return
	}
	p := Product{Id: id}
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			errorResponse(w, http.StatusNotFound, "No product")
		default:
			errorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJson(w, http.StatusOK, p)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&product); err != nil {
		errorResponse(w, http.StatusBadRequest, "Request is Invalid")
		return
	}
	defer r.Body.Close()
	if err := product.createProduct(a.DB); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, product)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid product id")
		return
	}
	decoder := json.NewDecoder(r.Body)
	var product Product
	if err := decoder.Decode(&product); err != nil {
		errorResponse(w, http.StatusBadRequest, "Bad request payload")
		return
	}
	defer r.Body.Close()
	product.Id = id
	if err := product.updateProduct(a.DB); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, product)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid product id")
		return
	}
	p := Product{Id: id}
	if err := p.deleteProduct(a.DB); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusNoContent, map[string]string{"result": "success"})
}

func errorResponse(w http.ResponseWriter, s int, e string) {
	respondWithJson(w, s, map[string]string{"error": e})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	jsonResponse, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonResponse)
}

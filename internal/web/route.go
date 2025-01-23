package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Маршрутизатор
func (app *WebApp) SetRoutes() *mux.Router {
	router := mux.NewRouter()

	// Ограничение количества запросов от одного IP
	router.Use(LimitMiddleware)

	router.HandleFunc("/debug", app.DebugHandler)

	// Валидация и пересылка
	router.HandleFunc("/", app.HandleValidate).Methods("GET")
	// Главная страница
	router.HandleFunc("/form", app.HandleForm).Methods("GET")

	router.HandleFunc("/is-admin", app.HandleIsAdmin).Methods("GET")
	router.HandleFunc("/admin-menu", app.HandleAdminMenu).Methods("GET")
	router.HandleFunc("/admin-command", app.HandleAdminCommand).Methods("GET")

	router.HandleFunc("/send-data", app.HandleSendData).Methods("POST")

	staticDir := "./ui/static/"
	fileServer := http.FileServer(http.Dir(staticDir))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	return router
}

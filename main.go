package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"text/template"

	"github.com/barelyhuman/go/env"
	"github.com/barelyhuman/notdone/lib"
	"github.com/barelyhuman/notdone/storage"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/lucsky/cuid"
	"github.com/sevlyar/go-daemon"
)

//go:embed templates/*.html templates/**/*.html
var templateFS embed.FS

//go:embed dist/style.css
var baseStyles string

var (
	stop   = make(chan struct{})
	config = &lib.Config{}
)

var (
	signal = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)
)

func main() {
	flag.Parse()

	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termSigHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termSigHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadSigHandler)

	config = lib.LoadConfig()
	cwd := lib.GetBaseDirectory()
	processName := getProcessName()
	pidName := filepath.Join(cwd, processName+".pid")
	logName := filepath.Join(cwd, processName+".log")

	daemonCtx := &daemon.Context{
		PidFileName: pidName,
		PidFilePerm: 0644,
		LogFileName: logName,
		LogFilePerm: 0640,
		WorkDir:     cwd,
		Umask:       027,
		Args:        []string{},
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := daemonCtx.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}

	d, err := daemonCtx.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}

	defer daemonCtx.Release()
	go startServer()

	err = daemon.ServeSignals()
	lib.Bail(err)
}

func termSigHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	return daemon.ErrStop
}

func reloadSigHandler(sig os.Signal) error {
	// TODO: reload config
	config = lib.LoadConfig()
	stop <- struct{}{}
	go startServer()
	return nil
}

func startServer() {
	r := mux.NewRouter()

	godotenv.Load(".env")
	store := storage.New()

	appData := AppData{
		Store: store,
	}

	appData.Load()

	templates, err := template.New("").Funcs(template.FuncMap{
		"baseStyles": func() string {
			return "<style>" + baseStyles + "</style>"
		},
	}).ParseFS(templateFS, "templates/**/*.html", "templates/*.html")

	lib.Bail(err)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		appData.Load()

		templates.ExecuteTemplate(w, "Index", map[string]interface{}{
			"Lists":          appData.Lists,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
	}).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		listName := r.FormValue("listName")

		isDelete := r.FormValue("_method") == "delete"

		if isDelete {
			listId := r.FormValue("listId")
			appData.DeleteList(listId)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		appData.AddList(listName)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}).Methods("POST")

	r.HandleFunc("/{listId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		listIndex := appData.FindListIndex(vars["listId"])

		if listIndex == -1 {
			lib.MustParse("error", "invalid index").Execute(w, map[string]interface{}{})
			return
		}

		err = templates.ExecuteTemplate(w, "ListDetails", map[string]interface{}{
			"List":           appData.Lists[listIndex],
			csrf.TemplateTag: csrf.TemplateField(r),
		})

	}).Methods("GET")

	r.HandleFunc("/{listId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		r.ParseForm()

		listId := vars["listId"]
		listIndex := appData.FindListIndex(listId)
		isDelete := r.FormValue("_method") == "delete"

		if isDelete {
			taskId := r.FormValue("taskId")
			if len(taskId) > 0 {
				appData.DeleteTask(listId, taskId)
				http.Redirect(w, r, "/"+listId, http.StatusSeeOther)
			}
			return
		}

		if listIndex == -1 {
			lib.MustParse("error", "invalid index").Execute(w, map[string]interface{}{})
			return
		}

		appData.AddItem(listId, r.FormValue("content"))
		appData.Load()
		http.Redirect(w, r, "/"+listId, http.StatusSeeOther)
	}).Methods("POST")

	r.HandleFunc("/{listId}/{taskId}/mark", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		r.ParseForm()

		listId := vars["listId"]
		taskId := vars["taskId"]
		appData.ToggleMarked(listId, taskId)

		fmt.Fprintf(w, "successful")

	}).Methods("POST")

	fmt.Printf("Listening on :%v", config.Port)

	csrfSecretToken := env.Get("CSRF_TOKEN", lib.CryptoRandomToken(32))

	CSRF := csrf.Protect([]byte(csrfSecretToken), csrf.SameSite(csrf.SameSiteStrictMode), csrf.Path("/"), csrf.Secure(false))

	h := &http.Server{Addr: ":" + config.Port, Handler: CSRF(r)}

	go func() {
		err := h.ListenAndServe()
		if err != nil {
			log.Printf("server failed to run: %s", err.Error())
		}
	}()

	<-stop
	h.Shutdown(context.Background())
}

func getProcessName() string {
	return `notdone`
}

type ListItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Marked  bool   `json:"marked"`
}

type List struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Items          []ListItem `json:"items"`
	CompletedItems int        `json:"items.completed"`
}

func (l *List) FindTaskIndex(taskId string) int {
	for ind, i := range l.Items {
		if i.ID == taskId {
			return ind
		}
	}
	return -1
}

func (l *List) UpdateCompleted() {
	l.CompletedItems = 0
	for _, li := range l.Items {
		if !li.Marked {
			continue
		}

		l.CompletedItems += 1
	}
}

type AppData struct {
	Store *storage.Storage `json:"-"`
	Lists []List           `json:"lists"`
}

func (ad *AppData) Load() {
	ad.Store.Load(&ad)
}

func (ad *AppData) Save() {
	ad.Store.Save(&ad)
}

func (ad *AppData) AddList(name string) {
	ad.Lists = append(ad.Lists, List{
		ID:    cuid.New(),
		Name:  name,
		Items: []ListItem{},
	})
	ad.RunComputedData()
	ad.Save()
}

func (ad *AppData) RunComputedData() {
	for i := range ad.Lists {
		ad.Lists[i].UpdateCompleted()
	}
}

func (ad *AppData) FindListIndex(listId string) int {
	for ind, l := range ad.Lists {
		if l.ID == listId {
			return ind
		}
	}
	return -1
}

func (ad *AppData) AddItem(listId string, content string) {
	listIndex := ad.FindListIndex(listId)
	if len(content) == 0 {
		return
	}
	ad.Lists[listIndex].Items = append(ad.Lists[listIndex].Items, ListItem{
		ID:      cuid.New(),
		Content: content,
		Marked:  false,
	})
	ad.RunComputedData()
	ad.Save()
}

func (ad *AppData) DeleteTask(listId string, taskId string) {
	listIndex := ad.FindListIndex(listId)

	newList := []ListItem{}

	for _, i := range ad.Lists[listIndex].Items {
		if i.ID == taskId {
			continue
		}
		newList = append(newList, i)
	}

	ad.Lists[listIndex].Items = newList

	ad.RunComputedData()
	ad.Save()
}

func (ad *AppData) ToggleMarked(listId string, taskId string) {
	listIndex := ad.FindListIndex(listId)
	taskIndex := ad.Lists[listIndex].FindTaskIndex(taskId)
	if taskIndex == -1 {
		return
	}
	ad.Lists[listIndex].Items[taskIndex].Marked = !ad.Lists[listIndex].Items[taskIndex].Marked
	ad.RunComputedData()
	ad.Save()
}

func (ad *AppData) DeleteList(listId string) {
	listIndex := ad.FindListIndex(listId)

	ad.Lists = append(ad.Lists[:listIndex], ad.Lists[listIndex+1:]...)

	ad.RunComputedData()
	ad.Save()
}

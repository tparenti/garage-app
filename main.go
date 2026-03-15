package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var db *sql.DB
var tmpl *template.Template

func main() {
	db = initDB("./wrenchlog.db")
	defer db.Close()

	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Format("Jan 02, 2006")
		},
		"formatMileage": func(m int) string {
			if m == 0 {
				return "—"
			}
			s := strconv.Itoa(m)
			// insert commas
			n := len(s)
			var b strings.Builder
			for i, c := range s {
				if i > 0 && (n-i)%3 == 0 {
					b.WriteRune(',')
				}
				b.WriteRune(c)
			}
			return b.String()
		},
		"vehicleTitle": func(v Vehicle) string {
			parts := []string{}
			if v.Year != "" {
				parts = append(parts, v.Year)
			}
			parts = append(parts, v.Make, v.Model)
			return strings.Join(parts, " ")
		},
		"add": func(a, b int) int { return a + b },
	}

	tmpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Dashboard
	mux.HandleFunc("/", handleDashboard)

	// Owners
	mux.HandleFunc("/owners", handleOwners)
	mux.HandleFunc("/owners/new", handleOwnerNew)
	mux.HandleFunc("/owners/", handleOwnerDetail) // /owners/{id}, /owners/{id}/edit, /owners/{id}/delete

	// Vehicles
	mux.HandleFunc("/vehicles/new", handleVehicleNew)
	mux.HandleFunc("/vehicles/", handleVehicleDetail)

	// Logs
	mux.HandleFunc("/logs/new", handleLogNew)
	mux.HandleFunc("/logs/", handleLogDetail)

	fmt.Println("┌─────────────────────────────────┐")
	fmt.Println("│  WrenchLog running on :8080     │")
	fmt.Println("│  http://localhost:8080          │")
	fmt.Println("└─────────────────────────────────┘")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Now()
	}
	return t
}

func pathSegments(r *http.Request, prefix string) []string {
	p := strings.TrimPrefix(r.URL.Path, prefix)
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	owners, _ := getAllOwners(db)
	recent, _ := getRecentLogs(db, 8)

	// Count totals
	var totalVehicles, totalLogs int
	db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&totalVehicles)
	db.QueryRow(`SELECT COUNT(*) FROM maintenance_logs`).Scan(&totalLogs)

	render(w, "dashboard.html", map[string]any{
		"Owners":        owners,
		"Recent":        recent,
		"TotalOwners":   len(owners),
		"TotalVehicles": totalVehicles,
		"TotalLogs":     totalLogs,
	})
}

// ── Owners ────────────────────────────────────────────────────────────────────

func handleOwners(w http.ResponseWriter, r *http.Request) {
	owners, _ := getAllOwners(db)
	render(w, "owners.html", map[string]any{"Owners": owners})
}

func handleOwnerNew(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		o := Owner{
			Name:  r.FormValue("name"),
			Phone: r.FormValue("phone"),
			Email: r.FormValue("email"),
			Notes: r.FormValue("notes"),
		}
		id, err := createOwner(db, o)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/owners/%d", id), http.StatusSeeOther)
		return
	}
	render(w, "owner_form.html", map[string]any{"Title": "New Owner", "Owner": Owner{}})
}

func handleOwnerDetail(w http.ResponseWriter, r *http.Request) {
	segs := pathSegments(r, "/owners/")
	if len(segs) == 0 {
		http.Redirect(w, r, "/owners", http.StatusSeeOther)
		return
	}
	id := parseInt(segs[0])
	action := ""
	if len(segs) > 1 {
		action = segs[1]
	}

	switch action {
	case "edit":
		owner, err := getOwner(db, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPost {
			r.ParseForm()
			owner.Name = r.FormValue("name")
			owner.Phone = r.FormValue("phone")
			owner.Email = r.FormValue("email")
			owner.Notes = r.FormValue("notes")
			updateOwner(db, owner)
			http.Redirect(w, r, fmt.Sprintf("/owners/%d", id), http.StatusSeeOther)
			return
		}
		render(w, "owner_form.html", map[string]any{"Title": "Edit Owner", "Owner": owner})

	case "delete":
		if r.Method == http.MethodPost {
			deleteOwner(db, id)
			http.Redirect(w, r, "/owners", http.StatusSeeOther)
		}

	default:
		owner, err := getOwner(db, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		vehicles, _ := getVehiclesByOwner(db, id)
		owner.Vehicles = vehicles
		render(w, "owner_detail.html", map[string]any{"Owner": owner})
	}
}

// ── Vehicles ──────────────────────────────────────────────────────────────────

func handleVehicleNew(w http.ResponseWriter, r *http.Request) {
	ownerID := parseInt(r.URL.Query().Get("owner_id"))
	if r.Method == http.MethodPost {
		r.ParseForm()
		v := Vehicle{
			OwnerID:  parseInt(r.FormValue("owner_id")),
			Year:     r.FormValue("year"),
			Make:     r.FormValue("make"),
			Model:    r.FormValue("model"),
			VIN:      r.FormValue("vin"),
			LicPlate: r.FormValue("lic_plate"),
			Color:    r.FormValue("color"),
			Notes:    r.FormValue("notes"),
		}
		id, err := createVehicle(db, v)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/vehicles/%d", id), http.StatusSeeOther)
		return
	}
	owners, _ := getAllOwners(db)
	render(w, "vehicle_form.html", map[string]any{
		"Title":   "New Vehicle",
		"Vehicle": Vehicle{OwnerID: ownerID},
		"Owners":  owners,
	})
}

func handleVehicleDetail(w http.ResponseWriter, r *http.Request) {
	segs := pathSegments(r, "/vehicles/")
	if len(segs) == 0 {
		http.Redirect(w, r, "/owners", http.StatusSeeOther)
		return
	}
	id := parseInt(segs[0])
	action := ""
	if len(segs) > 1 {
		action = segs[1]
	}

	switch action {
	case "edit":
		v, err := getVehicle(db, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPost {
			r.ParseForm()
			v.Year = r.FormValue("year")
			v.Make = r.FormValue("make")
			v.Model = r.FormValue("model")
			v.VIN = r.FormValue("vin")
			v.LicPlate = r.FormValue("lic_plate")
			v.Color = r.FormValue("color")
			v.Notes = r.FormValue("notes")
			updateVehicle(db, v)
			http.Redirect(w, r, fmt.Sprintf("/vehicles/%d", id), http.StatusSeeOther)
			return
		}
		owners, _ := getAllOwners(db)
		render(w, "vehicle_form.html", map[string]any{"Title": "Edit Vehicle", "Vehicle": v, "Owners": owners})

	case "delete":
		if r.Method == http.MethodPost {
			v, _ := getVehicle(db, id)
			deleteVehicle(db, id)
			http.Redirect(w, r, fmt.Sprintf("/owners/%d", v.OwnerID), http.StatusSeeOther)
		}

	default:
		v, err := getVehicle(db, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		logs, _ := getLogsByVehicle(db, id)
		v.Logs = logs
		render(w, "vehicle_detail.html", map[string]any{"Vehicle": v})
	}
}

// ── Maintenance Logs ──────────────────────────────────────────────────────────

func handleLogNew(w http.ResponseWriter, r *http.Request) {
	vehicleID := parseInt(r.URL.Query().Get("vehicle_id"))
	if r.Method == http.MethodPost {
		r.ParseForm()
		l := MaintenanceLog{
			VehicleID:  parseInt(r.FormValue("vehicle_id")),
			RecordDate: parseDate(r.FormValue("record_date")),
			Mileage:    parseInt(r.FormValue("mileage")),
			WorkDone:   r.FormValue("work_done"),
			PartsUsed:  r.FormValue("parts_used"),
			Notes:      r.FormValue("notes"),
		}
		_, err := createLog(db, l)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/vehicles/%d", l.VehicleID), http.StatusSeeOther)
		return
	}
	var v Vehicle
	if vehicleID > 0 {
		v, _ = getVehicle(db, vehicleID)
	}
	render(w, "log_form.html", map[string]any{
		"Title":     "New Log Entry",
		"Log":       MaintenanceLog{VehicleID: vehicleID, RecordDate: time.Now()},
		"Vehicle":   v,
		"TodayDate": time.Now().Format("2006-01-02"),
	})
}

func handleLogDetail(w http.ResponseWriter, r *http.Request) {
	segs := pathSegments(r, "/logs/")
	if len(segs) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id := parseInt(segs[0])
	action := ""
	if len(segs) > 1 {
		action = segs[1]
	}

	switch action {
	case "edit":
		l, err := getLog(db, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		v, _ := getVehicle(db, l.VehicleID)
		if r.Method == http.MethodPost {
			r.ParseForm()
			l.RecordDate = parseDate(r.FormValue("record_date"))
			l.Mileage = parseInt(r.FormValue("mileage"))
			l.WorkDone = r.FormValue("work_done")
			l.PartsUsed = r.FormValue("parts_used")
			l.Notes = r.FormValue("notes")
			updateLog(db, l)
			http.Redirect(w, r, fmt.Sprintf("/vehicles/%d", l.VehicleID), http.StatusSeeOther)
			return
		}
		render(w, "log_form.html", map[string]any{
			"Title":   "Edit Log Entry",
			"Log":     l,
			"Vehicle": v,
		})

	case "delete":
		if r.Method == http.MethodPost {
			l, _ := getLog(db, id)
			deleteLog(db, id)
			http.Redirect(w, r, fmt.Sprintf("/vehicles/%d", l.VehicleID), http.StatusSeeOther)
		}
	}
}

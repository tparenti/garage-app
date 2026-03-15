package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ── Models ────────────────────────────────────────────────────────────────────

type Owner struct {
	ID        int
	Name      string
	Phone     string
	Email     string
	Notes     string
	CreatedAt time.Time
	Vehicles  []Vehicle
}

type Vehicle struct {
	ID        int
	OwnerID   int
	OwnerName string
	Year      string
	Make      string
	Model     string
	VIN       string
	LicPlate  string
	Color     string
	Notes     string
	CreatedAt time.Time
	Logs      []MaintenanceLog
}

type MaintenanceLog struct {
	ID          int
	VehicleID   int
	RecordDate  time.Time
	Mileage     int
	WorkDone    string
	PartsUsed   string
	Notes       string
	CreatedAt   time.Time
}

// ── DB Init ───────────────────────────────────────────────────────────────────

func initDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	schema := `
	CREATE TABLE IF NOT EXISTS owners (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL,
		phone      TEXT,
		email      TEXT,
		notes      TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS vehicles (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		owner_id   INTEGER NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
		year       TEXT,
		make       TEXT NOT NULL,
		model      TEXT NOT NULL,
		vin        TEXT,
		lic_plate  TEXT,
		color      TEXT,
		notes      TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS maintenance_logs (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		vehicle_id  INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
		record_date DATE NOT NULL,
		mileage     INTEGER,
		work_done   TEXT NOT NULL,
		parts_used  TEXT,
		notes       TEXT,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err = db.Exec(schema); err != nil {
		log.Fatal(err)
	}
	return db
}

// ── Owner queries ─────────────────────────────────────────────────────────────

func getAllOwners(db *sql.DB) ([]Owner, error) {
	rows, err := db.Query(`SELECT id, name, phone, email, notes, created_at FROM owners ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var owners []Owner
	for rows.Next() {
		var o Owner
		rows.Scan(&o.ID, &o.Name, &o.Phone, &o.Email, &o.Notes, &o.CreatedAt)
		owners = append(owners, o)
	}
	return owners, nil
}

func getOwner(db *sql.DB, id int) (Owner, error) {
	var o Owner
	err := db.QueryRow(`SELECT id, name, phone, email, notes, created_at FROM owners WHERE id=?`, id).
		Scan(&o.ID, &o.Name, &o.Phone, &o.Email, &o.Notes, &o.CreatedAt)
	return o, err
}

func createOwner(db *sql.DB, o Owner) (int64, error) {
	res, err := db.Exec(`INSERT INTO owners (name,phone,email,notes) VALUES (?,?,?,?)`,
		o.Name, o.Phone, o.Email, o.Notes)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateOwner(db *sql.DB, o Owner) error {
	_, err := db.Exec(`UPDATE owners SET name=?,phone=?,email=?,notes=? WHERE id=?`,
		o.Name, o.Phone, o.Email, o.Notes, o.ID)
	return err
}

func deleteOwner(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM owners WHERE id=?`, id)
	return err
}

// ── Vehicle queries ───────────────────────────────────────────────────────────

func getVehiclesByOwner(db *sql.DB, ownerID int) ([]Vehicle, error) {
	rows, err := db.Query(`SELECT id, owner_id, year, make, model, vin, lic_plate, color, notes, created_at
		FROM vehicles WHERE owner_id=? ORDER BY year DESC, make`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		rows.Scan(&v.ID, &v.OwnerID, &v.Year, &v.Make, &v.Model, &v.VIN, &v.LicPlate, &v.Color, &v.Notes, &v.CreatedAt)
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

func getVehicle(db *sql.DB, id int) (Vehicle, error) {
	var v Vehicle
	err := db.QueryRow(`
		SELECT v.id, v.owner_id, o.name, v.year, v.make, v.model, v.vin, v.lic_plate, v.color, v.notes, v.created_at
		FROM vehicles v JOIN owners o ON o.id=v.owner_id WHERE v.id=?`, id).
		Scan(&v.ID, &v.OwnerID, &v.OwnerName, &v.Year, &v.Make, &v.Model, &v.VIN, &v.LicPlate, &v.Color, &v.Notes, &v.CreatedAt)
	return v, err
}

func createVehicle(db *sql.DB, v Vehicle) (int64, error) {
	res, err := db.Exec(`INSERT INTO vehicles (owner_id,year,make,model,vin,lic_plate,color,notes) VALUES (?,?,?,?,?,?,?,?)`,
		v.OwnerID, v.Year, v.Make, v.Model, v.VIN, v.LicPlate, v.Color, v.Notes)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateVehicle(db *sql.DB, v Vehicle) error {
	_, err := db.Exec(`UPDATE vehicles SET year=?,make=?,model=?,vin=?,lic_plate=?,color=?,notes=? WHERE id=?`,
		v.Year, v.Make, v.Model, v.VIN, v.LicPlate, v.Color, v.Notes, v.ID)
	return err
}

func deleteVehicle(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM vehicles WHERE id=?`, id)
	return err
}

// ── Maintenance log queries ───────────────────────────────────────────────────

func getLogsByVehicle(db *sql.DB, vehicleID int) ([]MaintenanceLog, error) {
	rows, err := db.Query(`SELECT id, vehicle_id, record_date, mileage, work_done, parts_used, notes, created_at
		FROM maintenance_logs WHERE vehicle_id=? ORDER BY record_date DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []MaintenanceLog
	for rows.Next() {
		var l MaintenanceLog
		rows.Scan(&l.ID, &l.VehicleID, &l.RecordDate, &l.Mileage, &l.WorkDone, &l.PartsUsed, &l.Notes, &l.CreatedAt)
		logs = append(logs, l)
	}
	return logs, nil
}

func getLog(db *sql.DB, id int) (MaintenanceLog, error) {
	var l MaintenanceLog
	err := db.QueryRow(`SELECT id, vehicle_id, record_date, mileage, work_done, parts_used, notes, created_at
		FROM maintenance_logs WHERE id=?`, id).
		Scan(&l.ID, &l.VehicleID, &l.RecordDate, &l.Mileage, &l.WorkDone, &l.PartsUsed, &l.Notes, &l.CreatedAt)
	return l, err
}

func createLog(db *sql.DB, l MaintenanceLog) (int64, error) {
	res, err := db.Exec(`INSERT INTO maintenance_logs (vehicle_id,record_date,mileage,work_done,parts_used,notes) VALUES (?,?,?,?,?,?)`,
		l.VehicleID, l.RecordDate.Format("2006-01-02"), l.Mileage, l.WorkDone, l.PartsUsed, l.Notes)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateLog(db *sql.DB, l MaintenanceLog) error {
	_, err := db.Exec(`UPDATE maintenance_logs SET record_date=?,mileage=?,work_done=?,parts_used=?,notes=? WHERE id=?`,
		l.RecordDate.Format("2006-01-02"), l.Mileage, l.WorkDone, l.PartsUsed, l.Notes, l.ID)
	return err
}

func deleteLog(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM maintenance_logs WHERE id=?`, id)
	return err
}

func getRecentLogs(db *sql.DB, limit int) ([]struct {
	Log       MaintenanceLog
	Vehicle   Vehicle
	OwnerName string
}, error) {
	rows, err := db.Query(`
		SELECT ml.id, ml.vehicle_id, ml.record_date, ml.mileage, ml.work_done, ml.parts_used, ml.notes, ml.created_at,
		       v.year, v.make, v.model, o.name
		FROM maintenance_logs ml
		JOIN vehicles v ON v.id = ml.vehicle_id
		JOIN owners o ON o.id = v.owner_id
		ORDER BY ml.record_date DESC, ml.created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []struct {
		Log       MaintenanceLog
		Vehicle   Vehicle
		OwnerName string
	}
	for rows.Next() {
		var r struct {
			Log       MaintenanceLog
			Vehicle   Vehicle
			OwnerName string
		}
		rows.Scan(&r.Log.ID, &r.Log.VehicleID, &r.Log.RecordDate, &r.Log.Mileage,
			&r.Log.WorkDone, &r.Log.PartsUsed, &r.Log.Notes, &r.Log.CreatedAt,
			&r.Vehicle.Year, &r.Vehicle.Make, &r.Vehicle.Model, &r.OwnerName)
		results = append(results, r)
	}
	return results, nil
}

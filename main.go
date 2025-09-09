package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

type License struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Vendor       string    `json:"vendor"`
	LicenseType  string    `json:"license_type"`
	ExpiryDate   time.Time `json:"expiry_date"`
	RenewalDate  time.Time `json:"renewal_date"`
	Cost         float64   `json:"cost"`
	Status       string    `json:"status"`
	AssignedTo   string    `json:"assigned_to"`
	CreatedAt    time.Time `json:"created_at"`
}

type Asset struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	AssetType    string    `json:"asset_type"`
	SerialNumber string    `json:"serial_number"`
	Location     string    `json:"location"`
	AssignedTo   string    `json:"assigned_to"`
	PurchaseDate time.Time `json:"purchase_date"`
	WarrantyEnd  time.Time `json:"warranty_end"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Event struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
	EventType   string    `json:"event_type"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type RiskItem struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Probability string    `json:"probability"`
	Impact      string    `json:"impact"`
	RiskLevel   string    `json:"risk_level"`
	Mitigation  string    `json:"mitigation"`
	Owner       string    `json:"owner"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type FOIRequest struct {
	ID          int       `json:"id"`
	RequestID   string    `json:"request_id"`
	Requester   string    `json:"requester"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	DueDate     time.Time `json:"due_date"`
	AssignedTo  string    `json:"assigned_to"`
	CreatedAt   time.Time `json:"created_at"`
}

type App struct {
	db *sql.DB
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal("Failed to initialize app:", err)
	}
	defer app.db.Close()

	func (app *App) calendarHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		// Get current month and year
		now := time.Now()
		year, month, _ := now.Date()
		
		// Get events for current month
		startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, -1)

		rows, err := app.db.Query(`
			SELECT title, description, event_date, event_type, priority 
			FROM events 
			WHERE event_date BETWEEN ? AND ? 
			ORDER BY event_date ASC`, 
			startOfMonth, endOfMonth)
		if err != nil {
			log.Printf("Error querying events: %v", err)
		}
		defer rows.Close()

		var events []Event
		for rows.Next() {
			var event Event
			rows.Scan(&event.Title, &event.Description, &event.EventDate, &event.EventType, &event.Priority)
			events = append(events, event)
		}

		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SL&AM Calendar</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { display: flex; gap: 2rem; }
        .nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; transition: background 0.3s; }
        .nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 2rem; }
        .calendar-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; }
        .event-list { display: grid; gap: 1rem; }
        .event-item { padding: 1.5rem; border-left: 4px solid #667eea; background: #f8f9ff; border-radius: 0 5px 5px 0; }
        .event-title { font-weight: bold; color: #333; font-size: 1.1rem; }
        .event-date { color: #666; margin: 0.5rem 0; }
        .event-type { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.8rem; color: white; margin-right: 0.5rem; }
        .type-renewal { background: #3498db; }
        .type-expiry { background: #e74c3c; }
        .type-audit { background: #f39c12; }
        .priority-high { border-left-color: #e74c3c; }
        .priority-critical { border-left-color: #c0392b; }
        .priority-medium { border-left-color: #f39c12; }
        .placeholder { text-align: center; padding: 3rem; color: #666; font-style: italic; }
        .logout { background: rgba(255,255,255,0.2); color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Calendar - Events & Renewals</h1>
        <div class="nav">
            <a href="/dashboard">Dashboard</a>
            <a href="/calendar" class="active">Calendar</a>
            <a href="/risks">Risk Register</a>
            <a href="/reports">Reports</a>
            <a href="/foi">FOI Requests</a>
            <a href="/licenses">Licenses</a>
            <a href="/assets">Assets</a>
            <a href="/settings">⚙️ Settings</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
    </div>

    <div class="container">
        <div class="card">
            <div class="calendar-header">
                <h2>{{.MonthYear}} Events</h2>
                <div class="placeholder" style="font-size: 0.9rem;">📅 Full calendar view - Coming Soon</div>
            </div>
            
            <div class="event-list">
                {{range .Events}}
                <div class="event-item priority-{{.Priority}}">
                    <div class="event-title">{{.Title}}</div>
                    <div class="event-date">📅 {{.EventDate.Format "January 2, 2006"}}</div>
                    <div>
                        <span class="event-type type-{{.EventType}}">{{.EventType}}</span>
                        <span style="color: #666;">Priority: {{.Priority}}</span>
                    </div>
                    <div style="margin-top: 0.5rem; color: #555;">{{.Description}}</div>
                </div>
                {{else}}
                <div class="placeholder">No events scheduled for this month</div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>`

		t, _ := template.New("calendar").Parse(tmpl)
		data := map[string]interface{}{
			"Events":    events,
			"MonthYear": now.Format("January 2006"),
		}
		t.Execute(w, data)
	})(w, r)
}

func (app *App) riskHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		rows, err := app.db.Query(`
			SELECT id, title, description, probability, impact, risk_level, mitigation, owner, status, created_at
			FROM risk_register 
			ORDER BY 
				CASE risk_level 
					WHEN 'Critical' THEN 1 
					WHEN 'High' THEN 2 
					WHEN 'Medium' THEN 3 
					WHEN 'Low' THEN 4 
				END, created_at DESC`)
		if err != nil {
			log.Printf("Error querying risks: %v", err)
		}
		defer rows.Close()

		var risks []RiskItem
		for rows.Next() {
			var risk RiskItem
			rows.Scan(&risk.ID, &risk.Title, &risk.Description, &risk.Probability, 
					 &risk.Impact, &risk.RiskLevel, &risk.Mitigation, &risk.Owner, 
					 &risk.Status, &risk.CreatedAt)
			risks = append(risks, risk)
		}

		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SL&AM Risk Register</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { display: flex; gap: 2rem; }
        .nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; transition: background 0.3s; }
        .nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 2rem; }
        .risk-item { padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px; border-left: 4px solid; }
        .risk-critical { border-left-color: #c0392b; background: #fdf2f2; }
        .risk-high { border-left-color: #e74c3c; background: #fef5f5; }
        .risk-medium { border-left-color: #f39c12; background: #fefcf3; }
        .risk-low { border-left-color: #27ae60; background: #f8fffe; }
        .risk-title { font-weight: bold; font-size: 1.1rem; margin-bottom: 0.5rem; }
        .risk-meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 1rem; margin: 1rem 0; }
        .risk-badge { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.8rem; color: white; }
        .badge-critical { background: #c0392b; }
        .badge-high { background: #e74c3c; }
        .badge-medium { background: #f39c12; }
        .badge-low { background: #27ae60; }
        .placeholder { text-align: center; padding: 3rem; color: #666; font-style: italic; }
        .logout { background: rgba(255,255,255,0.2); color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Risk Register</h1>
        <div class="nav">
            <a href="/dashboard">Dashboard</a>
            <a href="/calendar">Calendar</a>
            <a href="/risks" class="active">Risk Register</a>
            <a href="/reports">Reports</a>
            <a href="/foi">FOI Requests</a>
            <a href="/licenses">Licenses</a>
            <a href="/assets">Assets</a>
            <a href="/settings">⚙️ Settings</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
    </div>

    <div class="container">
        <div class="card">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
                <h2>Risk Assessment & Management</h2>
                <div class="placeholder" style="font-size: 0.9rem;">➕ Add New Risk - Coming Soon</div>
            </div>
            
            {{range .Risks}}
            <div class="risk-item risk-{{.RiskLevel | ToLower}}">
                <div class="risk-title">{{.Title}}</div>
                <div style="color: #666; margin-bottom: 1rem;">{{.Description}}</div>
                
                <div class="risk-meta">
                    <div><strong>Probability:</strong> {{.Probability}}</div>
                    <div><strong>Impact:</strong> {{.Impact}}</div>
                    <div><strong>Risk Level:</strong> <span class="risk-badge badge-{{.RiskLevel | ToLower}}">{{.RiskLevel}}</span></div>
                    <div><strong>Owner:</strong> {{.Owner}}</div>
                    <div><strong>Status:</strong> {{.Status}}</div>
                </div>
                
                {{if .Mitigation}}
                <div style="margin-top: 1rem;">
                    <strong>Mitigation:</strong> {{.Mitigation}}
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="placeholder">No risks registered</div>
            {{end}}
        </div>
    </div>
</body>
</html>`

		t, _ := template.New("risks").Funcs(template.FuncMap{
			"ToLower": strings.ToLower,
		}).Parse(tmpl)
		data := map[string]interface{}{
			"Risks": risks,
		}
		t.Execute(w, data)
	})(w, r)
}

func (app *App) reportsHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SL&AM Reports</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { display: flex; gap: 2rem; }
        .nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; transition: background 0.3s; }
        .nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 2rem; }
        .reports-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 2rem; }
        .report-card { padding: 2rem; border: 2px solid #e9ecef; border-radius: 10px; transition: border-color 0.3s; cursor: pointer; }
        .report-card:hover { border-color: #667eea; }
        .report-icon { font-size: 2.5rem; margin-bottom: 1rem; }
        .report-title { font-size: 1.2rem; font-weight: bold; margin-bottom: 0.5rem; color: #333; }
        .report-description { color: #666; margin-bottom: 1rem; }
        .btn { padding: 0.7rem 1.5rem; background: #667eea; color: white; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; display: inline-block; transition: background 0.3s; }
        .btn:hover { background: #5a6fd8; }
        .placeholder { text-align: center; padding: 1rem; color: #667eea; font-weight: 500; }
        .logout { background: rgba(255,255,255,0.2); color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Report Execution</h1>
        <div class="nav">
            <a href="/dashboard">Dashboard</a>
            <a href="/calendar">Calendar</a>
            <a href="/risks">Risk Register</a>
            <a href="/reports" class="active">Reports</a>
            <a href="/foi">FOI Requests</a>
            <a href="/licenses">Licenses</a>
            <a href="/assets">Assets</a>
            <a href="/settings">⚙️ Settings</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
    </div>

    <div class="container">
        <div class="card">
            <h2 style="margin-bottom: 2rem;">Available Reports</h2>
            
            <div class="reports-grid">
                <div class="report-card">
                    <div class="report-icon">📊</div>
                    <div class="report-title">License Compliance Report</div>
                    <div class="report-description">Comprehensive overview of software license status, compliance, and renewals</div>
                    <div class="placeholder">Module Coming Soon</div>
                </div>

                <div class="report-card">
                    <div class="report-icon">💰</div>
                    <div class="report-title">Cost Analysis Report</div>
                    <div class="report-description">Detailed breakdown of software licensing costs and budget allocation</div>
                    <div class="placeholder">Module Coming Soon</div>
                </div>

                <div class="report-card">
                    <div class="report-icon">📦</div>
                    <div class="report-title">Asset Inventory Report</div>
                    <div class="report-description">Complete inventory of hardware and software assets with status tracking</div>
                    <div class="placeholder">Module Coming Soon</div>
                </div>

                <div class="report-card">
                    <div class="report-icon">⚠️</div>
                    <div class="report-title">Risk Assessment Report</div>
                    <div class="report-description">Analysis of identified risks, their impact, and mitigation strategies</div>
                    <div class="placeholder">Module Coming Soon</div>
                </div>

                <div class="report-card">
                    <div class="report-icon">📅</div>
                    <div class="report-title">Renewal Calendar Report</div>
                    <div class="report-description">Upcoming renewals, expirations, and critical dates for planning</div>
                    <div class="placeholder">Module Coming Soon</div>
                </div>

                <div class="report-card">
                    <div class="report-icon">🔍</div>
                    <div class="report-title">Audit Trail Report</div>
                    <div class="report-description">Comprehensive audit trail of system activities and changes</div>
                    <div class="placeholder">Module Coming Soon</div>
                </div>
            </div>
        </div>

        <div class="card">
            <h3>Custom Report Builder</h3>
            <p style="color: #666; margin-bottom: 2rem;">Create custom reports with specific filters and data points</p>
            <div class="placeholder">Advanced Report Builder - Coming Soon</div>
        </div>
    </div>
</body>
</html>`

		t, _ := template.New("reports").Parse(tmpl)
		t.Execute(w, nil)
	})(w, r)
}

func (app *App) foiHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		rows, err := app.db.Query(`
			SELECT id, request_id, requester, subject, description, status, due_date, assigned_to, created_at
			FROM foi_requests 
			ORDER BY created_at DESC`)
		if err != nil {
			log.Printf("Error querying FOI requests: %v", err)
		}
		defer rows.Close()

		var requests []FOIRequest
		for rows.Next() {
			var req FOIRequest
			var assignedTo sql.NullString
			rows.Scan(&req.ID, &req.RequestID, &req.Requester, &req.Subject, 
					 &req.Description, &req.Status, &req.DueDate, &assignedTo, &req.CreatedAt)
			if assignedTo.Valid {
				req.AssignedTo = assignedTo.String
			}
			requests = append(requests, req)
		}

		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SL&AM FOI Requests</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { display: flex; gap: 2rem; }
        .nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; transition: background 0.3s; }
        .nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 2rem; }
        .foi-item { padding: 1.5rem; margin-bottom: 1rem; border: 1px solid #e9ecef; border-radius: 8px; }
        .foi-header { display: flex; justify-content: space-between; align-items: start; margin-bottom: 1rem; }
        .foi-id { font-weight: bold; color: #667eea; }
        .status-badge { padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.8rem; color: white; }
        .status-received { background: #3498db; }
        .status-processing { background: #f39c12; }
        .status-completed { background: #27ae60; }
        .status-overdue { background: #e74c3c; }
        .foi-meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin: 1rem 0; color: #666; font-size: 0.9rem; }
        .placeholder { text-align: center; padding: 3rem; color: #666; font-style: italic; }
        .logout { background: rgba(255,255,255,0.2); color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Freedom of Information Requests</h1>
        <div class="nav">
            <a href="/dashboard">Dashboard</a>
            <a href="/calendar">Calendar</a>
            <a href="/risks">Risk Register</a>
            <a href="/reports">Reports</a>
            <a href="/foi" class="active">FOI Requests</a>
            <a href="/licenses">Licenses</a>
            <a href="/assets">Assets</a>
            <a href="/settings">⚙️ Settings</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
    </div>

    <div class="container">
        <div class="card">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
                <h2>FOI Request Management</h2>
                <div class="placeholder" style="font-size: 0.9rem;">➕ New Request - Coming Soon</div>
            </div>
            
            {{range .Requests}}
            <div class="foi-item">
                <div class="foi-header">
                    <div class="foi-id">{{.RequestID}}</div>
                    <span class="status-badge status-{{.Status}}">{{.Status | ToUpper}}</span>
                </div>
                
                <div style="font-weight: bold; margin-bottom: 0.5rem; color: #333;">{{.Subject}}</div>
                <div style="color: #666; margin-bottom: 1rem;">{{.Description}}</div>
                
                <div class="foi-meta">
                    <div><strong>Requester:</strong> {{.Requester}}</div>
                    <div><strong>Due Date:</strong> {{.DueDate.Format "Jan 2, 2006"}}</div>
                    <div><strong>Assigned To:</strong> {{if .AssignedTo}}{{.AssignedTo}}{{else}}Unassigned{{end}}</div>
                    <div><strong>Created:</strong> {{.CreatedAt.Format "Jan 2, 2006"}}</div>
                </div>
            </div>
            {{else}}
            <div class="placeholder">No FOI requests found</div>
            {{end}}
        </div>
    </div>
</body>
</html>`

		t, _ := template.New("foi").Funcs(template.FuncMap{
			"ToUpper": strings.ToUpper,
		}).Parse(tmpl)
		data := map[string]interface{}{
			"Requests": requests,
		}
		t.Execute(w, data)
	})(w, r)
}

func (app *App) settingsHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		tmpl := `<!DOCTYPE html>
<html><head><title>SL&AM Settings</title><style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Segoe UI', sans-serif; background: #f5f5f5; }
.header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
.nav { display: flex; gap: 2rem; }
.nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; }
.nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
.container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
.card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
.placeholder { text-align: center; padding: 3rem; color: #667eea; font-style: italic; }
.logout { background: rgba(255,255,255,0.2); }
</style></head><body>
<div class="header"><h1>Settings</h1><div class="nav">
<a href="/dashboard">Dashboard</a><a href="/calendar">Calendar</a><a href="/risks">Risk Register</a>
<a href="/reports">Reports</a><a href="/foi">FOI Requests</a><a href="/licenses">Licenses</a>
<a href="/assets">Assets</a><a href="/settings" class="active">⚙️ Settings</a><a href="/logout" class="logout">Logout</a>
</div></div>
<div class="container"><div class="card">
<h2>System Settings</h2><div class="placeholder">Settings modules coming soon</div>
</div></div></body></html>`
		t, _ := template.New("settings").Parse(tmpl)
		t.Execute(w, nil)
	})(w, r)
}

func (app *App) licensesHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		tmpl := `<!DOCTYPE html>
<html><head><title>SL&AM Licenses</title><style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Segoe UI', sans-serif; background: #f5f5f5; }
.header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
.nav { display: flex; gap: 2rem; }
.nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; }
.nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
.container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
.card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
.placeholder { text-align: center; padding: 3rem; color: #667eea; font-style: italic; }
.logout { background: rgba(255,255,255,0.2); }
</style></head><body>
<div class="header"><h1>License Management</h1><div class="nav">
<a href="/dashboard">Dashboard</a><a href="/calendar">Calendar</a><a href="/risks">Risk Register</a>
<a href="/reports">Reports</a><a href="/foi">FOI Requests</a><a href="/licenses" class="active">Licenses</a>
<a href="/assets">Assets</a><a href="/settings">⚙️ Settings</a><a href="/logout" class="logout">Logout</a>
</div></div>
<div class="container"><div class="card">
<h2>Software License Management</h2><div class="placeholder">License management module coming soon</div>
</div></div></body></html>`
		t, _ := template.New("licenses").Parse(tmpl)
		t.Execute(w, nil)
	})(w, r)
}

func (app *App) assetsHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		tmpl := `<!DOCTYPE html>
<html><head><title>SL&AM Assets</title><style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Segoe UI', sans-serif; background: #f5f5f5; }
.header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
.nav { display: flex; gap: 2rem; }
.nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; }
.nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
.container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
.card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
.placeholder { text-align: center; padding: 3rem; color: #667eea; font-style: italic; }
.logout { background: rgba(255,255,255,0.2); }
</style></head><body>
<div class="header"><h1>Asset Management</h1><div class="nav">
<a href="/dashboard">Dashboard</a><a href="/calendar">Calendar</a><a href="/risks">Risk Register</a>
<a href="/reports">Reports</a><a href="/foi">FOI Requests</a><a href="/licenses">Licenses</a>
<a href="/assets" class="active">Assets</a><a href="/settings">⚙️ Settings</a><a href="/logout" class="logout">Logout</a>
</div></div>
<div class="container"><div class="card">
<h2>Hardware & Software Asset Management</h2><div class="placeholder">Asset management module coming soon</div>
</div></div></body></html>`
		t, _ := template.New("assets").Parse(tmpl)
		t.Execute(w, nil)
	})(w, r)
}

func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		app.db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session_id", Value: "", HttpOnly: true, Path: "/", MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
                 Initialize database tables
	if err := app.initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Setup routes
	http.HandleFunc("/", app.loginHandler)
	http.HandleFunc("/login", app.loginHandler)
	http.HandleFunc("/authenticate", app.authenticateHandler)
	http.HandleFunc("/dashboard", app.dashboardHandler)
	http.HandleFunc("/calendar", app.calendarHandler)
	http.HandleFunc("/risks", app.riskHandler)
	http.HandleFunc("/reports", app.reportsHandler)
	http.HandleFunc("/foi", app.foiHandler)
	http.HandleFunc("/settings", app.settingsHandler)
	http.HandleFunc("/licenses", app.licensesHandler)
	http.HandleFunc("/assets", app.assetsHandler)
	http.HandleFunc("/logout", app.logoutHandler)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	log.Println("SL&AM Management System starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func NewApp() (*App, error) {
	db, err := sql.Open("sqlite", "slam.db")
	if err != nil {
		return nil, err
	}

	return &App{db: db}, nil
}

func (app *App) initDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS licenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			vendor TEXT NOT NULL,
			license_type TEXT NOT NULL,
			expiry_date DATE NOT NULL,
			renewal_date DATE,
			cost DECIMAL(10,2),
			status TEXT DEFAULT 'active',
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			asset_type TEXT NOT NULL,
			serial_number TEXT,
			location TEXT,
			assigned_to TEXT,
			purchase_date DATE,
			warranty_end DATE,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			event_date DATE NOT NULL,
			event_type TEXT NOT NULL,
			priority TEXT DEFAULT 'medium',
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS risk_register (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			probability TEXT NOT NULL,
			impact TEXT NOT NULL,
			risk_level TEXT NOT NULL,
			mitigation TEXT,
			owner TEXT,
			status TEXT DEFAULT 'open',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS foi_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request_id TEXT UNIQUE NOT NULL,
			requester TEXT NOT NULL,
			subject TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'received',
			due_date DATE NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
	}

	for _, query := range queries {
		if _, err := app.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	}

	// Create default admin user if none exists
	var count int
	app.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		app.db.Exec("INSERT INTO users (username, email, password_hash, role) VALUES (?, ?, ?, ?)",
			"admin", "admin@slam.local", string(hashedPassword), "admin")
		log.Println("Created default admin user: admin/admin123")
	}

	// Insert sample data
	app.insertSampleData()

	return nil
}

func (app *App) insertSampleData() {
	// Sample licenses
	licenses := []License{
		{Name: "Microsoft Office 365", Vendor: "Microsoft", LicenseType: "Subscription", ExpiryDate: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), Cost: 15000.00, Status: "active"},
		{Name: "Adobe Creative Suite", Vendor: "Adobe", LicenseType: "Enterprise", ExpiryDate: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC), Cost: 8000.00, Status: "active"},
		{Name: "AutoCAD", Vendor: "Autodesk", LicenseType: "Named User", ExpiryDate: time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), Cost: 4500.00, Status: "expiring"},
	}

	for _, license := range licenses {
		app.db.Exec("INSERT OR IGNORE INTO licenses (name, vendor, license_type, expiry_date, cost, status) VALUES (?, ?, ?, ?, ?, ?)",
			license.Name, license.Vendor, license.LicenseType, license.ExpiryDate, license.Cost, license.Status)
	}

	// Sample events
	events := []Event{
		{Title: "Office 365 Renewal", Description: "Annual renewal due", EventDate: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), EventType: "renewal", Priority: "high"},
		{Title: "AutoCAD License Expiry", Description: "License expires soon", EventDate: time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), EventType: "expiry", Priority: "critical"},
		{Title: "Asset Audit", Description: "Quarterly asset verification", EventDate: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), EventType: "audit", Priority: "medium"},
	}

	for _, event := range events {
		app.db.Exec("INSERT OR IGNORE INTO events (title, description, event_date, event_type, priority) VALUES (?, ?, ?, ?, ?)",
			event.Title, event.Description, event.EventDate, event.EventType, event.Priority)
	}

	// Sample risk items
	risks := []RiskItem{
		{Title: "License Compliance Risk", Description: "Risk of non-compliance with software licenses", Probability: "Medium", Impact: "High", RiskLevel: "High", Mitigation: "Regular license audits and monitoring", Owner: "IT Manager"},
		{Title: "Asset Loss Risk", Description: "Risk of hardware theft or loss", Probability: "Low", Impact: "Medium", RiskLevel: "Low", Mitigation: "Asset tracking and security measures", Owner: "Security Officer"},
	}

	for _, risk := range risks {
		app.db.Exec("INSERT OR IGNORE INTO risk_register (title, description, probability, impact, risk_level, mitigation, owner) VALUES (?, ?, ?, ?, ?, ?, ?)",
			risk.Title, risk.Description, risk.Probability, risk.Impact, risk.RiskLevel, risk.Mitigation, risk.Owner)
	}
}

func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SL&AM Login</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); height: 100vh; display: flex; align-items: center; justify-content: center; }
        .login-container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 15px 35px rgba(0,0,0,0.1); width: 100%; max-width: 400px; }
        .logo { text-align: center; margin-bottom: 30px; }
        .logo h1 { color: #333; font-size: 28px; margin-bottom: 5px; }
        .logo p { color: #666; font-size: 14px; }
        .form-group { margin-bottom: 20px; }
        .form-group label { display: block; margin-bottom: 5px; color: #333; font-weight: 500; }
        .form-group input, .form-group select { width: 100%; padding: 12px; border: 2px solid #ddd; border-radius: 5px; font-size: 16px; transition: border-color 0.3s; }
        .form-group input:focus, .form-group select:focus { outline: none; border-color: #667eea; }
        .btn { width: 100%; padding: 12px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; border-radius: 5px; font-size: 16px; cursor: pointer; transition: opacity 0.3s; }
        .btn:hover { opacity: 0.9; }
        .alert { padding: 10px; margin-bottom: 20px; border-radius: 5px; background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">
            <h1>SL&AM</h1>
            <p>Software License & Asset Management</p>
        </div>
        {{if .Error}}<div class="alert">{{.Error}}</div>{{end}}
        <form method="POST" action="/authenticate">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" required>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            <div class="form-group">
                <label for="role">Access Group</label>
                <select id="role" name="role" required>
                    <option value="">Select Access Group</option>
                    <option value="admin">Administrator</option>
                    <option value="manager">License Manager</option>
                    <option value="analyst">Asset Analyst</option>
                    <option value="viewer">Viewer</option>
                </select>
            </div>
            <button type="submit" class="btn">Login</button>
        </form>
    </div>
</body>
</html>`
		
		t, _ := template.New("login").Parse(tmpl)
		t.Execute(w, nil)
	}
}

func (app *App) authenticateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	requestedRole := r.FormValue("role")

	var user User
	var passwordHash string
	err := app.db.QueryRow("SELECT id, username, email, password_hash, role, active FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Email, &passwordHash, &user.Role, &user.Active)

	if err != nil || !user.Active {
		app.renderLoginWithError(w, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		app.renderLoginWithError(w, "Invalid credentials")
		return
	}

	// Check if user has permission for requested role
	if !app.hasRolePermission(user.Role, requestedRole) {
		app.renderLoginWithError(w, "Access denied for selected group")
		return
	}

	// Create session
	sessionID, err := app.createSession(user.ID)
	if err != nil {
		app.renderLoginWithError(w, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400, // 24 hours
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *App) renderLoginWithError(w http.ResponseWriter, errorMsg string) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SL&AM Login</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); height: 100vh; display: flex; align-items: center; justify-content: center; }
        .login-container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 15px 35px rgba(0,0,0,0.1); width: 100%; max-width: 400px; }
        .logo { text-align: center; margin-bottom: 30px; }
        .logo h1 { color: #333; font-size: 28px; margin-bottom: 5px; }
        .logo p { color: #666; font-size: 14px; }
        .form-group { margin-bottom: 20px; }
        .form-group label { display: block; margin-bottom: 5px; color: #333; font-weight: 500; }
        .form-group input, .form-group select { width: 100%; padding: 12px; border: 2px solid #ddd; border-radius: 5px; font-size: 16px; transition: border-color 0.3s; }
        .form-group input:focus, .form-group select:focus { outline: none; border-color: #667eea; }
        .btn { width: 100%; padding: 12px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; border-radius: 5px; font-size: 16px; cursor: pointer; transition: opacity 0.3s; }
        .btn:hover { opacity: 0.9; }
        .alert { padding: 10px; margin-bottom: 20px; border-radius: 5px; background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">
            <h1>SL&AM</h1>
            <p>Software License & Asset Management</p>
        </div>
        <div class="alert">{{.Error}}</div>
        <form method="POST" action="/authenticate">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" required>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            <div class="form-group">
                <label for="role">Access Group</label>
                <select id="role" name="role" required>
                    <option value="">Select Access Group</option>
                    <option value="admin">Administrator</option>
                    <option value="manager">License Manager</option>
                    <option value="analyst">Asset Analyst</option>
                    <option value="viewer">Viewer</option>
                </select>
            </div>
            <button type="submit" class="btn">Login</button>
        </form>
    </div>
</body>
</html>`
	
	t, _ := template.New("login").Parse(tmpl)
	t.Execute(w, map[string]string{"Error": errorMsg})
}

func (app *App) hasRolePermission(userRole, requestedRole string) bool {
	roleHierarchy := map[string]int{
		"admin":   4,
		"manager": 3,
		"analyst": 2,
		"viewer":  1,
	}

	userLevel := roleHierarchy[userRole]
	requestedLevel := roleHierarchy[requestedRole]

	return userLevel >= requestedLevel
}

func (app *App) createSession(userID int) (string, error) {
	sessionID := make([]byte, 32)
	_, err := rand.Read(sessionID)
	if err != nil {
		return "", err
	}

	sessionIDStr := hex.EncodeToString(sessionID)
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = app.db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionIDStr, userID, expiresAt)
	if err != nil {
		return "", err
	}

	return sessionIDStr, nil
}

func (app *App) requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var userID int
		var expiresAt time.Time
		err = app.db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", cookie.Value).
			Scan(&userID, &expiresAt)

		if err != nil || time.Now().After(expiresAt) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Store user ID in context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	}
}

func (app *App) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	app.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		// Get current month events
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, -1)

		rows, err := app.db.Query(`
			SELECT title, description, event_date, event_type, priority 
			FROM events 
			WHERE event_date BETWEEN ? AND ? 
			ORDER BY event_date ASC`, 
			startOfMonth, endOfMonth)
		if err != nil {
			log.Printf("Error querying events: %v", err)
		}
		defer rows.Close()

		var events []Event
		for rows.Next() {
			var event Event
			rows.Scan(&event.Title, &event.Description, &event.EventDate, &event.EventType, &event.Priority)
			events = append(events, event)
		}

		// Get license statistics
		var totalLicenses, expiringLicenses int
		app.db.QueryRow("SELECT COUNT(*) FROM licenses").Scan(&totalLicenses)
		app.db.QueryRow("SELECT COUNT(*) FROM licenses WHERE expiry_date BETWEEN ? AND ?", 
			now, now.AddDate(0, 3, 0)).Scan(&expiringLicenses)

		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SL&AM Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { display: flex; gap: 2rem; }
        .nav a { color: white; text-decoration: none; padding: 0.5rem 1rem; border-radius: 5px; transition: background 0.3s; }
        .nav a:hover, .nav a.active { background: rgba(255,255,255,0.2); }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .dashboard-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 2rem; margin-bottom: 2rem; }
        .card { background: white; padding: 2rem; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .card h3 { margin-bottom: 1rem; color: #333; }
        .stat-card { text-align: center; }
        .stat-number { font-size: 2.5rem; font-weight: bold; color: #667eea; }
        .stat-label { color: #666; margin-top: 0.5rem; }
        .event-list { max-height: 300px; overflow-y: auto; }
        .event-item { padding: 1rem; border-left: 4px solid #667eea; margin-bottom: 1rem; background: #f8f9ff; border-radius: 0 5px 5px 0; }
        .event-title { font-weight: bold; color: #333; }
        .event-date { color: #666; font-size: 0.9rem; }
        .priority-high { border-left-color: #e74c3c; }
        .priority-critical { border-left-color: #c0392b; }
        .priority-medium { border-left-color: #f39c12; }
        .btn { padding: 0.7rem 1.5rem; background: #667eea; color: white; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; display: inline-block; transition: background 0.3s; }
        .btn:hover { background: #5a6fd8; }
        .logout { background: rgba(255,255,255,0.2); color: white; }
        .quick-actions { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; }
        .quick-action { display: block; padding: 1rem; text-align: center; background: #667eea; color: white; text-decoration: none; border-radius: 5px; transition: background 0.3s; }
        .quick-action:hover { background: #5a6fd8; }
    </style>
</head>
<body>
    <div class="header">
        <h1>SL&AM Dashboard</h1>
        <div class="nav">
            <a href="/dashboard" class="active">Dashboard</a>
            <a href="/calendar">Calendar</a>
            <a href="/risks">Risk Register</a>
            <a href="/reports">Reports</a>
            <a href="/foi">FOI Requests</a>
            <a href="/licenses">Licenses</a>
            <a href="/assets">Assets</a>
            <a href="/settings">⚙️ Settings</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
    </div>

    <div class="container">
        <div class="dashboard-grid">
            <div class="card stat-card">
                <div class="stat-number">{{.TotalLicenses}}</div>
                <div class="stat-label">Total Licenses</div>
            </div>
            <div class="card stat-card">
                <div class="stat-number">{{.ExpiringLicenses}}</div>
                <div class="stat-label">Expiring Soon</div>
            </div>
            <div class="card stat-card">
                <div class="stat-number">{{len .Events}}</div>
                <div class="stat-label">This Month's Events</div>
            </div>
        </div>

        <div class="dashboard-grid">
            <div class="card">
                <h3>Current Month Events</h3>
                <div class="event-list">
                    {{range .Events}}
                    <div class="event-item priority-{{.Priority}}">
                        <div class="event-title">{{.Title}}</div>
                        <div class="event-date">{{.EventDate.Format "Jan 2, 2006"}} - {{.EventType}}</div>
                        <div>{{.Description}}</div>
                    </div>
                    {{end}}
                </div>
            </div>

            <div class="card">
                <h3>Quick Actions</h3>
                <div class="quick-actions">
                    <a href="/calendar" class="quick-action">View Calendar</a>
                    <a href="/risks" class="quick-action">Risk Register</a>
                    <a href="/reports" class="quick-action">Generate Reports</a>
                    <a href="/foi" class="quick-action">FOI Requests</a>
                    <a href="/licenses" class="quick-action">Manage Licenses</a>
                    <a href="/assets" class="quick-action">Manage Assets</a>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`

		t, _ := template.New("dashboard").Parse(tmpl)
		data := map[string]interface{}{
			"Events":           events,
			"TotalLicenses":    totalLicenses,
			"ExpiringLicenses": expiringLicenses,
		}
		t.Execute(w, data)
	})(w, r)
}

//

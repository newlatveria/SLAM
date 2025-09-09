package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

// A global variable to hold our database connection.
var db *sql.DB

// Define structs for data from the database.
type License struct {
	ID         int
	Name       string
	Vendor     string
	ExpiryDate time.Time
}

// A generic struct to hold all the data for the template
type PageData struct {
	TotalAssets      int
	TotalLicenses    int
	ExpiringSoon     int
	UpcomingLicenses []License
	Assets           []Asset
}

const dashboardContent = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Software Licence & Asset Management Dashboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
        body {
            font-family: 'Inter', sans-serif;
            background-color: #f3f4f6;
        }
        .placeholder-page {
            display: none;
        }
        .active-page {
            display: block;
        }
    </style>
</head>
<body class="bg-gray-100 flex items-center justify-center min-h-screen">

    <!-- Dashboard Page -->
    <div id="dashboard-page" class="bg-white p-8 rounded-2xl shadow-xl w-full max-w-5xl transition-transform duration-500 ease-in-out transform scale-100 opacity-100">
        <div class="flex flex-col md:flex-row h-full">
            <!-- Sidebar Navigation -->
            <div class="md:w-1/4 p-4 border-b md:border-b-0 md:border-r border-gray-200">
                <h2 class="text-2xl font-semibold mb-6 text-gray-700">Navigation</h2>
                <ul id="main-nav" class="space-y-4">
                    <li><a href="/" class="block py-2 px-4 rounded-lg text-gray-600 font-medium hover:bg-gray-200 transition-colors duration-200">Dashboard</a></li>
                    <li><a href="/assets" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">Asset Register</a></li>
                    <li><a href="#compliance-audits" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">Compliance Audits</a></li>
                    <li><a href="#license-renewals" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">License Renewals</a></li>
                    <li><a href="#risk-register" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">Risk Register</a></li>
                    <li><a href="#report-execution" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">Report Execution</a></li>
                    <li><a href="#foi-requests" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">FOI Requests</a></li>
                    <li><a href="#settings" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">Settings</a></li>
                </ul>
            </div>
            <!-- Main Content Area -->
            <div class="md:w-3/4 p-6">
                <!-- Home Page -->
                <div id="home-page" class="placeholder-page active-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-6">SL&AM Dashboard</h2>
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        <!-- Upcoming Events Card -->
                        <div class="bg-gray-50 p-6 rounded-2xl shadow-sm border border-gray-200">
                            <h3 class="text-xl font-semibold text-gray-700 mb-4">Upcoming Events</h3>
                            <ul class="list-disc list-inside space-y-2 text-gray-600">
                                {{range .UpcomingLicenses}}
                                <li>{{.Name}} license expires on {{.ExpiryDate.Format "01/02/2006"}}</li>
                                {{else}}
                                <li>No upcoming license expiries.</li>
                                {{end}}
                            </ul>
                        </div>
                        <!-- Key Information Card -->
                        <div class="bg-gray-50 p-6 rounded-2xl shadow-sm border border-gray-200">
                            <h3 class="text-xl font-semibold text-gray-700 mb-4">Key Information</h3>
                            <ul class="list-disc list-inside space-y-2 text-gray-600">
                                <li>**Total Assets:** {{.TotalAssets}}</li>
                                <li>**Total Licenses:** {{.TotalLicenses}}</li>
                                <li>**Expiring Soon:** {{.ExpiringSoon}}</li>
                                <li>**High-Risk Items:** 3</li>
                            </ul>
                        </div>
                        <!-- Quick Actions Card -->
                        <div class="bg-gray-50 p-6 rounded-2xl shadow-sm border border-gray-200">
                            <h3 class="text-xl font-semibold text-gray-700 mb-4">Quick Actions</h3>
                            <ul class="space-y-2">
                                <li><a href="/assets" class="block py-2 px-4 rounded-md text-sm font-medium text-blue-600 bg-blue-100 hover:bg-blue-200 transition-colors duration-200">Add New Asset</a></li>
                                <li><a href="#" class="block py-2 px-4 rounded-md text-sm font-medium text-green-600 bg-green-100 hover:bg-green-200 transition-colors duration-200">Run Compliance Report</a></li>
                                <li><a href="#" class="block py-2 px-4 rounded-md text-sm font-medium text-red-600 bg-red-100 hover:bg-red-200 transition-colors duration-200">Review High-Risk Items</a></li>
                            </ul>
                        </div>
                    </div>
                </div>

                <!-- Asset Register Page -->
                <div id="assets-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">Asset Register</h2>
                    <p class="text-gray-600 mb-6">Add and view a detailed list of all software and hardware assets.</p>

                    <!-- Add New Asset Form -->
                    <div class="bg-gray-50 p-6 rounded-2xl shadow-sm border border-gray-200 mb-6">
                        <h3 class="text-xl font-semibold text-gray-700 mb-4">Add New Asset</h3>
                        <form action="/assets" method="post" class="space-y-4">
                            <div>
                                <label for="name" class="block text-sm font-medium text-gray-700">Asset Name</label>
                                <input type="text" name="name" id="name" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm">
                            </div>
                            <div>
                                <label for="asset-type" class="block text-sm font-medium text-gray-700">Asset Type</label>
                                <input type="text" name="asset-type" id="asset-type" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm">
                            </div>
                            <div>
                                <label for="location" class="block text-sm font-medium text-gray-700">Location</label>
                                <input type="text" name="location" id="location" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm">
                            </div>
                            <button type="submit" class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors duration-300">
                                Save Asset
                            </button>
                        </form>
                    </div>

                    <!-- Assets Table -->
                    <div class="bg-white p-6 rounded-2xl shadow-sm border border-gray-200 overflow-x-auto">
                        <h3 class="text-xl font-semibold text-gray-700 mb-4">Current Assets</h3>
                        <table class="min-w-full divide-y divide-gray-200">
                            <thead class="bg-gray-50">
                                <tr>
                                    <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Asset Name</th>
                                    <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                                    <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Location</th>
                                </tr>
                            </thead>
                            <tbody class="bg-white divide-y divide-gray-200">
                                {{range .Assets}}
                                <tr>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{.Name}}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{.AssetType}}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{.Location}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                </div>

                <!-- Other Placeholder Pages -->
                <div id="compliance-audits-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">Compliance Audits</h2>
                    <p class="text-gray-600">This page will provide tools and reports for compliance audits.</p>
                </div>
                <div id="license-renewals-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">License Renewals</h2>
                    <p class="text-gray-600">This page will show a calendar and list of upcoming license renewals and expiries.</p>
                </div>
                <div id="risk-register-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">Risk Register</h2>
                    <p class="text-gray-600">This page will show a risk register that will outline all key risks, their mitigations, and ownerships.</p>
                </div>
                <div id="report-execution-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">Report Execution</h2>
                    <p class="text-gray-600">This page will allow for the execution and generation of various reports on software and asset data.</p>
                </div>
                <div id="foi-requests-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">Freedom of Information Requests</h2>
                    <p class="text-gray-600">This page will be used to manage and track requests for information related to software and assets.</p>
                </div>
                <div id="settings-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">Database Settings</h2>
                    <p class="text-gray-600 mb-6">Select your preferred database type for future development. This is for demonstration purposes only and does not change the live connection.</p>
                    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                        <div class="p-4 bg-gray-50 rounded-lg border border-gray-200 text-center cursor-pointer hover:bg-blue-100 transition-colors duration-200">
                            <h3 class="font-medium text-lg text-gray-800">PostgreSQL</h3>
                            <p class="text-sm text-gray-500">A powerful, open-source relational database system.</p>
                        </div>
                        <div class="p-4 bg-gray-50 rounded-lg border border-gray-200 text-center cursor-pointer hover:bg-blue-100 transition-colors duration-200">
                            <h3 class="font-medium text-lg text-gray-800">Oracle Database</h3>
                            <p class="text-sm text-gray-500">A widely used enterprise-grade relational database.</p>
                        </div>
                        <div class="p-4 bg-gray-50 rounded-lg border border-gray-200 text-center cursor-pointer hover:bg-blue-100 transition-colors duration-200">
                            <h3 class="font-medium text-lg text-gray-800">MySQL</h3>
                            <p class="text-sm text-gray-500">The world's most popular open source database.</p>
                        </div>
                        <div class="p-4 bg-gray-50 rounded-lg border border-gray-200 text-center cursor-pointer hover:bg-blue-100 transition-colors duration-200">
                            <h3 class="font-medium text-lg text-gray-800">SQL Server</h3>
                            <p class="text-sm text-gray-500">Microsoft's relational database management system.</p>
                        </div>
                        <div class="p-4 bg-gray-50 rounded-lg border border-gray-200 text-center cursor-pointer hover:bg-blue-100 transition-colors duration-200">
                            <h3 class="font-medium text-lg text-gray-800">SQLite</h3>
                            <p class="text-sm text-gray-500">A lightweight, file-based database perfect for local use.</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        document.addEventListener('DOMContentLoaded', () => {
            const navLinks = document.querySelectorAll('#main-nav a');
            const pages = document.querySelectorAll('.placeholder-page');

            const path = window.location.pathname;
            let activePageId = 'home-page';

            if (path.startsWith('/assets')) {
                activePageId = 'assets-page';
            }

            pages.forEach(page => {
                if (page.id === activePageId) {
                    page.classList.add('active-page');
                } else {
                    page.classList.remove('active-page');
                }
            });

            navLinks.forEach(link => {
                const linkPath = new URL(link.href).pathname;
                if (linkPath === path) {
                    link.classList.remove('text-gray-600', 'hover:bg-gray-200');
                    link.classList.add('bg-blue-500', 'text-white', 'hover:bg-blue-600');
                } else {
                    link.classList.remove('bg-blue-500', 'text-white', 'hover:bg-blue-600');
                    link.classList.add('text-gray-600', 'hover:bg-gray-200');
                }
            });
        });
    </script>
</body>
</html>
`

// initDB initializes the SQLite database and creates the necessary tables.
func initDB() {
	dbPath := "./slam.db"
	var err error

	// Open the database connection
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Unable to open database: %v\n", err)
	}

	// Create tables if they don't exist
	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS licenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			vendor TEXT,
			expiry_date DATE,
			renewal_date DATE,
			status TEXT
		);
		CREATE TABLE IF NOT EXISTS assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			asset_type TEXT,
			location TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Error creating tables: %v\n", err)
	}

	log.Println("Database initialized successfully.")
}

// seedDB populates the database with initial data if tables are empty.
func seedDB() {
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM assets").Scan(&count)
	if err != nil || count > 0 {
		return
	}

	// Seed assets
	assetsSQL := `
		INSERT INTO assets (name, asset_type, location) VALUES
		('Dell XPS 15', 'Laptop', 'Office 1'),
		('ThinkPad X1 Carbon', 'Laptop', 'Office 2'),
		('HP ProDesk 400 G7', 'Desktop', 'Office 3');
	`
	_, err = db.ExecContext(context.Background(), assetsSQL)
	if err != nil {
		log.Printf("Error seeding assets: %v\n", err)
	}

	// Seed licenses
	licensesSQL := `
		INSERT INTO licenses (name, vendor, expiry_date, status) VALUES
		('Microsoft Office 365', 'Microsoft', date('now', '+35 days'), 'active'),
		('Adobe Creative Cloud', 'Adobe', date('now', '+20 days'), 'active'),
		('Autodesk AutoCAD', 'Autodesk', date('now', '-5 days'), 'expired');
	`
	_, err = db.ExecContext(context.Background(), licensesSQL)
	if err != nil {
		log.Printf("Error seeding licenses: %v\n", err)
	}

	log.Println("Database seeded with sample data.")
}

// getPageData fetches all necessary data for the dashboard and assets pages.
func getPageData() (*PageData, error) {
	data := &PageData{}

	// Fetch total assets and licenses
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM assets").Scan(&data.TotalAssets)
	if err != nil {
		return nil, fmt.Errorf("error fetching total assets: %w", err)
	}

	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM licenses").Scan(&data.TotalLicenses)
	if err != nil {
		return nil, fmt.Errorf("error fetching total licenses: %w", err)
	}

	// Fetch upcoming licenses
	rows, err := db.QueryContext(context.Background(), "SELECT id, name, vendor, expiry_date FROM licenses WHERE expiry_date BETWEEN date('now') AND date('now', '+30 days')")
	if err != nil {
		return nil, fmt.Errorf("error fetching upcoming licenses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var l License
		var expiryDateStr string
		if err := rows.Scan(&l.ID, &l.Name, &l.Vendor, &expiryDateStr); err != nil {
			return nil, fmt.Errorf("error scanning upcoming license: %w", err)
		}
		l.ExpiryDate, _ = time.Parse("2006-01-02", expiryDateStr)
		data.UpcomingLicenses = append(data.UpcomingLicenses, l)
	}
	data.ExpiringSoon = len(data.UpcomingLicenses)

	// Fetch all assets
	rows, err = db.QueryContext(context.Background(), "SELECT id, name, asset_type, location FROM assets")
	if err != nil {
		return nil, fmt.Errorf("error fetching assets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.ID, &a.Name, &a.AssetType, &a.Location); err != nil {
			return nil, fmt.Errorf("error scanning asset: %w", err)
		}
		data.Assets = append(data.Assets, a)
	}

	return data, nil
}

// renderTemplate is a helper function to parse and execute the embedded template.
func renderTemplate(w http.ResponseWriter, r *http.Request, data interface{}) {
	tmpl, err := template.New("dashboard").Parse(dashboardContent)
	if err != nil {
		log.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// homeHandler serves the main dashboard page with dynamic data.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getPageData()
	if err != nil {
		log.Printf("Error fetching page data: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, data)
}

// startServer sets up and starts the HTTP server.
func startServer() {
	router := mux.NewRouter()

	// Define routes for different pages
	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/assets", assetsHandler).Methods("GET", "POST")

	// A placeholder handler for other routes
	router.HandleFunc("/{page}", func(w http.ResponseWriter, r *http.Request) {
		data, err := getPageData()
		if err != nil {
			log.Printf("Error fetching page data: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		renderTemplate(w, r, data)
	})

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	// Initialize the database before starting the server
	initDB()
	seedDB()
	defer db.Close()

	startServer()
}

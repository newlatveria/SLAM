package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

// A global variable to hold our database connection.
var db *sql.DB

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

    <!-- Login Page -->
    <div id="login-page" class="bg-white p-8 rounded-2xl shadow-xl max-w-md w-full m-4 transition-transform duration-500 ease-in-out transform scale-100">
        <h1 class="text-3xl font-bold text-center text-gray-800 mb-6">SL&AM Dashboard</h1>
        <form id="login-form" class="space-y-6">
            <div>
                <label for="access-group" class="block text-sm font-medium text-gray-700">User Access Group</label>
                <div class="mt-1">
                    <input id="access-group" name="access-group" type="text" required class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                </div>
            </div>
            <div id="error-message" class="text-sm text-red-500 hidden">
                Invalid access group. Please try again.
            </div>
            <div>
                <button type="submit" class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors duration-300">
                    Login
                </button>
            </div>
        </form>
    </div>

    <!-- Dashboard Page -->
    <div id="dashboard-page" class="hidden bg-white p-8 rounded-2xl shadow-xl w-full max-w-5xl transition-transform duration-500 ease-in-out transform scale-95 opacity-0">
        <div class="flex flex-col md:flex-row h-full">
            <!-- Sidebar Navigation -->
            <div class="md:w-1/4 p-4 border-b md:border-b-0 md:border-r border-gray-200">
                <h2 class="text-2xl font-semibold mb-6 text-gray-700">Navigation</h2>
                <ul id="main-nav" class="space-y-4">
                    <li><a href="#home" class="block py-2 px-4 rounded-lg bg-blue-500 text-white font-medium hover:bg-blue-600 transition-colors duration-200">Dashboard</a></li>
                    <li><a href="#asset-register" class="block py-2 px-4 rounded-lg text-gray-600 hover:bg-gray-200 transition-colors duration-200">Asset Register</a></li>
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
                                <li>Adobe CC Suite license expires on 11/15/2024</li>
                                <li>Server OS patches due on 11/20/2024</li>
                                <li>SQL Server license renewal on 12/01/2024</li>
                                <li>Annual compliance audit starts on 12/05/2024</li>
                            </ul>
                        </div>
                        <!-- Key Information Card -->
                        <div class="bg-gray-50 p-6 rounded-2xl shadow-sm border border-gray-200">
                            <h3 class="text-xl font-semibold text-gray-700 mb-4">Key Information</h3>
                            <ul class="list-disc list-inside space-y-2 text-gray-600">
                                <li>**Total Assets:** 5,432</li>
                                <li>**Active Licenses:** 1,215</li>
                                <li>**Expiring Soon:** 12</li>
                                <li>**High-Risk Items:** 3</li>
                            </ul>
                        </div>
                        <!-- Quick Actions Card -->
                        <div class="bg-gray-50 p-6 rounded-2xl shadow-sm border border-gray-200">
                            <h3 class="text-xl font-semibold text-gray-700 mb-4">Quick Actions</h3>
                            <ul class="space-y-2">
                                <li><a href="#" class="block py-2 px-4 rounded-md text-sm font-medium text-blue-600 bg-blue-100 hover:bg-blue-200 transition-colors duration-200">Add New Asset</a></li>
                                <li><a href="#" class="block py-2 px-4 rounded-md text-sm font-medium text-green-600 bg-green-100 hover:bg-green-200 transition-colors duration-200">Run Compliance Report</a></li>
                                <li><a href="#" class="block py-2 px-4 rounded-md text-sm font-medium text-red-600 bg-red-100 hover:bg-red-200 transition-colors duration-200">Review High-Risk Items</a></li>
                            </ul>
                        </div>
                    </div>
                </div>

                <!-- Placeholder Pages -->
                <div id="asset-register-page" class="placeholder-page">
                    <h2 class="text-3xl font-bold text-gray-800 mb-4">Asset Register</h2>
                    <p class="text-gray-600">This page will display a detailed list of all software and hardware assets.</p>
                </div>
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
            const loginForm = document.getElementById('login-form');
            const loginPage = document.getElementById('login-page');
            const dashboardPage = document.getElementById('dashboard-page');
            const errorMessage = document.getElementById('error-message');
            const navLinks = document.querySelectorAll('#main-nav a');
            const pages = document.querySelectorAll('.placeholder-page');

            loginForm.addEventListener('submit', (event) => {
                event.preventDefault();
                const accessGroup = document.getElementById('access-group').value;
                if (accessGroup === "SL&AM Team") {
                    loginPage.classList.remove('scale-100');
                    loginPage.classList.add('scale-95', 'opacity-0');
                    setTimeout(() => {
                        loginPage.classList.add('hidden');
                        dashboardPage.classList.remove('hidden');
                        dashboardPage.classList.remove('scale-95', 'opacity-0');
                        dashboardPage.classList.add('scale-100', 'opacity-100');
                    }, 500);
                } else {
                    errorMessage.classList.remove('hidden');
                }
            });

            navLinks.forEach(link => {
                link.addEventListener('click', (event) => {
                    event.preventDefault();
                    // Update active link styling
                    navLinks.forEach(nav => nav.classList.remove('bg-blue-500', 'text-white', 'hover:bg-blue-600'));
                    navLinks.forEach(nav => nav.classList.add('text-gray-600', 'hover:bg-gray-200'));
                    event.target.classList.add('bg-blue-500', 'text-white', 'hover:bg-blue-600');
                    event.target.classList.remove('text-gray-600', 'hover:bg-gray-200');

                    // Show the corresponding page
                    const pageId = event.target.getAttribute('href').substring(1) + "-page";
                    pages.forEach(page => page.classList.remove('active-page'));
                    document.getElementById(pageId).classList.add('active-page');
                });
            });
        });
    </script>
</body>
</html>
`

func initDB() {
	var err error
	dbPath := "./slam.db"

	// Check if the database file exists
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		log.Println("Database file not found, creating a new one...")
		dbFile, err := os.Create(dbPath)
		if err != nil {
			log.Fatalf("Error creating database file: %v\n", err)
		}
		dbFile.Close()
	}

	// Open the database connection using the "sqlite" driver
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Unable to open database: %v\n", err)
	}

	// Create the assets table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			license_key TEXT,
			purchase_date TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Error creating assets table: %v\n", err)
	}

	log.Println("Database initialized successfully.")
}

// startServer is the main logic for starting the server.
// We've moved the logic to a separate function to avoid the "main redeclared" error.
func startServer() {
	// Create a file server to serve the HTML content
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, dashboardContent)
	})

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// The main function is the entry point of the program.
// It calls the initDB function to initialize the database
// and then calls the startServer function to start the web server.
func main() {
	// Initialize the database before starting the server
	initDB()
	defer db.Close()
	
	startServer()
}

A comprehensive Software License & Asset Management (SL&AM) system in Go with the following features:
🔑 Authentication & Access Control

Role-based access with 4 user groups: Administrator, License Manager, Asset Analyst, Viewer
Session management with secure cookie handling
Default admin account: username admin, password admin123

🏗️ Database Structure (SQLite with modernc.org/sqlite)

Users, licenses, assets, events, risk register, FOI requests
Session management for security
Sample data pre-populated for demo

📊 Core Dashboard Features

Current month overview with license statistics
Event calendar showing renewals, expiries, audits
Quick action navigation to all modules

🎯 Main Functional Areas
1. Calendar View

Events, renewals, and expiry tracking for current month
Color-coded priorities (Critical, High, Medium)
Ready for full calendar integration

2. Risk Register

Risk assessment with probability/impact analysis
Color-coded risk levels
Mitigation tracking and ownership

3. FOI Request Management

Request tracking with due dates
Status management (Received, Processing, Completed)
Assignment capabilities

4. Report Execution

Placeholder structure for 6 report types
License compliance, cost analysis, asset inventory
Risk assessment and audit trail reports

5. Placeholder Modules

Settings, License Management, Asset Management
Ready for additional code modules to be added later

🚀 To Run:

Install dependencies:
BaSH:
go mod init slam-system
go get modernc.org/sqlite
go get golang.org/x/crypto/bcrypt

Run the application:
go run main.go


SL&AM Dashboard Go Application
This is a simple web server application written in Go that serves a static HTML, CSS, and JavaScript dashboard. The application is designed to simulate a Software License & Asset Management (SL&AM) dashboard, providing a front-end interface for various modules. All the front-end code is self-contained within a single Go file, making it easy to deploy and run.

Features
Login Page: A basic login interface with a hardcoded access group check.

Responsive Dashboard: A clean, responsive dashboard layout built with Tailwind CSS.

Module Navigation: Links to various modules, including:

Asset Register

Compliance Audits

License Renewals

Risk Register

Report Execution

Freedom of Information Requests (FOI)

Placeholder Pages: Each module link leads to a placeholder page, indicating future development.

Login Details
To access the dashboard, you must enter the correct user access group on the login page. The hardcoded value for this prototype is:

SL&AM Team

You will not be able to proceed to the dashboard without entering this exact text.

Prerequisites
You need to have Go installed on your system to run this application. You can download and install it from the official Go website.

How to Run
Make sure you have a file named main.go with the provided code.

Open your terminal or command prompt.

Navigate to the directory where main.go is located.

Run the application using the following command:

go run main.go


Once the server is running, open your web browser and go to http://localhost:8080 to access the dashboard.

Future Enhancements
The current modules are placeholders. Future development could include:

Implementing a persistent database for asset and license data.

Building out the functionality for each module, such as asset search and filtering, automated audit tools, and report generation.

Adding user authentication and more robust access control.
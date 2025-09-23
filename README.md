# Tibula

Tibula is a powerful and flexible application framework for building data-centric web applications. It allows you to define and evolve your application's structure (data models, fields, permissions, UI hints) as metadata stored within the database itself. This approach significantly reduces the need to write traditional backend application code, enabling you to dynamically create and manage complex applications—like CRMs, inventory managers, or project trackers—primarily through its user-friendly web interface. 

## Key Features
* **Dynamic Data Modeling:** Define, modify, and evolve your application's data models (tables and fields) directly from the web UI. Your application's structure can change in real-time without redeploying or writing traditional backend code.
* **User-Friendly Web Interface:** A clean, intuitive interface for all data interactions, including creating, searching, editing, deleting, and linking records.
* **Headless JSON API:** In addition to the web UI, every action in Tibula is accessible via a JSON API, making it a perfect backend for custom front-ends or integrations.
* **Powerful Plugin System:** Extend Tibula's core functionality by writing Go functions that hook into specific modules, allowing for custom business logic, complex validations, or unique workflows where direct Go code is required.
* **Robust Permission System:** Manage access control with a flexible system based on users and groups. Define exactly who can see and do what.
* **Advanced Data Integration (SQL-driven):** For dynamic field values and selection options, Tibula leverages embedded SQL queries, making a basic understanding of SQL beneficial for advanced configurations.
* **Data Portability:** A practical Import/Export feature allows for smooth data interchange of entire module definitions and their content in JSON format.
* **Multi-Database Support:** Run your application on either SQLite for simplicity and portability or MySQL for production scale.

## Getting Started

To get started with Tibula, simply clone the repository and build the project using the following commands:
```
git clone https://github.com/eja/tibula
cd tibula
go build
./tibula --wizard
./tibula --start
```

Tibula will be accessible at `http://localhost:35248` by default.

## Command-line Options:

Tibula provides extensive command-line options to configure various aspects of its functionality. Key options include:

- **Database Configuration**
  - Users can specify database connection details such as hostname, port, type, name, username, and password.
    ```bash
    --db-host      # Database hostname
    --db-port      # Database port
    --db-type      # Database type (sqlite/mysql)
    --db-name      # Database name or filename
    --db-user      # Database username
    --db-pass      # Database password
    ```
    ***Note:***
    By default, the database type is set to `sqlite` and the default database name is `tibula.db` in the current directory.

- **Language and Logging:**
  - Default language code and log level can be configured.
    ```bash
    --language     # Default language code
    --log-level    # Log level (1-5): Error, Warn, Info, Debug, Trace
    ```
    ***Note:***
    The default language is set to "en" for English. The log level is set to 3 (Info) by default, providing information, errors and warnings.

- **Setup and Initialization:**
  - Options for initializing the database, setting up the admin user, and defining the importing path.
    ```bash
    --setup        # Initialize the database
    --setup-user   # Admin username
    --setup-pass   # Admin password
    --setup-path   # Setup files path
    ```
    ***Note:***
      If `--setup-path` is not provided, the embedded assets will be used to import the default modules.
      The admin user is set to `admin` by default, you can customize it using `--setup-user`.

- **Web Service Configuration:**
  - Configuration options for starting the web service, specifying the host, port, path, and SSL/TLS certificates.
    ```bash
    --start            # Start the web service
    --web-host         # Web listen address
    --web-port         # Web listen port
    --web-path         # Web path
    --web-tls-private  # SSL/TLS private certificate
    --web-tls-public   # SSL/TLS public certificate
    ```
    ***Note:***
      By default, the host is set to `localhost` and the port is set to `35248`.
      If both TLS options are provided, the web service will default to `https`.
      If `--web-path` is not provided, the embedded assets will be used instead.

- **Wizard Option:**
  - The `--wizard` option guides you through a step-by-step setup configuration.
    ```bash
    --wizard           # Use the wizard for step-by-step setup configuration
    ```
  
- **General Options:**
  - The `-help` option provides a summary of available command-line options.
    ```bash
    --help             # Show this message
    ```

## JSON Configuration:

- Users can also pass command-line options using a JSON file.
  - Create a JSON file with parameters and use `--config` to specify it.
    ```bash
    --config          # Specify a JSON config file
    ```

- **Note:** Replace '-' with '_' and remove '--' for each command option when using JSON configuration.


## License

Tibula is released under the [GPL-3.0](LICENSE).

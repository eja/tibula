# Tibula

Tibula is a powerful and flexible web-based Relational Database Management System (RDBMS) that allows for dynamic table and field customization. With an intuitive web interface for data interaction and the ability to easily organize and customize your database structure.

## Key Features

* **User-friendly web interface:** Tibula provides a user-friendly web interface for easy data interaction.
* **Dynamic table and field customization:** Tibula allows for the flexible addition of tables and fields, giving you the ability to adapt and evolve your database structure as needed.
* **Hierarchical access:** Tibula allows for hierarchical access, giving you fine-grained control over who can access your data.
* **Group management:** Tibula includes group management features, making it easy to organize and manage access to your data.
* **Extensible:** Tibula efficiently manages database tables and their data, incorporating a practical Import/Export feature for smooth interchange in JSON format.
* **Support for multiple databases:** Tibula supports both SQLite3 and MySQL, giving you the flexibility to choose the database that best meets your needs.

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

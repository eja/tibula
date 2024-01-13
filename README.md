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
./tibula --setup
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

- **Language and Logging:**
  - Default language code and log level can be configured.
    ```bash
    --language     # Default language code
    --log-level    # Log level (1-5): Error, Warn, Info, Debug, Trace
    ```

- **Setup and Initialization:**
  - Options for initializing the database, setting up the admin user, and defining paths.
    ```bash
    --setup        # Initialize the database
    --setup-user   # Admin username during setup
    --setup-pass   # Admin password during setup
    --setup-path   # Setup files path
    ```

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

- **Note:** Replace '-' with '_' and remove '--' for each command when using JSON configuration.


## License

Tibula is released under the [GPL-3.0](LICENSE).

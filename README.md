# Tibula

Tibula is a powerful and flexible web-based Relational Database Management System (RDBMS) that allows for dynamic table and field customization. With an intuitive web interface for data interaction and the ability to easily organize and customize your database structure.

## Key Features

* **User-friendly web interface:** Tibula provides a user-friendly web interface for easy data interaction.
* **Dynamic table and field customization:** Tibula allows for the flexible addition of tables and fields, giving you the ability to adapt and evolve your database structure as needed.
* **Hierarchical access:** Tibula allows for hierarchical access, giving you fine-grained control over who can access your data.
* **Group management:** Tibula includes group management features, making it easy to organize and manage access to your data.
* **Extensible:** Tibula can be easily extended using simple SQL, allowing for enhanced customization and expanded functionality.
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

## License

Tibula is released under the [GPL-3.0](LICENSE).

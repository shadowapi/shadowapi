# ShadowAPI

ShadowAPI is a versatile tool designed to fetch, store, and search data from various sources, all while providing an intuitive API interface to manage these operations. It simplifies the interaction with data streams, enabling developers to build scalable and robust applications with ease.

## Running the Development Server

First, you need to install the [task](https://taskfile.dev/installation/) tool, which we use instead of Makefiles. Then, please follow these steps:

- Run the init process with `task init`. It will build Docker images for this project.
- Run `docker compose watch`. It will start all necessary containers. Please leave the terminal open; DO NOT CLOSE IT.
- In a new terminal window, run `task sync-db`. It will apply all migrations and sync the development database.
- open in the browser `http://localtest.me`, and click to [Signup](http://localtest.me/signup) link.

To start the development server, ensure you have Docker Compose installed. Then, use the following command:

Next time, you only need to run this:

```bash
docker compose watch
```

NOTE: sa-backend requires some time to up (air tool inside it installs all dependencies), check logs when it filishes works.

### Resetting the Development Environment

It is recommended to reset the development environment when significant changes have been made to the code or the Compose file. To wipe out the data and images, run the following command:

```bash
docker compose down -v --rmi all --remove-orphans
```

**Warning:** This command will remove the PostgreSQL database and all associated data.

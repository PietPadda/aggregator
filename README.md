# Gator: Your Personal RSS Feed Aggregator 🐊

Gator is a Command Line Interface (CLI) tool, built in Go, that allows you to aggregate RSS feeds from across the internet. Keep up with your favorite blogs, news sites, podcasts, and more, all from your terminal!

While the project is named "Gator," the executable installed via `go install` will be named `aggregator` based on the Go module name (`github.com/PietPadda/aggregator`).

## Features

* **Add RSS Feeds**: Easily add RSS feeds you want to follow.
* **PostgreSQL Storage**: Collected posts are stored in a PostgreSQL database.
* **Follow/Unfollow**: Manage your feed subscriptions, including following feeds added by others (on the same local setup).
* **Terminal Viewing**: View summaries of aggregated posts directly in your terminal, with links to the full content.
* **User System**: Supports multiple users on a single device (no authentication, relies on database access).
* **Continuous Aggregation**: A long-running service fetches new posts periodically.

## Prerequisites

Before you begin, ensure you have the following installed:

1.  **Go**: Version 1.23 or later. You can find installation instructions on the [official Go website](https://golang.org/doc/install).
2.  **PostgreSQL**: Version 15 or later is recommended. Installation guides are available on the [PostgreSQL website](https://www.postgresql.org/download/). You will also need the `psql` command-line utility, which is typically included with the PostgreSQL installation.
3.  **Goose**: The database migration tool. Install it via:
    ```bash
    go install [github.com/pressly/goose/v3/cmd/goose@latest](https://github.com/pressly/goose/v3/cmd/goose@latest)
    ```
    Verify by running `goose version`.

## Installation

1.  **Clone the Repository (for migrations):**
    While you can install the `aggregator` CLI directly, you'll need the repository for database migration files.
    ```bash
    git clone [https://github.com/PietPadda/aggregator.git](https://github.com/PietPadda/aggregator.git)
    cd aggregator
    ```

2.  **Install Gator (the `aggregator` CLI):**
    You can install the `aggregator` CLI to your `$GOPATH/bin` directory using its Go module path:
    ```bash
    go install [github.com/PietPadda/aggregator](https://github.com/PietPadda/aggregator)
    ```
    Ensure your `$GOPATH/bin` (or `$HOME/go/bin` by default) is in your system's `PATH` environment variable to run `aggregator` from any directory.

## Configuration

Gator uses a JSON configuration file located in your home directory: `~/.gatorconfig.json`.

1.  **Create the Configuration File:**
    Manually create the file `~/.gatorconfig.json`. You can use a text editor:
    ```bash
    nano ~/.gatorconfig.json
    ```

2.  **Add Configuration Details:**
    Paste the following structure into the file. The `db_url` is pre-filled with the details you provided.
    ```json
    {
      "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
      "current_user_name": null
    }
    ```
    * **`db_url`**: This is your PostgreSQL connection string. If your setup for username, password, host, port, or database name (`gator`) differs from the above, please adjust it accordingly. `sslmode=disable` is recommended for local development.
    * **`current_user_name`**: This will be set by the `aggregator register` or `aggregator login` commands. You can leave it as `null` or omit it initially.

## Setting Up the Database

These steps assume you have successfully installed PostgreSQL.

1.  **Access the PostgreSQL Prompt:**
    * **On Linux:** Open your terminal and connect as the `postgres` user (this typically connects you to the default `postgres` database):
        ```bash
        sudo -u postgres psql
        ```
        You should see a prompt like `postgres=#`.
    * **On macOS:** Open your terminal and connect to the default `postgres` database (you might be connecting as your own macOS user if it has superuser privileges in PostgreSQL, or as the `postgres` user depending on your setup):
        ```bash
        psql postgres
        ```
        You should see a prompt like `youruser=#` or `postgres=#`.

2.  **Create the Project Database:**
    At the `psql` prompt (`postgres=#` or similar), create the database that Gator will use. The project guide used the name `gator`:
    ```sql
    CREATE DATABASE gator;
    ```
    *(If you choose a different database name, update your `db_url` in `~/.gatorconfig.json` accordingly.)*

3.  **Connect to the New Database:**
    Still within `psql`, switch your connection to the newly created `gator` database:
    ```sql
    \c gator
    ```
    Your prompt should change to `gator=#` (or `gator=>`).

4.  **Set Database User Password (if applicable):**
    * **For the Linux setup described in your project notes (using `postgres` user):** You set a password for the `postgres` database user *after* connecting to the `gator` database. If you followed this, the command was:
        ```sql
        ALTER USER postgres PASSWORD 'postgres';
        ```
        *(This is distinct from the OS-level password for the `postgres` system user set earlier with `sudo passwd postgres`.)*
    * **For macOS or other setups/users:** Ensure the user specified in your `db_url` in `~/.gatorconfig.json` can connect to the `gator` database with the specified credentials.

5.  **Exit `psql`:**
    You can now exit the PostgreSQL prompt:
    ```sql
    \q
    ```

6.  **Run Database Migrations:**
    Navigate to the `sql/schema` directory within your cloned `aggregator` repository (e.g., `cd ~/workspace/github.com/PietPadda/aggregator/sql/schema`).
    ```bash
    cd /your/local/path/to/PietPadda/aggregator/sql/schema 
    ```
    **(Replace `/your/local/path/to/PietPadda/aggregator/` with the actual local path where you cloned the `PietPadda/aggregator` repository.)**
    Run the "up" migrations using Goose with the connection string you provided:
    ```bash
    goose postgres "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable" up
    ```
    If your setup differs, adjust the connection string accordingly. This command creates the required tables (`users`, `feeds`, `feed_follows`, `posts`).

## Usage

Once installed and configured, you can use Gator via the `aggregator` command. For example:

    aggregator <command> [arguments...]

### Available Commands

Here's a list of available commands:

* **`register <username>`**
    * Registers a new user with the given `<username>`, adds them to the database, and logs them in.
    * Example: `aggregator register PietPadda`

* **`login <username>`**
    * Logs in an existing user. Sets the `current_user_name` in `~/.gatorconfig.json`.
    * Exits with an error if the user doesn't exist in the database.
    * Example: `aggregator login PietPadda`

* **`users`**
    * Lists all registered users in the database, indicating the currently logged-in user.
    * Example: `aggregator users`

* **`addfeed <feed_name> "<feed_url>"`**
    * Adds a new RSS feed with the given `<feed_name>` and `<feed_url>` for the currently logged-in user. The user automatically follows this new feed.
    * Enclose `<feed_url>` in quotes if it contains special characters.
    * Example: `aggregator addfeed "Go Blog" "https://go.dev/blog/feed.atom"`

* **`feeds`**
    * Lists all feeds currently stored in the database, showing the feed's name, URL, and the username of the user who originally added it.
    * Example: `aggregator feeds`

* **`follow "<feed_url>"`**
    * Allows the currently logged-in user to follow an existing feed specified by its `<feed_url>`.
    * Example: `aggregator follow "https://go.dev/blog/feed.atom"`

* **`unfollow "<feed_url>"`**
    * Allows the currently logged-in user to unfollow a feed specified by its `<feed_url>`.
    * Example: `aggregator unfollow "https://go.dev/blog/feed.atom"`

* **`following`**
    * Prints the names of all RSS feeds that the currently logged-in user is following.
    * Example: `aggregator following`

* **`agg <duration>`**
    * Starts the long-running aggregator service. This service will periodically fetch all registered feeds, parse new posts, and store them in the database.
    * `<duration>` specifies the interval between fetch cycles (e.g., `30s` for 30 seconds, `5m` for 5 minutes, `1h` for 1 hour).
    * The command will print "Collecting feeds every Xs" and then log its activity.
    * This command runs indefinitely. To stop it, press `Ctrl+C`. It's intended to be run in a separate terminal window or in the background.
    * Example: `aggregator agg 10m` (fetches every 10 minutes)

* **`browse [limit]`**
    * Displays posts from the feeds that the currently logged-in user is following.
    * `[limit]` is an optional integer specifying the maximum number of posts to display. If not provided, it defaults to 2 posts.
    * Posts are shown with their title, URL, publication date, and content.
    * Example (default limit): `aggregator browse`
    * Example (custom limit): `aggregator browse 10`

* **`reset`**
    * Resets the Gator database by deleting all users, feeds, feed follows, and posts.
    * **Caution**: This is a destructive operation and primarily intended for development or testing purposes.
    * Example: `aggregator reset`

## Development

If you want to contribute or build from source:

1.  **Clone the Repository:**
    ```bash
    git clone [https://github.com/PietPadda/aggregator.git](https://github.com/PietPadda/aggregator.git)
    cd aggregator
    ```
2.  **Initialize Go Module (if not already done for your local clone):**
    Your `go.mod` file should already be initialized to `github.com/PietPadda/aggregator`. To ensure dependencies are up to date:
    ```bash
    go mod tidy
    ```
3.  **Build the Executable:**
    From the root of the project directory (`aggregator`):
    ```bash
    go build
    ```
    This creates an executable named `aggregator` in the project root. You can run it as `./aggregator <command>`.

4.  **Run for Development:**
    To run the application directly without building an executable each time (assuming your `main.go` is in the root of the `aggregator` directory):
    ```bash
    go run . <command> [args...]
    ```

## Contributing

We welcome contributions to Gator! If you have suggestions, bug reports, or want to contribute code, please feel free to:

1.  Open an issue on the GitHub repository to discuss the change or bug.
2.  Fork the repository, create a new branch for your feature or fix (`git checkout -b feature/your-feature-name`), and make your changes.
3.  Submit a pull request with a clear description of your work.

Please ensure any new code follows the existing style and all tests pass (if tests are implemented).

## License

This project is licensed under the MIT License.

You should create a `LICENSE.md` file in the root of your repository and add the following text to it:

```text
MIT License

Copyright (c) 2025 PietPadda


Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
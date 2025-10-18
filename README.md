# Go Yuga Dist

This project is a geo-distributed application using Golang, YugabyteDB, and Kubernetes.

## Architecture

The application consists of the following components:

- **Go Backend:** A stateless Go application that provides a RESTful API for user registration, login, and posts.
- **YugabyteDB:** A distributed SQL database used as the persistent storage layer. The data is geo-partitioned based on the user's region.
- **Nginx Ingress:** An Nginx ingress controller that acts as a load balancer and routes traffic to the Go backend pods.
- **Kubernetes:** The container orchestration platform used to deploy and manage all the components.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Helm](https://helm.sh/docs/intro/install/)
- A running Kubernetes cluster (e.g., Minikube, Kind, or a cloud provider's managed Kubernetes service)

## Setup (Automated with Tilt)

This project uses [Tilt](https://tilt.dev/) to provide a fully automated local development environment. A single command will bring up the entire stack, including the simulated geo-distributed YugabyteDB cluster and the Go application with live-reloading.

### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Helm](https://helm.sh/docs/intro/install/)
- [Tilt](https://docs.tilt.dev/install.html)
- [Docker](https://docs.docker.com/get-docker/)
- A running local Kubernetes cluster (e.g., Minikube, Docker Desktop)

### Running the Environment

1.  **Start Tilt:**
    Run the following command from the root of the project:
    ```bash
    tilt up
    ```
2.  **Access the Application:**
    Tilt will stream all the logs and provide a web UI (usually at `http://localhost:10350/`) to monitor the status of all services. Once the `go-yuga-dist` service is green, the API will be available through the Nginx Ingress.

Tilt handles the following steps for you:

- Deploys the YugabyteDB cluster using Helm.
- Creates the database and user.
- Creates the Kubernetes secret for database credentials.
- Runs the database schema migrations.
- Builds the Go application's Docker image.
- Deploys the Go application and Nginx ingress.
- Watches your Go files for changes, automatically rebuilds, and updates the running application.

### Running the Integration Tests

The project includes an integration test suite that verifies the entire API and data partitioning logic.

1.  **Start the Environment:**
    Make sure the full application stack is running via Tilt:
    ```bash
    tilt up
    ```
2.  **Run the Tests:**
    Once all services in Tilt are green, open a new terminal window, navigate to the `app` directory, and run the tests:

    ```bash
    cd app
    go test -v
    ```

    The tests will execute against the live API port-forwarded by Tilt and confirm that the user creation, authentication, and post creation flows are working correctly, including the data sovereignty check.

3.  **Inspect the Data:**

    1. Start Tilt:

    ```bash
    tilt up
    ```

    2. Open the YugabyteDB Master UI in your browser:

    ```text
    http://localhost:7000
    ```

    3. Connect to the database shell from within the tserver pod:

    ```bash
    kubectl exec -it -n go-yuga-dist yb-tserver-0 -- ysqlsh -h yb-tservers.go-yuga-dist -U go_yuga_dist_user -d go_yuga_dist
    ```

    4. Run these SQL queries to inspect test data:

    ```sql
    SELECT * FROM users;
    SELECT * FROM global_users;
    SELECT * FROM posts;
    SELECT * FROM users_usa;
    SELECT * FROM posts_usa;
    ```

    The integration test creates a user in the USA region and a test post; verify the geo-partitioning by checking the regional partition tables (`users_usa`, `posts_usa`).

## API Endpoints

- `POST /signup`: Register a new user.
- `POST /signin`: Login as a user.
- `POST /user-posts`: Create a new post.
- `GET /user-posts`: List all posts of the logged in user.
- `GET /user-post/:post-id`: Fetch a particular post.

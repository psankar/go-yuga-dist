# Tiltfile for go-yuga-dist
# See https://docs.tilt.dev/

# --- Configuration ---

# Define the name of our Go application's image.
# Tilt will build this image and inject the tag into our k8s manifests.
APP_IMAGE = 'go-yuga-dist'

# --- Docker Image Build ---

# Define how to build our Go application.
# On file changes, Tilt will sync the files and re-compile the binary inside the container.
docker_build(APP_IMAGE, 'app',
    live_update=[
        sync('./app', '/app'),
        # The WORKDIR in the Dockerfile is /app, so this command runs in the right place.
        run('go build -o /go-yuga-dist .', trigger='./app/**/*.go'),
    ])

# Build the migration image which packages the SQL migration inside an image.
MIGRATION_IMAGE = 'go-yuga-dist-migration'
# Build migration image using repo root as context so the Dockerfile can COPY the
# top-level migrations/001_initial_schema.up.sql file. The Dockerfile itself
# lives in migration/Dockerfile.
docker_build(MIGRATION_IMAGE, 'migrations', dockerfile='migrations/Dockerfile')

# --- Kubernetes Resources ---

# Deploy YugabyteDB using helm_remote
load('ext://helm_remote', 'helm_remote')

helm_remote('yugabyte',
    repo_name='yugabytedb',
    repo_url='https://charts.yugabyte.com',
    namespace='go-yuga-dist',
    create_namespace=True,
    set=[
        'replicas.master=1',
        'replicas.tserver=5',
        'tserver.extra_flags={--placement_cloud=local,--placement_region=ind,--placement_zone=a,--placement_cloud=local,--placement_region=usa,--placement_zone=a,--placement_cloud=local,--placement_region=chn,--placement_zone=a,--placement_cloud=local,--placement_region=deu,--placement_zone=a,--placement_cloud=local,--placement_region=sgp,--placement_zone=a}',
        'resource.master.requests.memory=512Mi',
        'resource.master.limits.memory=512Mi',
        'resource.master.requests.cpu=100m',
        'resource.master.limits.cpu=100m',
        'resource.tserver.requests.memory=512Mi',
        'resource.tserver.limits.memory=512Mi',
        'resource.tserver.requests.cpu=100m',
        'resource.tserver.limits.cpu=100m',
        'enable_geo_partitioning=true',
    ]
)

# 2. Load all Kubernetes manifests
# Tilt will manage all these resources.
k8s_yaml([
    'k8s/secret.yaml',
    'k8s/rbac.yaml',
    'k8s/db-setup-job.yaml',
    'k8s/migration-job.yaml',
    'k8s/deployment.yaml',
    'k8s/ingress.yaml',
])

# --- Resource Dependencies ---
# Define the startup order to ensure resources are created correctly.

# The db-setup job depends on YugabyteDB being ready
k8s_resource('db-setup', resource_deps=['yb-demo-yugabyte'])

# The db-migration job depends on the db-setup job finishing.
k8s_resource('db-migration', resource_deps=['db-setup'])

# Configure port forwards for YugabyteDB services
k8s_resource(
    'yb-master',  # The StatefulSet name
    port_forwards=[
        port_forward(7000, 7000, name='YugabyteDB Admin UI'),
    ],
    resource_deps=['yb-demo-yugabyte'],
)

k8s_resource(
    'yb-tserver',  # The StatefulSet name
    port_forwards=[
        port_forward(5433, 5433, name='YugabyteDB YSQL'),
    ],
    resource_deps=['yb-demo-yugabyte'],
)

# The application 'go-yuga-dist' deployment depends on the database migration finishing.
k8s_resource('go-yuga-dist', resource_deps=['db-migration'], 
    port_forwards=[
        # API server
        port_forward(8080, 8080, name='API Server'),
    ])

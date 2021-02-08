module flow.stacc.dev/database-provisioning-poc

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang-migrate/migrate v3.5.4+incompatible
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.10.1
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/prometheus/common v0.4.1
	github.com/sethvargo/go-password v0.2.0
	go.mongodb.org/mongo-driver v1.1.2
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)

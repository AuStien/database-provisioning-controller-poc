# Database Provisioning Controller PoC
A controller built on [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) automatically creating a user and database on a database server.

# About
Currently supports Postgresql, Mysql and Mongodb.
## Custom Resources
### DatabaseServer
DatabaseServer contains info about where the database server is located and how to connect to it.

A yaml example of how to define a DatabaseServer:
```YAML
apiVersion: database.stacc.com/v1alpha1
kind: DatabaseServer
metadata:
  name: postgres-server
spec:
  type: postgresql
  secretName: postgres-server-secret
  postgres:
    host: localhost
    username: postgres
    port: 5432
    useSsl: false

```
- type: The type of database server being used. [postgres, mysql or mongo]
- secretName: Name of the secret where the password used to login is stored. [Must contain a field called "password"]
- postgres: Object with the name of type of database. Must be either postgres, mysql, or mongo.
  - host: Hostname of server
  - username: Username used to login
  - port: Port of the server
  - useSsl: If SSL shall be used. (Currently not implemented)

### Database
Provides the name of the database and user to be created.

Example:
```YAML
apiVersion: database.stacc.com/v1alpha1
kind: Database
metadata:
  name: postgres-db
spec:
  name: postgres-db
  username: db-user
  deletable: true
  server:
    name: postgres-server
    namespace: default
  secret:
    name: postgres-db-secret
    namespace: default

```
- name: The name of the database
- username(Optional): The username to be associated with the database. If omitted will default to the name of the database.
- deletable: If the user and database shall be deleted when this resource is.
- server: Points to the DatabaseServer resource this database shall be hosted on.
  - name: Name of the resource.
  - namespace: The namespace the resource currently is.
- secret: A secret will be created with fields "username" and "password", used to login to the new database.
  - name: The name of the secret.
  - namespace: In which namespace the secret will be stored.
  
### More examples
Examples for both resources made for all types of databases can be found [here](https://github.com/AuStien/database-provisioning-controller-poc/tree/main/config/samples).
- DatabaseServer
  - [Postgres](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_databaseserver_postgres.yaml)
  - [Mysql](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_databaseserver_mysql.yaml)
  - [Mongo](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_databaseserver_mongo.yaml)
- Database
  - [Postgres](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_database_postgres.yaml)
  - [Mysql](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_database_mysql.yaml)
  - [Mongo](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_database_mongo.yaml)
  
  # Getting started
  (Coming soon)

/my-vkr-project
├── auth-service/               # Go + Echo
├── order-service/              # Go + Echo + Kafka producer
├── inventory-service/          # Go + Echo + Kafka consumer
├── charts/
│   ├── auth-service/
│   ├── order-service/
│   ├── inventory-service/
│   ├── postgres-auth/
│   ├── postgres-orders/
│   ├── postgres-inventory/
│   ├── redis/
│   ├── kafka/
│   ├── krakend/
│   └── signoz/                 # или используем официальный чарт
├── krakend/
│   └── krakend.json
├── k8s/
│   └── minikube-config.yaml
└── README.md


Теперь нужно создать Helm-чарты для:

PostgreSQL (x3) — по одной БД на сервис
Redis
Kafka + ZooKeeper (или KRaft)
KrakenD
SigNoz (можно использовать официальный чарт )

1. kubectl proxy --address=0.0.0.0 --port=8001 --accept-hosts='^.*$'

2. Redis
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
touch redis-values.yaml
helm install my-redis bitnami/redis -f charts/redis/redis-values.yaml --namespace default
kubectl get pods
kubectl top pod
kubectl logs <redis-pod-name>
kubectl exec -it <redis-pod-name> -- redis-cli

3. SigNoz
helm repo add signoz https://charts.signoz.io
helm repo update
signoz-values.yaml
helm install my-signoz signoz/signoz -f charts/signoz/signoz-values.yaml --namespace observability --create-namespace
kubectl get pods -n observability
kubectl top pod -n observability
kubectl logs -n observability <pod-name>
kubectl port-forward --address 0.0.0.0 -n observability svc/my-signoz 8080:8080

3. PostgreSQL
helm install postgres1 bitnami/postgresql -f values/auth-db-values.yaml --namespace db1 --create-namespace
helm install postgres2 bitnami/postgresql -f values/order-db-values.yaml --namespace db2 --create-namespace
helm install postgres3 bitnami/postgresql -f values/inventory-db-values.yaml --namespace db3 --create-namespace

kubectl exec -it -n db1 postgres1-postgresql-0 -- psql -U user1 -d db1 -c "CREATE USER monitoring_user1 WITH PASSWORD 'monitoring_password1'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO monitoring_user1;"
kubectl exec -it -n db2 postgres2-postgresql-0 -- psql -U user2 -d db2 -c "CREATE USER monitoring_user2 WITH PASSWORD 'monitoring_password2'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO monitoring_user2;"
kubectl exec -it -n db3 postgres3-postgresql-0 -- psql -U user3 -d db3 -c "CREATE USER monitoring_user3 WITH PASSWORD 'monitoring_password3'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO monitoring_user3;"


minikube image pull docker.io/bitnami/postgresql:17.6.0-debian-12-r4

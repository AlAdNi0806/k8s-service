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

--------------------------------------------------------------------------------
helm list --all-namespaces
kubectl get svc -n default
helm uninstall auth-db --namespace db-auth
kubectl get pods -w --namespace db-auth -l app.kubernetes.io/instance=auth-db

0.
kubectl port-forward --address 0.0.0.0 -n observability svc/my-signoz 8080:8080
kubectl proxy --address=0.0.0.0 --port=8001 --accept-hosts='^.*$'


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

kubectl port-forward --address 0.0.0.0 -n default svc/my-redis 6379:6379

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
kubectl create secret generic mariadb-root-password --from-literal=mariadb-root-password='your_root_password_here' --namespace default

helm upgrade --install auth-db bitnami/mariadb -f values/maria-db-values.yaml

kubectl exec -it auth-db-mariadb-0 -- mysql -u root -p
CREATE DATABASE IF NOT EXISTS auth_db;
CREATE USER IF NOT EXISTS 'auth_user'@'%' IDENTIFIED BY 'auth_pass';
GRANT ALL PRIVILEGES ON auth_db.* TO 'auth_user'@'%';

CREATE DATABASE IF NOT EXISTS order_db;
CREATE USER IF NOT EXISTS 'order_user'@'%' IDENTIFIED BY 'order_pass';
GRANT ALL PRIVILEGES ON order_db.* TO 'order_user'@'%';

CREATE DATABASE IF NOT EXISTS inventory_db;
CREATE USER IF NOT EXISTS 'inventory_user'@'%' IDENTIFIED BY 'inventory_pass';
GRANT ALL PRIVILEGES ON inventory_db.* TO 'inventory_user'@'%';

FLUSH PRIVILEGES;
exit;

kubectl port-forward --address 0.0.0.0 -n default svc/auth-db-mariadb 3306:3306

mysql -h 127.0.0.1 -P 3306 -u auth_user -p auth_db

4. Open tellemtry
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update

helm install my-otel-collector open-telemetry/opentelemetry-collector -f values/otel-collector-config.yaml -n default --create-namespace

kubectl logs -n default -l app.kubernetes.io/name=opentelemetry-collector --follow

kubectl port-forward --address 0.0.0.0 -n default svc/my-otel-collector-opentelemetry-collector 4318:4318

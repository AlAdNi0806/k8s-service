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
helm uninstall auth-db --namespace default
kubectl get pods -w --namespace db-auth -l app.kubernetes.io/instance=auth-db
kubectl get ns // namespaces

kubectl port-forward --address 0.0.0.0 -n default svc/my-signoz-otel-collector 4317:4317
kubectl port-forward --address 0.0.0.0 -n default svc/my-signoz 8080:8080
kubectl port-forward --address 0.0.0.0 -n default svc/my-redis-master 6379:6379
kubectl port-forward --address 0.0.0.0 -n default svc/mariadb 3306:3306
kubectl proxy --address=0.0.0.0 --port=41391 --accept-hosts='^.*$'

0.
kubectl port-forward --address 0.0.0.0 -n observability svc/my-signoz 8080:8080
kubectl proxy --address=0.0.0.0 --port=8001 --accept-hosts='^.*$'


1. kubectl proxy --address=0.0.0.0 --port=8001 --accept-hosts='^.*$'

2. Redis
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
touch redis-values.yaml
helm install my-redis bitnami/redis -f values/redis-values.yaml --namespace default
kubectl get pods
kubectl top pod
kubectl logs <redis-pod-name>
kubectl exec -it <redis-pod-name> -- redis-cli
kubectl port-forward --address 0.0.0.0 -n default svc/my-redis 6379:6379
kubectl get secret --namespace default my-redis -o jsonpath="{.data.redis-password}" | base64 --decode

3. SigNoz
helm repo add signoz https://charts.signoz.io
helm repo update
signoz-values.yaml
helm uninstall my-signoz
helm install my-signoz signoz/signoz -f values/signoz-values.yaml --namespace default
kubectl get pods -n observability
kubectl top pod -n observability
kubectl logs -n observability <pod-name>
kubectl port-forward --address 0.0.0.0 -n observability svc/my-signoz 8080:8080

3. PostgreSQL
kubectl create secret generic mariadb-root-password --from-literal=mariadb-root-password='mariadb-root-pass' --namespace default

helm upgrade --install mariadb bitnami/mariadb -f values/maria-db-values.yaml

kubectl exec -it mariadb-0 -- mysql -u root -p
CREATE DATABASE IF NOT EXISTS auth_db;
CREATE USER IF NOT EXISTS 'auth_user'@'%' IDENTIFIED BY 'auth_pass';
GRANT ALL PRIVILEGES ON auth_db.* TO 'auth_user'@'%';

CREATE DATABASE IF NOT EXISTS order_db;
CREATE USER IF NOT EXISTS 'order_user'@'%' IDENTIFIED BY 'order_pass';
GRANT ALL PRIVILEGES ON order_db.* TO 'order_user'@'%';

CREATE DATABASE IF NOT EXISTS inventory_db;
CREATE USER IF NOT EXISTS 'inventory_user'@'%' IDENTIFIED BY 'inventory_pass';
GRANT ALL PRIVILEGES ON inventory_db.* TO 'inventory_user'@'%';

CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL
);

CREATE TABLE orders (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  product_id BIGINT NOT NULL,
  quantity INT NOT NULL,
  status VARCHAR(50) DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock (
  product_id BIGINT PRIMARY KEY,
  quantity INT NOT NULL CHECK (quantity >= 0)
);

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

5. Kafka
kubectl apply -f https://github.com/strimzi/strimzi-kafka-operator/releases/download/0.40.0/strimzi-cluster-operator-0.40.0.yaml
kubectl delete kafka my-cluster -n kafka
kubectl apply -f values/kafka-values.yaml

#!/bin/bash

set -e  # Остановить при любой ошибке

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

echo "🚀 Starting VKR deployment to Minikube..."

# 1. Запуск Minikube (если не запущен)
if ! minikube status &>/dev/null; then
    echo "▶️ Starting Minikube..."
    minikube start --memory=8192 --cpus=4 --driver=docker
else
    echo "✅ Minikube already running"
fi

# 2. Настройка Docker для Minikube
echo "🐳 Configuring Docker to use Minikube's daemon..."
eval $(minikube docker-env)

# 4. Добавление Helm-репозиториев
echo "📡 Adding Helm repositories..."
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add signoz https://charts.signoz.io
helm repo update

# 5. Установка PostgreSQL (3 экземпляра) с инициализацией таблиц
echo "🗄️ Installing PostgreSQL instances with tables..."
helm upgrade --install postgres-auth bitnami/postgresql \
  --version 15.x.x \
  --set auth.username=auth_user \
  --set auth.password=auth_pass \
  --set auth.database=auth_db \
  --set primary.persistence.enabled=true \
  --set primary.persistence.size=1Gi \
  --set primary.resources.limits.cpu=500m \
  --set primary.resources.limits.memory=512Mi \
  --set primary.extendedConfiguration="max_connections = 100" \
  --set primary.initdbScripts.create_monitor_user.sql="CREATE USER monitor_user WITH PASSWORD 'monitor_pass'; GRANT pg_monitor TO monitor_user;" \
  --set-file primary.initdbScripts.auth-init.sql="$PROJECT_ROOT/init-sql/auth-init.sql"

helm upgrade --install postgres-orders bitnami/postgresql \
  --version 15.x.x \
  --set auth.username=order_user \
  --set auth.password=order_pass \
  --set auth.database=order_db \
  --set primary.persistence.enabled=true \
  --set primary.persistence.size=1Gi \
  --set primary.resources.limits.cpu=500m \
  --set primary.resources.limits.memory=512Mi \
  --set primary.extendedConfiguration="max_connections = 100" \
  --set primary.initdbScripts.create_monitor_user.sql="CREATE USER monitor_user WITH PASSWORD 'monitor_pass'; GRANT pg_monitor TO monitor_user;" \
  --set-file primary.initdbScripts.orders-init.sql="$PROJECT_ROOT/init-sql/orders-init.sql"

helm upgrade --install postgres-inventory bitnami/postgresql \
  --version 15.x.x \
  --set auth.username=inventory_user \
  --set auth.password=inventory_pass \
  --set auth.database=inventory_db \
  --set primary.persistence.enabled=true \
  --set primary.persistence.size=1Gi \
  --set primary.resources.limits.cpu=500m \
  --set primary.resources.limits.memory=512Mi \
  --set primary.extendedConfiguration="max_connections = 100" \
  --set primary.initdbScripts.create_monitor_user.sql="CREATE USER monitor_user WITH PASSWORD 'monitor_pass'; GRANT pg_monitor TO monitor_user;" \
  --set-file primary.initdbScripts.inventory-init.sql="$PROJECT_ROOT/init-sql/inventory-init.sql"

# 6. Установка Redis
echo "💾 Installing Redis..."
helm upgrade --install redis bitnami/redis \
  --set auth.enabled=false \
  --set master.persistence.enabled=false

# 7. Установка Kafka (в режиме KRaft)
echo "📬 Installing Kafka..."
helm upgrade --install kafka bitnami/kafka \
  --set controller.replicaCount=1 \
  --set listeners.client.plaintext.port=9092 \
  --set zookeeper.enabled=false \
  --set kraft.enabled=true \
  --set kraft.controller.quorum="1@kafka-controller-0.kafka-controller-headless.default.svc.cluster.local:9093" \
  --set persistence.enabled=false

# 8. Установка SigNoz
echo "🔭 Installing SigNoz..."
helm upgrade --install signoz signoz/signoz \
  --namespace signoz --create-namespace \
  -f "$PROJECT_ROOT/signoz-values-minikube.yaml"

# 9. Ожидание готовности зависимостей
echo "⏳ Waiting for databases and Kafka to be ready..."
kubectl wait --for=condition=ready pod/postgres-auth-postgresql-0 --timeout=600s
kubectl wait --for=condition=ready pod/postgres-orders-postgresql-0 --timeout=600s
kubectl wait --for=condition=ready pod/postgres-inventory-postgresql-0 --timeout=600s

# kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=redis --timeout=120s
# kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=kafka --timeout=180s

# 10. Создание Kafka-топика
echo "🎫 Creating Kafka topic 'order.created'..."
kubectl run kafka-client --restart='Never' --image docker.io/bitnami/kafka:latest --command -- sleep infinity
sleep 10
kubectl exec kafka-client -- kafka-topics.sh --create \
  --topic order.created \
  --bootstrap-server kafka:9092 \
  --partitions 1 \
  --replication-factor 1
kubectl delete pod kafka-client

# 11. Установка микросервисов
echo "🧩 Deploying microservices..."
helm upgrade --install auth-service charts/auth-service
helm upgrade --install order-service charts/order-service
helm upgrade --install inventory-service charts/inventory-service

# 12. Установка KrakenD
echo "🌉 Deploying KrakenD API Gateway..."
helm upgrade --install krakend charts/krakend

# 13. Ожидание готовности сервисов
echo "⏳ Waiting for services to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=auth-service --timeout=120s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=order-service --timeout=120s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=inventory-service --timeout=120s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=krakend --timeout=120s

# 14. Проброс портов (в фоне)
echo "🔌 Starting port forwards in background..."
kubectl port-forward svc/krakend 8080:8080 >/dev/null 2>&1 &
kubectl port-forward svc/signoz-frontend -n signoz 3301:3301 >/dev/null 2>&1 &
KRAKEND_PID=$!
SIGNOZ_PID=$!

# 15. Финальное сообщение
echo ""
echo "🎉 Deployment completed successfully!"
echo ""
echo "🔗 API Gateway (KrakenD): http://localhost:8080"
echo "   - Register: POST /auth/register"
echo "   - Login:    POST /auth/login"
echo "   - Orders:   POST /orders (with Bearer token)"
echo ""
echo "📊 SigNoz UI: http://localhost:3301"
echo ""
echo "📌 To stop port forwards, run:"
echo "   kill $KRAKEND_PID $SIGNOZ_PID"
echo ""
echo "💡 Test command:"
echo "curl -X POST http://localhost:8080/auth/register -H 'Content-Type: application/json' -d '{\"email\":\"test@example.com\",\"password\":\"123456\"}'"

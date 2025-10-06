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

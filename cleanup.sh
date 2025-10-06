#!/bin/bash
helm uninstall postgres-auth postgres-orders postgres-inventory redis kafka signoz auth-service order-service inventory-service krakend -n default
helm uninstall signoz -n signoz
minikube delete

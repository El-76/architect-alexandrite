# Jaeger в Minikube с сервисами

## Описание
Развертывание Jaeger в Minikube с двумя сервисами, которые:
1. Взаимодействуют между собой
2. Отправляют трейсы в Jaeger

Вместо изначальных service-a и service-b сделаны orders-service (сервис заказов) и models (сервис 3D-моделей), в каждом по одной GET-ручке, которая получает соответствующую сущность по ID, orders-service идёт за данными модели в models-service в коде ручки. Ручка в models-service с вероятностью 0,5 может "тормозить" - выполняться 1 секунду.

## Требования
- Minikube
- kubectl
- Docker

## Установка

### 1. Запуск Minikube 
```bash
minikube start --addons=ingress 
```
Ingress нужен для вызовов

### 2. Установка cert-manager
```bash
minikube kubectl -- apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml
```

### 3. Развертывание Jaeger
```bash
minikube kubectl -- create namespace observability

curl -L https://github.com/jaegertracing/jaeger-operator/releases/download/v1.51.0/jaeger-operator.yaml -o k8s/jaeger-operator.yaml

sed -i 's/gcr\.io\/kubebuilder\/kube-rbac-proxy/registry.k8s.io\/kubebuilder\/kube-rbac-proxy/g' k8s/jaeger-operator.yaml

minikube kubectl -- create -f k8s/jaeger-operator.yaml -n observability

# Надо дождаться, когда поды запустятся

minikube kubectl -- apply -f k8s/jaeger-instance.yaml
```

### 4. Сборка и деплой сервисов
```bash
# Сборка образов
minikube image build -t orders-service:latest services/orders-service/
minikube image build -t models-service:latest services/models-service/

# Развертывание
minikube kubectl -- apply -f k8s/services.yaml
```

## Проверка работы

### Доступ к Jaeger UI
```bash
minikube kubectl -- port-forward --address 0.0.0.0 svc/simplest-query 16686:16686
```
Откройте в браузере: http://localhost:16686

### Тестирование сервисов
```bash
# Вызов orders-service, который вызывает models-service
minikube kubectl -- exec -it $( minikube kubectl -- get pods -l app=orders-service -o jsonpath='{.items[0].metadata.name}' ) -- wget http://orders-service:8080/orders/123456 -O -
```

### Трейсы

#### Список

![Список](./screenshots/Traces%20list.png)

#### Models отвечает медленно

![Slow Models](./screenshots/Slow%20models%20service%20trace.png)

#### Models отвечает быстро

![Fast Models](./screenshots/Fast%20models%20service%20trace.png)

## Структура проекта
- `services/orders=service/` - Исходный код orders-service
- `services/orders-service/` - Исходный код orders-service  
- `k8s/services.yaml` - Конфигурация Kubernetes для сервисов
- `k8s/jaeger-instance.yaml` - Конфигурация Jaeger

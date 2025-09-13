import json
import time
import uuid
import requests
from kafka import KafkaProducer
from kafka.errors import KafkaError

# -------------------------------
# Генерация UUID для уникальных данных
# -------------------------------
def generate_uuid():
    return str(uuid.uuid4())

# -------------------------------
# Тестовые данные для заказа
# -------------------------------
def create_test_order():
    order_uid = generate_uuid()
    transaction_uid = generate_uuid()
    item_rid = generate_uuid()

    return {
        "order_uid": order_uid,
        "track_number": "WBILMTESTTRACK",
        "entry": "WBIL",
        "delivery": {
            "name": "Test Testov",
            "phone": "+9720000000",
            "zip": "2639809",
            "city": "Kiryat Mozkin",
            "address": "Ploshad Mira 15",
            "region": "Kraiot",
            "email": "test@gmail.com"
        },
        "payment": {
            "transaction": transaction_uid,
            "request_id": "",
            "currency": "USD",
            "provider": "wbpay",
            "amount": 1817,
            "payment_dt": 1637907727,
            "bank": "alpha",
            "delivery_cost": 1500,
            "goods_total": 317,
            "custom_fee": 0
        },
        "items": [
            {
                "chrt_id": 9934930,
                "track_number": "WBILMTESTTRACK",
                "price": 453,
                "rid": item_rid,
                "name": "Mascaras",
                "sale": 30,
                "size": "0",
                "total_price": 317,
                "nm_id": 2389212,
                "brand": "Vivienne Sabo",
                "status": 202
            }
        ],
        "locale": "en",
        "internal_signature": "",
        "customer_id": "test",
        "delivery_service": "meest",
        "shardkey": "9",
        "sm_id": 99,
        "date_created": "2021-11-26T06:22:19Z",
        "oof_shard": "1"
    }

# -------------------------------
# Kafka Producer
# -------------------------------
def create_producer():
    return KafkaProducer(
        bootstrap_servers=['localhost:9092'],
        value_serializer=lambda v: json.dumps(v).encode('utf-8'),
        acks='all',
        retries=3
    )

def send_message(producer, topic, message):
    try:
        future = producer.send(topic, message)
        future.get(timeout=10)
        print(f"✅ Сообщение отправлено: {message['order_uid']}")
        return True
    except KafkaError as e:
        print(f"❌ Ошибка отправки: {e}")
        return False

# -------------------------------
# HTTP GET запрос к API
# -------------------------------
def get_order(order_uid):
    url = f"http://localhost:8081/orders/{order_uid}"
    try:
        response = requests.get(url)
        if response.status_code == 200:
            return response.json()
        else:
            print(f"⚠️ Заказ {order_uid} не найден (код {response.status_code})")
            return None
    except Exception as e:
        print(f"❌ Ошибка запроса к серверу: {e}")
        return None

# -------------------------------
# Тест кеша
# -------------------------------
def test_cache_effect(order_uid):
    print("\n--- Проверка кеша ---")

    # Первый запрос — из БД
    start_db = time.time()
    order1 = get_order(order_uid)
    time_db = time.time() - start_db
    assert order1 is not None, "❌ Первый запрос не вернул данные"
    print(f"Первый запрос (из БД) занял: {time_db:.4f} сек")

    # Второй запрос — из кеша
    start_cache = time.time()
    order2 = get_order(order_uid)
    time_cache = time.time() - start_cache
    assert order2 is not None, "❌ Второй запрос не вернул данные"
    print(f"Второй запрос (из кеша) занял: {time_cache:.4f} сек")

    assert time_cache < time_db, "❌ Кеш не ускорил запрос!"
    print("✅ Кеш работает корректно")

# -------------------------------
# Главная функция тестов
# -------------------------------
def main():
    producer = create_producer()
    topic = "orders"

    # Создаём тестовый заказ
    test_order = create_test_order()

    # Отправляем сообщение в Kafka
    send_message(producer, topic, test_order)

    time.sleep(2)
    
    # Проверяем кеш
    test_cache_effect(test_order["order_uid"])

    # Можно отправить несколько заказов для массового тестирования
    for _ in range(3):
        order = create_test_order()
        send_message(producer, topic, order)

    producer.close()
    print("\n✅ Все тесты завершены")

# -------------------------------
# Запуск модуля
# -------------------------------
if __name__ == "__main__":
    main()

package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "todo-proj/api/proto" // Путь к сгенерированным файлам
)

func main() {
	// 1. Устанавливаем соединение с сервером (порт 50051)
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewNotifierClient(conn)

	// 2. Формируем запрос
	req := &pb.NotificationRequest{
		TaskTitle: "Новая задача из проекта",
		Message:   "Проверьте список дел на сегодня",
	}

	// 3. Отправляем запрос
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.SendNotification(ctx, req)
	if err != nil {
		log.Fatalf("ошибка при отправке: %v", err)
	}

	log.Printf("Ответ от сервера: %v", res.Success)
}
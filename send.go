package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/HelenMatveevaN/todo-proj/api/proto"
)

func main() {
	// 1. Устанавливаем соединение с сервером (порт 50051)
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewNotifierServiceClient(conn)

	// 2. Формируем запрос
	req := &pb.SendNotificationRequest{
		UserId:      "user-123",
		Message:     "Тестовое уведомление из TODO сервиса!",
		Destination: "test@example.com",
	}

	// 3. Отправляем запрос
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Send(ctx, req)
	if err != nil {
		log.Fatalf("ошибка при отправке: %v", err)
	}

	log.Printf("Ответ от сервера: %v (Success: %v)", res.NotificationId, res.Success)
}
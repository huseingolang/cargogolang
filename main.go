package main

import (
	conf "chat/Conf"
	controller "chat/Controller"
	model "chat/Model"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	mutex     sync.Mutex
	statusses = []string{
		"доставлен",
		"в ожидании",
		"не доставлен",
	}
)

func GetNextStatus(current string) string {

	for i, s := range statusses {

		if s == current && i < len(statusses)-1 {

			return statusses[i+1]
		}
	}
	fmt.Println(current)
	return current
}
func ChangeStatus() {
	for {
		fmt.Println("обновление....")
		time.Sleep(10 * time.Second)
		var orders []model.Order
		mutex.Lock()
		if err := conf.DB.Where("current_status != ? ", "Доставлен").Find(&orders).Error; err != nil {
			log.Println("заказ не найден", err.Error())

		}
		log.Printf("Найден заказ:%d\n", len(orders))

		mutex.Unlock()

		if len(orders) == 0 {
			log.Println("нет заказов для обновление")
		}
		for _, order := range orders {
			mutex.Lock()
			nextStatus := GetNextStatus(order.CurrentStatus)
			if nextStatus != " " {
				order.CurrentStatus = nextStatus
				if err := conf.DB.Model(&order).Updates(map[string]interface{}{
					"current_status": order.CurrentStatus,
				}); err == nil {
					log.Println("error updates")
				}
				fmt.Println("Updates", nextStatus)

				mutex.Unlock()
			}
		}

	}

}

func main() {
	conf.Init()

	conf.DB.AutoMigrate(model.User{}, model.Notification{}, model.Status{}, model.Order{}, model.Chake{})

	go ChangeStatus()

	handler := gin.Default()

	handler.POST("/register", controller.Register)
	handler.POST("/login", controller.Login)
	handler.POST("/order", controller.CreateOrder)
	handler.GET("/order", controller.GetOrder)
	handler.GET("/chake", controller.FindChake)
	handler.POST("/chake", controller.CreateChake)
	handler.POST("/status", controller.CreateStatus)
	handler.GET("/status", controller.GetStatus)
	handler.GET("/order/:id", controller.GetOrderOne)
	handler.GET("/profile/:id", controller.GetProfile)
	handler.GET("/natification/:UserID", controller.GetNotification)
	handler.PATCH("/order/:id", controller.UpdateOrder)
	handler.PATCH("/status/:id", controller.UpdateStatus)
	handler.DELETE("/order/:id", controller.DeleteOrder)

	handler.Run(":8001")

}

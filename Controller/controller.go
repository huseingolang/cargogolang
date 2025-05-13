package controller

import (
	conf "chat/Conf"
	"chat/Controller/Hook"
	model "chat/Model"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

var (
	mutex     sync.Mutex
	statusses = []string{
		"доставлен",
		"в ожидании",
		"не доставлен",
	}

	c       = context.Background()
	cx      = context.Background()
	chanche = "user:All"
	list    = "orders:All"
)

func Register(ctx *gin.Context) {

	var body struct {
		Email    string
		Password string
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error json",
		})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error hash",
		})
		return

	}

	user := model.User{Email: body.Email, Password: string(hash)}
	token := Hook.GenerateJWT(user.Email)
	chanc := map[string]interface{}{
		"Email":    user.Email,
		"Password": user.Password,
	}

	if err := conf.DB.Where("email = ? ", body.Email).First(&user).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error email ok",
		})
		return
	}
	userValue, err := conf.Rdb.HGetAll(c, chanche).Result()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error redis nil",
		})
		return
	} else {
		err := conf.Rdb.HSet(c, chanche, chanc).Err()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error redis create",
			})
			return

		}
		conf.Rdb.Expire(c, chanche, 5*time.Minute)

		value, err := conf.Rdb.HGet(c, chanche, "Email").Result()
		if err != nil {
			fmt.Println("error")
		}
		fmt.Println(value)
		ctx.JSON(http.StatusOK, gin.H{
			"rds": userValue,
		})
		fmt.Println(value)

	}

	conf.DB.Create(&user)

	ctx.JSON(http.StatusOK, gin.H{
		"message": user,
		"token":   token,
	})
}
func Login(ctx *gin.Context) {

	var body struct {
		Email    string
		Password string
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error json",
		})
		return
	}

	var user model.User

	if err := conf.DB.Where("email = ? ", body.Email).First(&user).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error valid email",
		})
		return

	}

	if user.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error noun email",
		})
		return

	}
	if !user.CheckPassword(body.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error invalid password",
		})
		return

	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   user,
		"messageOK": "StatusOK",
	})

}
func ValidateUSer(ctx *gin.Context) {
	var user model.User

	conf.DB.Find(&user)
	ctx.JSON(http.StatusOK, gin.H{
		"message": user.ID,
	})
}
func GetProfile(ctx *gin.Context) {
	id := ctx.Param("id")
	var user model.User

	if err := conf.DB.Where("id = ?", id).First(&user).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return

	}
	if err := conf.DB.Preload("Orders").First(&user).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid orders",
		})
		return

	}

	conf.DB.Find(&user)

	ctx.JSON(http.StatusOK, gin.H{
		"message":   user,
		"messageOK": "StatusOK",
	})

}
func GetNextStatus(current string) string {

	for i, s := range statusses {

		if s == current && i < len(statusses)-1 {

			return statusses[i+1]
		}
	}
	fmt.Println(current)
	return current
}
func CreateOrder(ctx *gin.Context) {

	var body struct {
		Name          string
		Description   string
		Status        []model.Status
		CurrentStatus string

		UserID uint
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error json",
		})
		return
	}

	order := model.Order{Name: body.Name, Description: body.Description, Status: body.Status, CurrentStatus: body.CurrentStatus, UserID: body.UserID}

	order.CurrentTime = time.Now().Add(10 * time.Second)
	order.CurrentStatus = statusses[0]
	mutex.Lock()
	if err := conf.DB.Where("name = ? ", body.Name).First(&order).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "такой заказ уже есть",
		})
		return

	}
	mutex.Unlock()

	mutex.Lock()

	rdsorder := map[string]interface{}{
		"Name":          order.Name,
		"Description":   order.Description,
		"CurrentStatus": order.CurrentStatus,
		"UserID":        order.UserID,
	}
	userValue, err := conf.Rdb.HGetAll(cx, list).Result()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error redis nil",
		})
		return
	} else {
		err := conf.Rdb.HSet(cx, list, rdsorder).Err()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error redis create",
			})
			return

		}
		conf.Rdb.SAdd(cx, list, rdsorder)
		conf.Rdb.Expire(cx, list, 5*time.Minute)

		value, err := conf.Rdb.HGet(cx, list, "Name").Result()
		if err != nil {
			fmt.Println("error")
		}
		fmt.Println(value)
		ctx.JSON(http.StatusOK, gin.H{
			"rds": userValue,
		})
		fmt.Println(value)

	}

	conf.DB.Create(&order)
	userID, _ := ctx.Get("UserId")

	mutex.Unlock()

	if err := conf.DB.Preload("Status").Find(&order).Error; err != nil {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "invalid Product reader",
		})
		return

	}
	statusUpdate := model.Status{
		OrderID:       order.ID,
		CurrentStatus: order.CurrentStatus,
	}
	chakeUpdate := model.Chake{
		OrderID:       order.ID,
		CurrentStatus: order.CurrentStatus,
	}
	conf.DB.Create(&chakeUpdate)
	conf.DB.Create(&statusUpdate)

	ctx.JSON(http.StatusOK, gin.H{
		"message":   order,
		"id":        userID,
		"messageOK": "StatusOK",
		"redis":     userValue,
	})

}
func CreateChake(ctx *gin.Context) {
	var chake model.Chake

	if err := ctx.ShouldBindJSON(&chake); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error json",
		})
		return
	}
	var order model.Order

	err := conf.DB.First(&order, chake.OrderID).Error

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error orderId nil ",
		})
		return
	}

	if order.CurrentStatus == "доставлен" {
		chake.CurrentStatus = order.CurrentStatus
		conf.DB.Create(&chake)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error status nil ",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   chake,
		"messageOK": "StatusOK",
	})

}
func FindChake(ctx *gin.Context) {
	var chake []model.Chake

	conf.DB.Find(&chake)
	ctx.JSON(http.StatusOK, gin.H{
		"message":   chake,
		"messageOK": "StatusOK",
	})
}
func GetOrder(ctx *gin.Context) {
	var order []model.Order

	orders, err := conf.Rdb.HGetAll(cx, list).Result()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message":   "Error redis found",
			"messageOK": "StatuNot",
		})
	}

	mutex.Lock()

	conf.DB.Find(&order)

	mutex.Unlock()

	if err := conf.DB.Preload("Status").Find(&order).Error; err != nil {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "invalid Product reader",
		})
		return

	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   order,
		"messageOK": "StatusOK",
		"redis":     orders,
	})
}
func GetOrderOne(ctx *gin.Context) {
	id := ctx.Param("id")

	var order model.Order
	if err := conf.DB.Where("id = ? ", id).First(&order).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "id не существует",
		})
		return

	}
	conf.DB.Find(&order)
	if err := conf.DB.Preload("Status").Find(&order).Error; err != nil {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "invalid Product reader",
		})
		return

	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   order,
		"messageOK": "StatusOK",
	})
}
func UpdateOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	var body struct {
		Name string
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error json",
		})
		return
	}

	var order model.Order
	if err := conf.DB.Where("id = ? ", id).First(&order).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "id не существует",
		})
		return

	}
	conf.DB.Model(&order).Updates(model.Order{Name: body.Name})

	ctx.JSON(http.StatusOK, gin.H{
		"message":   order,
		"messageOK": "StatusOK",
	})
}
func DeleteOrder(ctx *gin.Context) {
	var order model.Order
	id := ctx.Param("id")

	if err := conf.DB.Where("id = ? ", id).First(&order).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "id не существует",
		})
		return

	}

	conf.DB.Delete(&order)

	ctx.JSON(http.StatusOK, gin.H{
		"message":   order,
		"messageOK": "StatusOK",
	})

}
func CreateStatus(ctx *gin.Context) {
	var Status model.Status

	if err := ctx.ShouldBindJSON(&Status); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error json",
		})
		return
	}
	var order model.Order
	if err := conf.DB.First(&order, order.ID).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error orderID",
		})
		return
	}
	Status.CurrentStatus = order.CurrentStatus
	conf.DB.Save(&order)
	message := fmt.Sprintf("Статус вашего заказа %d изменилсья на :%s", order.ID, order.CurrentStatus)
	notification := model.Notification{
		UserID:  order.ID,
		Message: message,
		IsSent:  false,
	}
	conf.DB.Create(&Status)
	conf.DB.Create(&notification)

	ctx.JSON(http.StatusOK, gin.H{
		"message":   Status,
		"messageOK": "StatusOK",
	})

}

func GetStatus(ctx *gin.Context) {
	var Status []model.Status

	conf.DB.Find(&Status)

	ctx.JSON(http.StatusOK, gin.H{
		"message":   Status,
		"messageOK": "StatusOK",
	})

}
func GetNotification(ctx *gin.Context) {
	id := ctx.Param("UserID")
	var notifications []model.Notification

	if err := conf.DB.Where("UserID = ? ", id).First(&notifications).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "id не существует",
		})
		return

	}
	ctx.JSON(http.StatusOK, gin.H{
		"message":   notifications,
		"messageOK": "StatusOK",
	})

}
func UpdateStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var body struct {
		CurrentStatus string
	}
	var status model.Status

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error json",
		})
		return
	}
	if err := conf.DB.Where("id = ? ", id).First(&status).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "error valid id",
		})
		return

	}
	conf.DB.Model(&status).Updates(model.Status{CurrentStatus: body.CurrentStatus})
	ctx.JSON(http.StatusOK, gin.H{
		"message":   status,
		"messageOK": "StatusOK",
	})

}

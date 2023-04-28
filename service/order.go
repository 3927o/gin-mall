package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"

	util "mall/pkg/utils"
	"mall/pkg/utils/ctl"
	"mall/repository/cache"
	"mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/types"
)

const OrderTimeKey = "OrderTime"

var OrderSrvIns *OrderSrv
var OrderSrvOnce sync.Once

type OrderSrv struct {
}

func GetOrderSrv() *OrderSrv {
	OrderSrvOnce.Do(func() {
		OrderSrvIns = &OrderSrv{}
	})
	return OrderSrvIns
}

func (s *OrderSrv) OrderCreate(ctx context.Context, id uint, req *types.OrderServiceReq) (resp interface{}, err error) {
	order := &model.Order{
		UserID:    id,
		ProductID: req.ProductID,
		BossID:    req.BossID,
		Num:       int(req.Num),
		Money:     float64(req.Money),
		Type:      1,
	}
	addressDao := dao.NewAddressDao(ctx)
	address, err := addressDao.GetAddressByAid(req.AddressID)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}

	order.AddressID = address.ID
	number := fmt.Sprintf("%09v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000000))
	productNum := strconv.Itoa(int(req.ProductID))
	userNum := strconv.Itoa(int(id))
	number = number + productNum + userNum
	orderNum, _ := strconv.ParseUint(number, 10, 64)
	order.OrderNum = orderNum

	orderDao := dao.NewOrderDao(ctx)
	err = orderDao.CreateOrder(order)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}

	// 订单号存入Redis中，设置过期时间
	data := redis.Z{
		Score:  float64(time.Now().Unix()) + 15*time.Minute.Seconds(),
		Member: orderNum,
	}
	cache.RedisClient.ZAdd(OrderTimeKey, data)
	return ctl.RespSuccess(), nil
}

func (s *OrderSrv) OrderList(ctx context.Context, uId uint, req *types.OrderServiceReq) (resp interface{}, err error) {
	orders, total, err := dao.NewOrderDao(ctx).ListOrderByCondition(uId, req)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}

	return ctl.RespList(orders, total), nil
}

func (s *OrderSrv) OrderShow(ctx context.Context, uId uint, req *types.OrderServiceReq) (resp interface{}, err error) {
	orderDao := dao.NewOrderDao(ctx)
	order, err := orderDao.ShowOrderById(uId, req.OrderId)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}

	return ctl.RespSuccessWithData(order), nil
}

func (s *OrderSrv) OrderDelete(ctx context.Context, uId uint, req *types.OrderServiceReq) (resp interface{}, err error) {
	err = dao.NewOrderDao(ctx).DeleteOrderById(req.OrderId, uId)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}

	return
}

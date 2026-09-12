package timer

import (
	"mall/adaptor"
	"mall/adaptor/redis"
	"mall/service/goods"
	"mall/service/order"

	"github.com/go-co-op/gocron"

	"mall/config"
	"time"
)

type Service struct {
	conf     *config.Config
	schedule *gocron.Scheduler
	order    *order.Service
	goods    *goods.Service
	rdsOrder redis.IOrder
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:     adaptor.GetConfig(),
		schedule: gocron.NewScheduler(time.Local),
		order:    order.NewService(adaptor),
		goods:    goods.NewService(adaptor),
		rdsOrder: redis.NewOrder(adaptor),
	}
}

func (s *Service) Start() {
	// 待支付订单的超时取消动作，每分钟检查一次
	_, err := s.schedule.Every(1).Minute().Do(s.OrderTimeOutCancel)
	if err != nil {
		panic(err)
	}

	// 优先等待支付回调，订单支付结果主动轮询，每5s中我就查询一次，定时和回调谁先到，谁处理订单支付结果
	_, err = s.schedule.Every(5).Seconds().Do(s.OrderPayResultQuery)
	if err != nil {
		panic(err)
	}

	// 订单退款查询
	_, err = s.schedule.Every(5).Seconds().Do(s.OrderRefundQuery)
	if err != nil {
		panic(err)
	}

	// 异步启动，不阻塞
	s.schedule.StartAsync()
}

func (s *Service) Stop() {
	s.schedule.Stop()
}

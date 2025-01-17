package listen

import (
	"context"
	"looklook/app/order/cmd/mq/internal/config"
	kqMq "looklook/app/order/cmd/mq/internal/mqs/kq"
	"looklook/app/order/cmd/mq/internal/svc"

	"github.com/tal-tech/go-queue/kq"
	"github.com/tal-tech/go-zero/core/service"
)

//kq
//消息队列
func KqMqs(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {

	return []service.Service{
		//监听消费流水状态变更
		kq.MustNewQueue(c.PaymentUpdateStatusConf, kqMq.NewPaymentUpdateStatusMq(ctx, svcContext)),
		//.....
	}

}

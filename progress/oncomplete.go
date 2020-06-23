package progress

import (
	"github.com/vbauerster/mpb/decor"
)

type onCompleteFailed struct {
	decor.WC
	failed      *bool
	failedMsg   string
	completeMsg *string
}

func OnCompleteFailed(failed *bool, failedMsg string, wcc ...decor.WC) decor.Decorator {
	var wc decor.WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	return &onCompleteFailed{WC: wc, failed: failed, failedMsg: failedMsg}
}

func (o *onCompleteFailed) Decor(stats *decor.Statistics) string {
	if stats.Completed {
		if !*o.failed && o.completeMsg != nil {
			return o.FormatMsg(*o.completeMsg)
		} else if *o.failed {
			return o.FormatMsg(o.failedMsg)
		}
	}
	return ""
}

func (o *onCompleteFailed) OnCompleteMessage(msg string) {
	o.completeMsg = &msg
}

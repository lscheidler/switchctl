package progress

import (
	"fmt"
	"time"

	"github.com/vbauerster/mpb/decor"
)

type onStepName struct {
	decor.WC
	current   int
	stepNames []string
}

func OnStepName(steps []string, wcc ...decor.WC) decor.Decorator {
	var wc decor.WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	return &onStepName{WC: wc, stepNames: steps, current: 0}
}

func (o *onStepName) Decor(stats *decor.Statistics) string {
	if o.current == 0 {
		return o.FormatMsg("")
	} else if o.current > len(o.stepNames) {
		return o.FormatMsg(fmt.Sprintf("%v", o.stepNames))
	} else {
		return o.FormatMsg(fmt.Sprintf("%v", o.stepNames[0:o.current]))
	}
}

func (o *onStepName) NextAmount(amount int, times ...time.Duration) {
	o.current = o.current + amount
}

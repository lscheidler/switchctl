package progress

import (
	"fmt"
	"time"

	"github.com/vbauerster/mpb/decor"
)

type onStepFunction struct {
	decor.WC
	current       int
	stepFunctions []func() string
}

func OnStepFunction(steps []func() string, wcc ...decor.WC) decor.Decorator {
	var wc decor.WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	return &onStepFunction{WC: wc, stepFunctions: steps, current: 0}
}

func (o *onStepFunction) Decor(stats *decor.Statistics) string {
	str := ""
	if o.current > 0 && o.current <= len(o.stepFunctions) {
		names := []string{}
		length := o.current
		if o.current > len(o.stepFunctions) {
			length = len(o.stepFunctions)
		}
		for _, function := range o.stepFunctions[0:length] {
			names = append(names, function())
		}
		str = fmt.Sprintf("%v", names)
	}
	return o.FormatMsg(str)
}

func (o *onStepFunction) NextAmount(amount int, times ...time.Duration) {
	o.current = o.current + amount
}

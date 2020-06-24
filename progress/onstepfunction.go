/*
  Copyright 2020 Lars Eric Scheidler

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/
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

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

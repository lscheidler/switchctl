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

// +build amp

package gopentracing

import (
	"github.com/opentracing/opentracing-go"
	"go.elastic.co/apm/module/apmot"
)

func init() {
	opentracing.SetGlobalTracer(apmot.New())
}

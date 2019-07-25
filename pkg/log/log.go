// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"io"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// ConfigFunc sets configuration Options for `zapcore.EncoderConfig`.
type ConfigFunc func(config *zapcore.EncoderConfig)

// ZapLoggerTo returns a new Logger implementation using Zap which logs
// to the given destination, instead of stderr.  It otherise behaves like
// ZapLogger.
// This function is mainly copied from controller-runtime `sigs.k8s.io/controller-runtime/pkg/runtime/log`
// and extended to pass options for log encoding.
// TODO: Switch back to `sigs.k8s.io/controller-runtime/pkg/runtime/log` once proper options are exposed:
// https://github.com/kubernetes-sigs/controller-runtime/issues/442
func ZapLoggerTo(destWriter io.Writer, development bool, configFuncs ...ConfigFunc) logr.Logger {
	// this basically mimics New<type>Config, but with a custom sink
	sink := zapcore.AddSync(destWriter)

	var enc zapcore.Encoder
	var lvl zap.AtomicLevel
	var opts []zap.Option
	if development {
		encCfg := zap.NewDevelopmentEncoderConfig()
		for _, f := range configFuncs {
			f(&encCfg)
		}
		enc = zapcore.NewConsoleEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		encCfg := zap.NewProductionEncoderConfig()
		for _, f := range configFuncs {
			f(&encCfg)
		}
		enc = zapcore.NewJSONEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		opts = append(opts, zap.AddStacktrace(zap.WarnLevel),
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				return zapcore.NewSampler(core, time.Second, 100, 100)
			}))
	}
	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	log := zap.New(zapcore.NewCore(&log.KubeAwareEncoder{Encoder: enc, Verbose: development}, sink, lvl))
	log = log.WithOptions(opts...)
	return zapr.NewLogger(log)
}

// ZapLogger is a Logger implementation.
// If development is true, a Zap development config will be used
// (stacktraces on warnings, no sampling), otherwise a Zap production
// config will be used (stacktraces on errors, sampling).
// Additionally, the time encoding is adjusted to `zapcore.ISO8601TimeEncoder`.
// This function is mainly copied from controller-runtime `sigs.k8s.io/controller-runtime/pkg/runtime/log`
// and extended to pass options for log encoding.
// TODO: Switch back to `sigs.k8s.io/controller-runtime/pkg/runtime/log` once proper options are exposed:
// https://github.com/kubernetes-sigs/controller-runtime/issues/442
func ZapLogger(development bool) logr.Logger {
	timestampConfig := func(config *zapcore.EncoderConfig) {
		config.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	return ZapLoggerTo(os.Stderr, development, timestampConfig)
}

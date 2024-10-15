/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package goscr

import (
	"fmt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/goscr/fwlib"
	"git.golaxy.org/framework/plugins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"sync/atomic"
	"time"
)

// IGoScr golang脚本支持
type IGoScr interface {
}

func newGoScr(setting ...option.Setting[GoScrOptions]) IGoScr {
	return &_GoScr{
		options: option.Make(With.Default(), setting...),
	}
}

type _GoScr struct {
	svcCtx  service.Context
	options GoScrOptions
	intp    *interp.Interpreter
}

// InitSP 初始化服务插件
func (s *_GoScr) InitSP(svcCtx service.Context) {
	s.svcCtx = svcCtx

	intp, err := s.load()
	if err != nil {
		log.Panicln(s.svcCtx, err)
	}

	s.intp = intp

	if s.options.AutoHotFix {
		s.hotFix()
	}
}

// ShutSP 关闭服务插件
func (s *_GoScr) ShutSP(svcCtx service.Context) {
	log.Infof(svcCtx, "shut plugin %q", self.Name)
}

func (s *_GoScr) load() (*interp.Interpreter, error) {
	intp := interp.New(interp.Options{})
	intp.Use(stdlib.Symbols)
	intp.Use(fwlib.Symbols)

	for _, path := range s.options.PathList {
		if _, err := intp.EvalPath(path); err != nil {
			return nil, fmt.Errorf("load script path %q failed, %s", path, err)
		}
		if _, err := intp.Eval(fmt.Sprintf(`import "%s"`, path)); err != nil {
			return nil, fmt.Errorf("import script path %q failed, %s", path, err)
		}
	}

	return intp, nil
}

func (s *_GoScr) hotFix() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panicf(s.svcCtx, "hotfix script watch %+v failed, %s", s.options.PathList, err)
	}

	for _, path := range s.options.PathList {
		if err = watcher.AddWith(path); err != nil {
			log.Panicf(s.svcCtx, "hotfix script watch %q failed, %s", path, err)
		}
	}

	go func() {
		var reloading atomic.Bool

		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					return
				}

				if !reloading.CompareAndSwap(false, true) {
					continue
				}

				log.Infof(s.svcCtx, "hotfix script detecting %q changes, preparing to reload in 10s", e)

				go func() {
					defer reloading.Store(false)

					time.Sleep(10 * time.Second)

					intp, err := s.load()
					if err != nil {
						log.Errorf(s.svcCtx, "hotfix script reload %q failed, %s", e, err)
						return
					}

					s.intp = intp
				}()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Errorf(s.svcCtx, "hotfix script watch failed, %s", err)
			}
		}
	}()

	log.Infof(s.svcCtx, "hotfix script watch %+v ok", s.options.PathList)
}

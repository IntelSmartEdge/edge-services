// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2019 Intel Corporation

package service

import (
	"context"
	"errors"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = func() (_ struct{}) {
	os.Args = append(os.Args, "-config=../../configs/appliance.json")
	return
}()

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main suite")
}

type fakeAgent struct {
	ContextCancelled bool
	EndedWork        bool
	CfgPath          string
}

func (a *fakeAgent) run(parentCtx context.Context, cfg string) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	select {
	case <-time.After(10 * time.Millisecond):
		a.EndedWork = true
	case <-ctx.Done():
		a.ContextCancelled = true
	}
	a.CfgPath = cfg
	return nil
}

func failingRun(parentCtx context.Context, cfg string) error {
	return errors.New("Fail")
}

func successfulRun(parentCtx context.Context, cfg string) error {
	return nil
}

func setConfigPath(f StartFunction) {
	Cfg.Services = make(map[string]string)
	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	// An example of funcName:
	// applications.services.smart-edge-open.edge-services/pkg/certsigner.(*CertificateSigner).Run-fm
	// we need to find the position of the first dot after last slash
	lastSlashPos := strings.LastIndex(funcName, "/")
	// If there's no slash in the function name then reset the position to 0
	if lastSlashPos == -1 {
		lastSlashPos = 0
	}
	firstDotAfterLastSlashPos := strings.Index(funcName[lastSlashPos:], ".") + lastSlashPos
	srvName := funcName[:firstDotAfterLastSlashPos]

	Cfg.Services[srvName] = "config.json"
}

var _ = Describe("runServices", func() {
	var (
		controlAgent fakeAgent
		controlRun   StartFunction = controlAgent.run
	)

	BeforeEach(func() {
		Cfg.LogLevel = "debug"

		controlAgent = fakeAgent{}
		setConfigPath(controlRun)
	})

	Describe("Starts an Agent that will fail", func() {
		It("Will return failure and context cancellation will be issued",
			func() {
				Expect(RunServices([]StartFunction{failingRun,
					successfulRun, controlRun})).Should(BeFalse())
				Expect(controlAgent.ContextCancelled).Should(BeTrue())
				Expect(controlAgent.EndedWork).Should(BeFalse())
				Expect(controlAgent.CfgPath).Should(Equal("config.json"))
			})
	})

	Describe("Starts an Agent that will succeed", func() {
		It("Will return success and other agents will finish work normally",
			func() {
				Expect(RunServices([]StartFunction{successfulRun,
					controlRun})).Should(BeTrue())
				Expect(controlAgent.EndedWork).Should(BeTrue())
				Expect(controlAgent.ContextCancelled).Should(BeFalse())
				Expect(controlAgent.CfgPath).Should(Equal("config.json"))
			})
	})
})

var _ = Describe("init", func() {
	Describe("Init config with not existing cfg file", func() {
		It("Will return failure",
			func() {
				err := InitConfig("testdata/notExistFile.json")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to load config"))
			})
	})

	Describe("Init config with incorrect parse level", func() {
		It("Will return failure",
			func() {
				err := InitConfig("testdata/parseLevel.json")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to parse log level"))
			})
	})

	Describe("Init config with incorrect syslog address", func() {
		It("Will return failure",
			func() {
				err := InitConfig("testdata/useSyslog.json")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to connect to syslog"))
			})
	})
})
